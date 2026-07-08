package repository

import (
	"context"

	"api-source-proxy/internal/model"
)

type ApiSourceRepository interface {
	Create(ctx context.Context, source *model.ApiSource) error
	GetByID(ctx context.Context, id string) (*model.ApiSource, error)
	GetByName(ctx context.Context, name string) (*model.ApiSource, error)
	List(ctx context.Context) ([]model.ApiSource, error)
	Update(ctx context.Context, source *model.ApiSource) error
	Delete(ctx context.Context, id string) error
}
