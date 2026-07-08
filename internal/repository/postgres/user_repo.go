package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperrors "github.com/teodosiopiera/api-source-proxy/pkg/errors"

	"github.com/teodosiopiera/api-source-proxy/internal/model"
)

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	user.ID = uuid.New().String()
	_, err := r.db.NamedExecContext(ctx, `
		INSERT INTO users (id, username, email, password_hash, role, is_active, created_at, updated_at)
		VALUES (:id, :username, :email, :password_hash, :role, :is_active, :created_at, :updated_at)
	`, user)
	if err != nil {
		if isDuplicate(err) {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE username = $1", username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()
	_, err := r.db.NamedExecContext(ctx, `
		UPDATE users SET username=:username, email=:email, password_hash=:password_hash,
		role=:role, is_active=:is_active, updated_at=:updated_at WHERE id=:id
	`, user)
	if err != nil {
		if isDuplicate(err) {
			return apperrors.ErrConflict
		}
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (r *UserRepo) List(ctx context.Context) ([]model.User, error) {
	var users []model.User
	err := r.db.SelectContext(ctx, &users, "SELECT * FROM users ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func isDuplicate(err error) bool {
	return err != nil && contains(err.Error(), "duplicate key")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
