package model

import "time"

type ActivityLog struct {
	ID            string    `json:"id" bson:"_id"`
	UserID        string    `json:"user_id" bson:"user_id"`
	ApiKeyID      string    `json:"api_key_id" bson:"api_key_id"`
	Username      string    `json:"username" bson:"username"`
	Client        string    `json:"client" bson:"client"`
	ApiSourceName string    `json:"api_source_name" bson:"api_source_name"`
	Method        string    `json:"method" bson:"method"`
	Path          string    `json:"path" bson:"path"`
	RequestBody   string    `json:"request_body,omitempty" bson:"request_body,omitempty"`
	ResponseCode  int       `json:"response_code" bson:"response_code"`
	ResponseBody  string    `json:"response_body,omitempty" bson:"response_body,omitempty"`
	ClientIP      string    `json:"client_ip" bson:"client_ip"`
	DurationMs    int64     `json:"duration_ms" bson:"duration_ms"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
}
