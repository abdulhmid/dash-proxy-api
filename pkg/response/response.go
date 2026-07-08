package response

import (
	"net/http"

	"github.com/go-chi/render"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type APIResponse struct {
	Status  string       `json:"status"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
	Errors  []FieldError `json:"errors,omitempty"`
	Meta    *Pagination  `json:"meta,omitempty"`
}

func Success(w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	render.JSON(w, r, APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func Created(w http.ResponseWriter, r *http.Request, message string, data interface{}) {
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	w.WriteHeader(statusCode)
	render.JSON(w, r, APIResponse{
		Status:  "error",
		Message: message,
	})
}

func ValidationError(w http.ResponseWriter, r *http.Request, errors []FieldError) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	render.JSON(w, r, APIResponse{
		Status:  "error",
		Message: "Validation failed",
		Errors:  errors,
	})
}

func Paginated(w http.ResponseWriter, r *http.Request, message string, data interface{}, page, limit, total int) {
	render.JSON(w, r, APIResponse{
		Status:  "success",
		Message: message,
		Data:    data,
		Meta: &Pagination{
			Page:  page,
			Limit: limit,
			Total: total,
		},
	})
}
