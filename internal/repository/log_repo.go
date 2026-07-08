package repository

import (
	"context"

	"api-source-proxy/internal/model"
)

type LogRepository interface {
	Insert(ctx context.Context, log *model.ActivityLog) error
	List(ctx context.Context, filter map[string]interface{}, page, limit int) ([]model.ActivityLog, int64, error)
}
