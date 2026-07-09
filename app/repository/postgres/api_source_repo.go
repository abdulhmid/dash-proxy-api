package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"api-source-proxy/app/model"
	apperrors "api-source-proxy/pkg/errors"
)

type ApiSourceRepo struct {
	db *sqlx.DB
}

func NewApiSourceRepo(db *sqlx.DB) *ApiSourceRepo {
	return &ApiSourceRepo{db: db}
}

func (r *ApiSourceRepo) Create(ctx context.Context, source *model.ApiSource) error {
	source.ID = uuid.New().String()
	source.CreatedAt = time.Now()
	source.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO api_sources (id, name, base_url, username, auth_type, auth_headers, extra_params, accepted_fields, method, timeout_ms, is_active, created_at, updated_at)
		VALUES (:id, :name, :base_url, :username, :auth_type, :auth_headers, :extra_params, :accepted_fields, :method, :timeout_ms, :is_active, :created_at, :updated_at)
	`, source)
	if err != nil {
		return fmt.Errorf("create api source: %w", err)
	}
	return nil
}

func (r *ApiSourceRepo) GetByID(ctx context.Context, id string) (*model.ApiSource, error) {
	var source model.ApiSource
	err := r.db.GetContext(ctx, &source, "SELECT * FROM api_sources WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get api source by id: %w", err)
	}
	return &source, nil
}

func (r *ApiSourceRepo) GetByName(ctx context.Context, name string) (*model.ApiSource, error) {
	var source model.ApiSource
	err := r.db.GetContext(ctx, &source, "SELECT * FROM api_sources WHERE name = $1", name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get api source by name: %w", err)
	}
	return &source, nil
}

func (r *ApiSourceRepo) List(ctx context.Context) ([]model.ApiSource, error) {
	var sources []model.ApiSource
	err := r.db.SelectContext(ctx, &sources, "SELECT * FROM api_sources ORDER BY name ASC")
	if err != nil {
		return nil, fmt.Errorf("list api sources: %w", err)
	}
	return sources, nil
}

func (r *ApiSourceRepo) Update(ctx context.Context, source *model.ApiSource) error {
	source.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE api_sources SET name=:name, base_url=:base_url, username=:username,
		auth_type=:auth_type, auth_headers=:auth_headers, extra_params=:extra_params,
		accepted_fields=:accepted_fields, method=:method, timeout_ms=:timeout_ms,
		is_active=:is_active, updated_at=:updated_at
		WHERE id=:id
	`, source)
	if err != nil {
		return fmt.Errorf("update api source: %w", err)
	}
	return nil
}

func (r *ApiSourceRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM api_sources WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete api source: %w", err)
	}
	return nil
}
