package auth

import (
	"context"
	"time"
)

// PasswordResetNotification carrega os dados necessarios para notificar o usuario.
type PasswordResetNotification struct {
	ToEmail   string
	Token     string
	ExpiresAt time.Time
	ResetURL  string
}

// PasswordResetNotifier define a porta de notificacao de recuperacao de senha.
type PasswordResetNotifier interface {
	SendPasswordReset(ctx context.Context, notification PasswordResetNotification) error
}

// NoopPasswordResetNotifier evita efeitos colaterais quando nenhum provedor foi configurado.
type NoopPasswordResetNotifier struct{}

// SendPasswordReset implementa um envio nulo para ambientes sem SMTP.
func (NoopPasswordResetNotifier) SendPasswordReset(ctx context.Context, notification PasswordResetNotification) error {
	return nil
}
