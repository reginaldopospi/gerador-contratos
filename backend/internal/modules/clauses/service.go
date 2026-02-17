package clauses

import (
	"context"
	"strings"

	"github.com/google/uuid"

	"gerador-contratos/backend/internal/modules/auth"
	"gerador-contratos/backend/internal/modules/common"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, actor auth.AuthClaims) ([]ClauseTemplate, error) {
	return s.repo.List(ctx, actor.TenantID)
}

func (s *Service) Upsert(ctx context.Context, actor auth.AuthClaims, in UpsertClauseInput) (*ClauseTemplate, error) {
	if actor.Role != auth.RoleAdmin && actor.Role != auth.RoleGestor {
		return nil, common.NewForbidden("insufficient_permissions", "perfil sem permissao para editar clausulas")
	}

	in.ClauseKey = strings.TrimSpace(in.ClauseKey)
	in.Title = strings.TrimSpace(in.Title)
	in.Content = strings.TrimSpace(in.Content)

	if in.ClauseKey == "" {
		return nil, common.NewBadRequest("invalid_clause_key", "chave da clausula e obrigatoria")
	}
	if in.Title == "" {
		return nil, common.NewBadRequest("invalid_title", "titulo da clausula e obrigatorio")
	}
	if in.Content == "" {
		return nil, common.NewBadRequest("invalid_content", "conteudo da clausula e obrigatorio")
	}

	clause := ClauseTemplate{
		ID:        uuid.NewString(),
		TenantID:  actor.TenantID,
		ClauseKey: in.ClauseKey,
		Title:     in.Title,
		Content:   in.Content,
		IsActive:  in.IsActive,
	}
	if err := s.repo.Upsert(ctx, clause); err != nil {
		return nil, err
	}
	clauses, err := s.repo.List(ctx, actor.TenantID)
	if err != nil {
		return nil, err
	}
	for _, c := range clauses {
		if c.ClauseKey == in.ClauseKey && c.TenantID == actor.TenantID {
			return &c, nil
		}
	}
	return nil, common.NewInternal("clause_upsert_failed", "nao foi possivel recuperar clausula apos gravacao")
}

func (s *Service) Delete(ctx context.Context, actor auth.AuthClaims, clauseID string) error {
	if actor.Role != auth.RoleAdmin && actor.Role != auth.RoleGestor {
		return common.NewForbidden("insufficient_permissions", "perfil sem permissao para remover clausulas")
	}
	return s.repo.Delete(ctx, actor.TenantID, strings.TrimSpace(clauseID))
}
