package model

import "time"

type User struct {
	ID           string    `json:"id" db:"id" bson:"_id"`
	Username     string    `json:"username" db:"username" bson:"username"`
	Email        string    `json:"email" db:"email" bson:"email"`
	PasswordHash string    `json:"-" db:"password_hash" bson:"password_hash"`
	Role         string    `json:"role" db:"role" bson:"role"`
	IsActive     bool      `json:"is_active" db:"is_active" bson:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at" bson:"updated_at"`
}
