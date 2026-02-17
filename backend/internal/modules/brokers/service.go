package brokers

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

func (s *Service) List(ctx context.Context, actor auth.AuthClaims) ([]Broker, error) {
	return s.repo.List(ctx, actor.TenantID)
}

func (s *Service) Create(ctx context.Context, actor auth.AuthClaims, in CreateBrokerInput) (*Broker, error) {
	in.Nome = strings.TrimSpace(in.Nome)
	if in.Nome == "" {
		return nil, common.NewBadRequest("invalid_name", "nome do corretor e obrigatorio")
	}

	broker := Broker{
		ID:       uuid.NewString(),
		TenantID: actor.TenantID,
		Nome:     in.Nome,
		CPF:      strings.TrimSpace(in.CPF),
		CRECI:    strings.TrimSpace(in.CRECI),
		Banco:    strings.TrimSpace(in.Banco),
		Agencia:  strings.TrimSpace(in.Agencia),
		Conta:    strings.TrimSpace(in.Conta),
		Pix:      strings.TrimSpace(in.Pix),
	}
	if err := s.repo.Create(ctx, broker); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, actor.TenantID, broker.ID)
}

func (s *Service) Update(ctx context.Context, actor auth.AuthClaims, brokerID string, in UpdateBrokerInput) (*Broker, error) {
	in.Nome = strings.TrimSpace(in.Nome)
	if in.Nome == "" {
		return nil, common.NewBadRequest("invalid_name", "nome do corretor e obrigatorio")
	}

	broker := Broker{
		ID:       strings.TrimSpace(brokerID),
		TenantID: actor.TenantID,
		Nome:     in.Nome,
		CPF:      strings.TrimSpace(in.CPF),
		CRECI:    strings.TrimSpace(in.CRECI),
		Banco:    strings.TrimSpace(in.Banco),
		Agencia:  strings.TrimSpace(in.Agencia),
		Conta:    strings.TrimSpace(in.Conta),
		Pix:      strings.TrimSpace(in.Pix),
	}

	if err := s.repo.Update(ctx, broker); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, actor.TenantID, broker.ID)
}

func (s *Service) Delete(ctx context.Context, actor auth.AuthClaims, brokerID string) error {
	return s.repo.Delete(ctx, actor.TenantID, strings.TrimSpace(brokerID))
}
