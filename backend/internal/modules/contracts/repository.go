package contracts

import "context"

type Repository interface {
	CreateContract(ctx context.Context, contract Contract) error
	GetContract(ctx context.Context, tenantID, contractID string) (*Contract, error)
	ListContracts(ctx context.Context, tenantID string, filter ListContractsFilter) ([]Contract, error)

	CreateVersion(ctx context.Context, version ContractVersion) error
	GetLatestVersion(ctx context.Context, contractID string) (*ContractVersion, error)
	GetVersion(ctx context.Context, contractID string, versionNumber int) (*ContractVersion, error)
	ListVersions(ctx context.Context, contractID string) ([]ContractVersion, error)
}
