package model

import "time"

type ApiSource struct {
	ID           string    `json:"id" db:"id" bson:"_id"`
	Name         string    `json:"name" db:"name" bson:"name"`
	BaseURL      string    `json:"base_url" db:"base_url" bson:"base_url"`
	Username     string    `json:"username" db:"username" bson:"username"`
	AuthType     string    `json:"auth_type" db:"auth_type" bson:"auth_type"`
	AuthHeaders  string    `json:"auth_headers" db:"auth_headers" bson:"auth_headers"`
	ExtraParams  string    `json:"extra_params" db:"extra_params" bson:"extra_params"`
	AcceptedFields string  `json:"accepted_fields" db:"accepted_fields" bson:"accepted_fields"`
	Method       string    `json:"method" db:"method" bson:"method"`
	TimeoutMs    int       `json:"timeout_ms" db:"timeout_ms" bson:"timeout_ms"`
	IsActive     bool      `json:"is_active" db:"is_active" bson:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at" bson:"updated_at"`
}
