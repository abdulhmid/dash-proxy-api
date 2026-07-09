package model

import "time"

type ApiKey struct {
	ID          string    `json:"id" db:"id" bson:"_id"`
	UserID      string    `json:"user_id" db:"user_id" bson:"user_id"`
	KeyHash     string    `json:"-" db:"key_hash" bson:"key_hash"`
	KeyPrefix   string    `json:"key_prefix" db:"key_prefix" bson:"key_prefix"`
	Name        string    `json:"name" db:"name" bson:"name"`
	IsActive    bool      `json:"is_active" db:"is_active" bson:"is_active"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty" db:"last_used_at" bson:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at" bson:"expires_at,omitempty"`
	CreatedAt   time.Time `json:"created_at" db:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at" bson:"updated_at"`
}
