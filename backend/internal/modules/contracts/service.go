package contracts

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"gerador-contratos/backend/internal/modules/auth"
	"gerador-contratos/backend/internal/modules/common"
	"gerador-contratos/backend/internal/modules/rules"
)

type Service struct {
	repo   Repository
	rules  *rules.Service
}

func NewService(repo Repository, rulesService *rules.Service) *Service {
	return &Service{repo: repo, rules: rulesService}
}

func (s *Service) CreateContract(ctx context.Context, actor auth.AuthClaims, in CreateContractInput) (*ContractDetails, error) {
	in.Numero = strings.TrimSpace(in.Numero)
	in.Tipo = strings.TrimSpace(in.Tipo)
	if in.Numero == "" {
		return nil, common.NewBadRequest("invalid_contract_number", "numero do contrato e obrigatorio")
	}
	if in.Tipo == "" {
		return nil, common.NewBadRequest("invalid_contract_type", "tipo do contrato e obrigatorio")
	}
	if in.Status == "" {
		in.Status = "rascunho"
	}
	if in.Data == nil {
		in.Data = map[string]any{}
	}

	contract := Contract{
		ID:        uuid.NewString(),
		TenantID:  actor.TenantID,
		Numero:    in.Numero,
		Tipo:      in.Tipo,
		Status:    in.Status,
		CreatedBy: actor.UserID,
	}
	if err := s.repo.CreateContract(ctx, contract); err != nil {
		return nil, err
	}

	version := ContractVersion{
		ID:            uuid.NewString(),
		ContractID:    contract.ID,
		VersionNumber: 1,
		Data:          in.Data,
		CreatedBy:     actor.UserID,
	}
	if err := s.repo.CreateVersion(ctx, version); err != nil {
		return nil, err
	}

	return s.GetContractDetails(ctx, actor, contract.ID)
}

func (s *Service) AddVersion(ctx context.Context, actor auth.AuthClaims, contractID string, in AddVersionInput) (*ContractVersion, error) {
	if strings.TrimSpace(contractID) == "" {
		return nil, common.NewBadRequest("invalid_contract_id", "contract id invalido")
	}
	if in.Data == nil {
		return nil, common.NewBadRequest("invalid_payload", "dados da versao sao obrigatorios")
	}

	if _, err := s.repo.GetContract(ctx, actor.TenantID, contractID); err != nil {
		return nil, err
	}

	latest, err := s.repo.GetLatestVersion(ctx, contractID)
	next := 1
	if err == nil {
		next = latest.VersionNumber + 1
	} else {
		var appErr *common.AppError
		if errors.As(err, &appErr) && appErr.Code != "contract_version_not_found" {
			return nil, err
		}
	}

	version := ContractVersion{
		ID:            uuid.NewString(),
		ContractID:    contractID,
		VersionNumber: next,
		Data:          in.Data,
		CreatedBy:     actor.UserID,
	}
	if err := s.repo.CreateVersion(ctx, version); err != nil {
		return nil, err
	}
	return s.repo.GetVersion(ctx, contractID, next)
}

func (s *Service) ListContracts(ctx context.Context, actor auth.AuthClaims, filter ListContractsFilter) ([]Contract, error) {
	return s.repo.ListContracts(ctx, actor.TenantID, filter)
}

func (s *Service) GetContractDetails(ctx context.Context, actor auth.AuthClaims, contractID string) (*ContractDetails, error) {
	contract, err := s.repo.GetContract(ctx, actor.TenantID, contractID)
	if err != nil {
		return nil, err
	}

	versions, err := s.repo.ListVersions(ctx, contract.ID)
	if err != nil {
		return nil, err
	}

	var latest *ContractVersion
	if len(versions) > 0 {
		latest = &versions[0]
	}

	return &ContractDetails{
		Contract:      *contract,
		LatestVersion: latest,
		Versions:      versions,
	}, nil
}

func (s *Service) PreviewLatest(ctx context.Context, actor auth.AuthClaims, contractID string) (*ContractPreview, error) {
	details, err := s.GetContractDetails(ctx, actor, contractID)
	if err != nil {
		return nil, err
	}
	if details.LatestVersion == nil {
		return nil, common.NewNotFound("contract_version_not_found", "contrato sem versoes")
	}

	preview := s.rules.BuildPreview(details.Contract.Numero, details.Contract.Tipo, details.LatestVersion.Data)
	return &ContractPreview{
		Title:    preview.Title,
		Sections: preview.Sections,
		FullText: preview.FullText,
	}, nil
}

func (s *Service) PreviewFromData(numero, tipo string, data map[string]any) (*ContractPreview, error) {
	if strings.TrimSpace(tipo) == "" {
		return nil, common.NewBadRequest("invalid_contract_type", "tipo do contrato e obrigatorio")
	}
	if data == nil {
		data = map[string]any{}
	}

	preview := s.rules.BuildPreview(numero, tipo, data)
	return &ContractPreview{
		Title:    preview.Title,
		Sections: preview.Sections,
		FullText: preview.FullText,
	}, nil
}

func (s *Service) GetVersion(ctx context.Context, actor auth.AuthClaims, contractID string, versionNumber int) (*ContractVersion, error) {
	if _, err := s.repo.GetContract(ctx, actor.TenantID, contractID); err != nil {
		return nil, err
	}
	if versionNumber < 1 {
		return nil, common.NewBadRequest("invalid_version", "numero de versao invalido")
	}
	version, err := s.repo.GetVersion(ctx, contractID, versionNumber)
	if err != nil {
		return nil, err
	}
	return version, nil
}

func statusFromAny(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
