package auth

import (
	"context"
	"fmt"
	"net/smtp"
	"net/url"
	"strings"
	"time"
)

// SMTPPasswordResetNotifierConfig concentra a configuracao de envio SMTP.
type SMTPPasswordResetNotifierConfig struct {
	Host             string
	Port             int
	Username         string
	Password         string
	From             string
	PasswordResetURL string
}

// SMTPPasswordResetNotifier envia notificacoes de recuperacao por SMTP.
type SMTPPasswordResetNotifier struct {
	host             string
	port             int
	username         string
	password         string
	from             string
	passwordResetURL string
}

// NewSMTPPasswordResetNotifier valida e constroi o adapter SMTP.
func NewSMTPPasswordResetNotifier(cfg SMTPPasswordResetNotifierConfig) (*SMTPPasswordResetNotifier, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		return nil, fmt.Errorf("smtp host is required")
	}

	from := strings.TrimSpace(cfg.From)
	if from == "" {
		return nil, fmt.Errorf("smtp from is required")
	}

	port := cfg.Port
	if port <= 0 {
		port = 587
	}

	return &SMTPPasswordResetNotifier{
		host:             host,
		port:             port,
		username:         strings.TrimSpace(cfg.Username),
		password:         cfg.Password,
		from:             from,
		passwordResetURL: strings.TrimSpace(cfg.PasswordResetURL),
	}, nil
}

// SendPasswordReset publica o token de recuperacao para o e-mail do usuario.
func (n *SMTPPasswordResetNotifier) SendPasswordReset(ctx context.Context, notification PasswordResetNotification) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", n.host, n.port)
	msg := n.buildMessage(notification)
	recipients := []string{notification.ToEmail}

	var smtpAuth smtp.Auth
	if n.username != "" {
		// PlainAuth usa o host sem porta para autenticacao SMTP.
		smtpAuth = smtp.PlainAuth("", n.username, n.password, n.host)
	}

	if err := smtp.SendMail(addr, smtpAuth, n.from, recipients, []byte(msg)); err != nil {
		return fmt.Errorf("send smtp mail: %w", err)
	}

	return nil
}

// buildMessage monta um e-mail texto simples com token e link de recuperacao.
func (n *SMTPPasswordResetNotifier) buildMessage(notification PasswordResetNotification) string {
	resetLink := n.buildResetLink(notification.ResetURL, notification.Token)
	bodyLines := []string{
		"Voce solicitou a recuperacao de senha.",
		"",
		fmt.Sprintf("Token: %s", notification.Token),
		fmt.Sprintf("Expira em: %s", notification.ExpiresAt.Format(time.RFC3339)),
	}
	if resetLink != "" {
		bodyLines = append(bodyLines, fmt.Sprintf("Link para redefinir senha: %s", resetLink))
	}
	bodyLines = append(bodyLines, "", "Se voce nao solicitou, ignore este e-mail.")

	headers := []string{
		fmt.Sprintf("From: %s", n.from),
		fmt.Sprintf("To: %s", notification.ToEmail),
		"Subject: Recuperacao de senha",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
	}

	return strings.Join(headers, "\r\n") + "\r\n\r\n" + strings.Join(bodyLines, "\n")
}

// buildResetLink gera um link com token quando a URL base foi configurada.
func (n *SMTPPasswordResetNotifier) buildResetLink(overrideURL string, token string) string {
	baseURL := strings.TrimSpace(overrideURL)
	if baseURL == "" {
		baseURL = n.passwordResetURL
	}
	if baseURL == "" {
		return ""
	}

	sep := "?"
	if strings.Contains(baseURL, "?") {
		sep = "&"
	}

	return baseURL + sep + "token=" + url.QueryEscape(token)
}
