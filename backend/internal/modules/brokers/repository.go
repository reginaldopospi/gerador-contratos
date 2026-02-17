package brokers

import "context"

type Repository interface {
	List(ctx context.Context, tenantID string) ([]Broker, error)
	GetByID(ctx context.Context, tenantID, brokerID string) (*Broker, error)
	Create(ctx context.Context, broker Broker) error
	Update(ctx context.Context, broker Broker) error
	Delete(ctx context.Context, tenantID, brokerID string) error
}
