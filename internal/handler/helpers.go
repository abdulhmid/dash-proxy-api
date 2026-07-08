package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	apperrors "github.com/teodosiopiera/api-source-proxy/pkg/errors"
	"github.com/teodosiopiera/api-source-proxy/pkg/response"
)

var validate = validator.New()

func decodeJSON(r *http.Request, v interface{}) error {
	ct := r.Header.Get("Content-Type")
	if strings.Contains(ct, "application/x-www-form-urlencoded") || strings.Contains(ct, "multipart/form-data") {
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("parse form: %w", err)
		}
		data := make(map[string]string)
		for k := range r.Form {
			data[k] = r.Form.Get(k)
		}
		jsonData, _ := json.Marshal(data)
		return json.Unmarshal(jsonData, v)
	}
	return json.NewDecoder(r.Body).Decode(v)
}

func formatValidationErrors(err error) []response.FieldError {
	var fieldErrors []response.FieldError
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range validationErrors {
			fieldErrors = append(fieldErrors, response.FieldError{
				Field:   strings.ToLower(fe.Field()),
				Message: fmt.Sprintf("Field '%s' failed validation: %s", fe.Field(), fe.Tag()),
			})
		}
	}
	return fieldErrors
}

func toAppError(err error) *apperrors.AppError {
	if appErr, ok := err.(*apperrors.AppError); ok {
		return appErr
	}
	return &apperrors.AppError{
		Code:    http.StatusInternalServerError,
		Message: "Internal server error",
	}
}

func extractNameFromURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "api-source"
	}
	host := u.Hostname()
	host = strings.ReplaceAll(host, "www.", "")
	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		host = parts[len(parts)-2]
	}
	re := regexp.MustCompile(`[^a-zA-Z0-9-]`)
	host = re.ReplaceAllString(host, "-")
	host = strings.Trim(host, "-")
	if host == "" {
		return "api-source"
	}
	return strings.ToLower(host)
}
