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

type ApiKeyRepo struct {
	db *sqlx.DB
}

func NewApiKeyRepo(db *sqlx.DB) *ApiKeyRepo {
	return &ApiKeyRepo{db: db}
}

func (r *ApiKeyRepo) Create(ctx context.Context, key *model.ApiKey) error {
	key.ID = uuid.New().String()
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO api_keys (id, user_id, key_hash, key_prefix, name, is_active, expires_at, created_at, updated_at)
		VALUES (:id, :user_id, :key_hash, :key_prefix, :name, :is_active, :expires_at, :created_at, :updated_at)
	`, key)
	if err != nil {
		return fmt.Errorf("create api key: %w", err)
	}
	return nil
}

func (r *ApiKeyRepo) GetByID(ctx context.Context, id string) (*model.ApiKey, error) {
	var key model.ApiKey
	err := r.db.GetContext(ctx, &key, "SELECT * FROM api_keys WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get api key by id: %w", err)
	}
	return &key, nil
}

func (r *ApiKeyRepo) GetByUserID(ctx context.Context, userID string) ([]model.ApiKey, error) {
	var keys []model.ApiKey
	err := r.db.SelectContext(ctx, &keys, "SELECT * FROM api_keys WHERE user_id = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		return nil, fmt.Errorf("get api keys by user id: %w", err)
	}
	return keys, nil
}

func (r *ApiKeyRepo) GetByKeyHash(ctx context.Context, keyHash string) (*model.ApiKey, error) {
	var key model.ApiKey
	err := r.db.GetContext(ctx, &key, "SELECT * FROM api_keys WHERE key_hash = $1", keyHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get api key by hash: %w", err)
	}
	return &key, nil
}

func (r *ApiKeyRepo) List(ctx context.Context) ([]model.ApiKey, error) {
	var keys []model.ApiKey
	err := r.db.SelectContext(ctx, &keys, "SELECT * FROM api_keys ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	return keys, nil
}

func (r *ApiKeyRepo) Deactivate(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE api_keys SET is_active = false, updated_at = $1 WHERE id = $2", time.Now(), id)
	if err != nil {
		return fmt.Errorf("deactivate api key: %w", err)
	}
	return nil
}

func (r *ApiKeyRepo) UpdateLastUsed(ctx context.Context, id string) error {
	now := time.Now()
	_, err := r.db.ExecContext(ctx, "UPDATE api_keys SET last_used_at = $1 WHERE id = $2", now, id)
	if err != nil {
		return fmt.Errorf("update last used: %w", err)
	}
	return nil
}
