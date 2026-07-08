package errors

type AppError struct {
	Err     error
	Message string
	Code    int
}

func (e *AppError) Error() string {
	return e.Message
}

var (
	ErrNotFound     = &AppError{Code: 404, Message: "Resource not found"}
	ErrConflict     = &AppError{Code: 409, Message: "Resource already exists"}
	ErrUnauthorized = &AppError{Code: 401, Message: "Unauthorized"}
	ErrForbidden    = &AppError{Code: 403, Message: "Forbidden"}
	ErrValidation   = &AppError{Code: 422, Message: "Validation failed"}
	ErrInternal     = &AppError{Code: 500, Message: "Internal server error"}
)

func Wrap(err *AppError, msg string) *AppError {
	return &AppError{
		Err:     err,
		Message: msg,
		Code:    err.Code,
	}
}
