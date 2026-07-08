package dto

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role" validate:"omitempty,oneof=admin user"`
}

type UpdateUserRequest struct {
	Username string `json:"username,omitempty" validate:"omitempty,min=3,max=50"`
	Email    string `json:"email,omitempty" validate:"omitempty,email"`
	Password string `json:"password,omitempty" validate:"omitempty,min=8"`
	Role     string `json:"role,omitempty" validate:"omitempty,oneof=admin user"`
}

type CreateApiKeyRequest struct {
	UserID    string `json:"user_id" validate:"required,uuid"`
	Name      string `json:"name" validate:"required,min=1,max=100"`
	ExpiryDay *int   `json:"expiry_day,omitempty"`
}

type CreateApiKeyResponse struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	KeyPrefix string `json:"key_prefix"`
	Name      string `json:"name"`
}

type CreateApiSourceRequest struct {
	Name        string `json:"name,omitempty"`
	BaseURL     string `json:"base_url" validate:"required,url"`
	Username    string `json:"username,omitempty"`
	AuthType    string `json:"auth_type,omitempty" validate:"omitempty,oneof=none basic bearer api-key custom"`
	AuthHeaders string `json:"auth_headers,omitempty"`
	ExtraParams    string `json:"extra_params,omitempty"`
	AcceptedFields string `json:"accepted_fields,omitempty"`
	Method         string `json:"method,omitempty" validate:"omitempty,oneof=GET POST PUT DELETE PATCH"`
	TimeoutMs      int    `json:"timeout_ms,omitempty"`
}

type UpdateApiSourceRequest struct {
	Name           string `json:"name,omitempty"`
	BaseURL        string `json:"base_url,omitempty" validate:"omitempty,url"`
	Username       string `json:"username,omitempty"`
	AuthType       string `json:"auth_type,omitempty" validate:"omitempty,oneof=none basic bearer api-key custom"`
	AuthHeaders    string `json:"auth_headers,omitempty"`
	ExtraParams    string `json:"extra_params,omitempty"`
	AcceptedFields string `json:"accepted_fields,omitempty"`
	Method         string `json:"method,omitempty" validate:"omitempty,oneof=GET POST PUT DELETE PATCH"`
	TimeoutMs      *int   `json:"timeout_ms,omitempty"`
	IsActive       *bool  `json:"is_active,omitempty"`
}

type ProxyRequest struct {
	Msisdn string `json:"msisdn" form:"msisdn" validate:"required"`
}

type ProxyTestRequest struct {
	Method string                 `json:"method,omitempty"`
	Params map[string]interface{} `json:"params,omitempty"`
}

type PaginationQuery struct {
	Page  int `json:"page" form:"page"`
	Limit int `json:"limit" form:"limit"`
}

type LogFilterQuery struct {
	UserID        string `json:"user_id" form:"user_id"`
	ApiSourceName string `json:"api_source_name" form:"api_source_name"`
	StartDate     string `json:"start_date" form:"start_date"`
	EndDate       string `json:"end_date" form:"end_date"`
	Page          int    `json:"page" form:"page"`
	Limit         int    `json:"limit" form:"limit"`
}
