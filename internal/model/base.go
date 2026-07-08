package model

import "time"

type Base struct {
	CreatedAt time.Time `json:"created_at" db:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" bson:"updated_at"`
}
