package repository

import (
	"context"

	"api-source-proxy/internal/model"
)

type ApiKeyRepository interface {
	Create(ctx context.Context, key *model.ApiKey) error
	GetByID(ctx context.Context, id string) (*model.ApiKey, error)
	GetByUserID(ctx context.Context, userID string) ([]model.ApiKey, error)
	GetByKeyHash(ctx context.Context, keyHash string) (*model.ApiKey, error)
	List(ctx context.Context) ([]model.ApiKey, error)
	Deactivate(ctx context.Context, id string) error
	UpdateLastUsed(ctx context.Context, id string) error
}
