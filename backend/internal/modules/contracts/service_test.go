package contracts

import (
	"context"
	"fmt"
	"testing"
	"time"

	"gerador-contratos/backend/internal/modules/auth"
	"gerador-contratos/backend/internal/modules/common"
	"gerador-contratos/backend/internal/modules/rules"
)

type fakeRepo struct {
	contracts map[string]Contract
	versions  map[string][]ContractVersion
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		contracts: map[string]Contract{},
		versions:  map[string][]ContractVersion{},
	}
}

func (f *fakeRepo) CreateContract(ctx context.Context, contract Contract) error {
	for _, c := range f.contracts {
		if c.TenantID == contract.TenantID && c.Numero == contract.Numero {
			return common.NewConflict("contract_exists", "exists")
		}
	}
	contract.CreatedAt = time.Now().UTC()
	contract.UpdatedAt = time.Now().UTC()
	f.contracts[contract.ID] = contract
	return nil
}

func (f *fakeRepo) GetContract(ctx context.Context, tenantID, contractID string) (*Contract, error) {
	c, ok := f.contracts[contractID]
	if !ok || c.TenantID != tenantID {
		return nil, common.NewNotFound("contract_not_found", "not found")
	}
	return &c, nil
}

func (f *fakeRepo) ListContracts(ctx context.Context, tenantID string, filter ListContractsFilter) ([]Contract, error) {
	items := make([]Contract, 0)
	for _, c := range f.contracts {
		if c.TenantID == tenantID {
			items = append(items, c)
		}
	}
	return items, nil
}

func (f *fakeRepo) CreateVersion(ctx context.Context, version ContractVersion) error {
	version.CreatedAt = time.Now().UTC()
	f.versions[version.ContractID] = append(f.versions[version.ContractID], version)
	return nil
}

func (f *fakeRepo) GetLatestVersion(ctx context.Context, contractID string) (*ContractVersion, error) {
	list := f.versions[contractID]
	if len(list) == 0 {
		return nil, common.NewNotFound("contract_version_not_found", "not found")
	}
	latest := list[len(list)-1]
	return &latest, nil
}

func (f *fakeRepo) GetVersion(ctx context.Context, contractID string, versionNumber int) (*ContractVersion, error) {
	for _, v := range f.versions[contractID] {
		if v.VersionNumber == versionNumber {
			copy := v
			return &copy, nil
		}
	}
	return nil, common.NewNotFound("contract_version_not_found", "not found")
}

func (f *fakeRepo) ListVersions(ctx context.Context, contractID string) ([]ContractVersion, error) {
	list := f.versions[contractID]
	if len(list) == 0 {
		return []ContractVersion{}, nil
	}
	out := make([]ContractVersion, len(list))
	for i := range list {
		out[len(list)-1-i] = list[i]
	}
	return out, nil
}

func TestCreateContractAndVersioning(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, rules.NewService())
	claims := auth.AuthClaims{UserID: "user-1", TenantID: "tenant-1", Role: auth.RoleAdmin}

	details, err := svc.CreateContract(context.Background(), claims, CreateContractInput{
		Numero: "1981",
		Tipo:   "Compromisso de Venda e Compra de Imovel",
		Data: map[string]any{
			"preco_total": "R$ 100.000,00",
		},
	})
	if err != nil {
		t.Fatalf("create contract failed: %v", err)
	}
	if details.Contract.Numero != "1981" {
		t.Fatalf("unexpected contract number: %s", details.Contract.Numero)
	}
	if details.LatestVersion == nil || details.LatestVersion.VersionNumber != 1 {
		t.Fatalf("expected version 1")
	}

	version, err := svc.AddVersion(context.Background(), claims, details.Contract.ID, AddVersionInput{
		Data: map[string]any{"preco_total": "R$ 110.000,00"},
	})
	if err != nil {
		t.Fatalf("add version failed: %v", err)
	}
	if version.VersionNumber != 2 {
		t.Fatalf("expected version 2, got %d", version.VersionNumber)
	}

	preview, err := svc.PreviewLatest(context.Background(), claims, details.Contract.ID)
	if err != nil {
		t.Fatalf("preview latest failed: %v", err)
	}
	if preview.Title == "" || len(preview.Sections) == 0 {
		t.Fatalf("invalid preview: %#v", preview)
	}
}

func TestCreateContractConflict(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo, rules.NewService())
	claims := auth.AuthClaims{UserID: "u", TenantID: "t", Role: auth.RoleAdmin}

	_, err := svc.CreateContract(context.Background(), claims, CreateContractInput{Numero: "1", Tipo: "A", Data: map[string]any{}})
	if err != nil {
		t.Fatalf("first create should pass: %v", err)
	}

	_, err = svc.CreateContract(context.Background(), claims, CreateContractInput{Numero: "1", Tipo: "A", Data: map[string]any{}})
	if err == nil {
		t.Fatalf("expected conflict error")
	}
	if fmt.Sprintf("%T", err) == "" {
		t.Fatalf("expected concrete error")
	}
}
