package orpc

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Standard oRPC error codes matching @orpc/client specification.
const (
	CodeBadRequest          = "BAD_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
)

// RequestEnvelope models the incoming oRPC / SuperJSON request envelope.
type RequestEnvelope struct {
	JSON json.RawMessage `json:"json"`
	Meta []any           `json:"meta,omitempty"`
}

// ErrorData represents the inner error payload expected by oRPC.
type ErrorData struct {
	Defined bool   `json:"defined"`
	Code    string `json:"code"`
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// ErrorEnvelope models the top-level error response envelope.
type ErrorEnvelope struct {
	JSON ErrorData `json:"json"`
}

// SuccessEnvelope models the top-level success response envelope.
type SuccessEnvelope struct {
	JSON any `json:"json"`
}

// ORPCError is an error type that conveys HTTP status and oRPC error code.
type ORPCError struct {
	Status  int
	Code    string
	Message string
	Data    any
}

func (e *ORPCError) Error() string {
	return fmt.Sprintf("[%s %d] %s", e.Code, e.Status, e.Message)
}

// NewError creates an ORPCError with the specified status, code, and message.
func NewError(status int, code, message string) *ORPCError {
	return &ORPCError{Status: status, Code: code, Message: message}
}

// NewBadRequest returns a 400 BAD_REQUEST error.
func NewBadRequest(message string) *ORPCError {
	return NewError(http.StatusBadRequest, CodeBadRequest, message)
}

// NewUnauthorized returns a 401 UNAUTHORIZED error.
func NewUnauthorized(message string) *ORPCError {
	return NewError(http.StatusUnauthorized, CodeUnauthorized, message)
}

// NewForbidden returns a 403 FORBIDDEN error.
func NewForbidden(message string) *ORPCError {
	return NewError(http.StatusForbidden, CodeForbidden, message)
}

// NewNotFound returns a 404 NOT_FOUND error.
func NewNotFound(message string) *ORPCError {
	return NewError(http.StatusNotFound, CodeNotFound, message)
}

// NewConflict returns a 409 CONFLICT error.
func NewConflict(message string) *ORPCError {
	return NewError(http.StatusConflict, CodeConflict, message)
}

// NewInternalError returns a 500 INTERNAL_SERVER_ERROR.
func NewInternalError(message string) *ORPCError {
	return NewError(http.StatusInternalServerError, CodeInternalServerError, message)
}

// WriteSuccess writes an oRPC success envelope ({"json": data}) with HTTP 200.
func WriteSuccess(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	env := SuccessEnvelope{JSON: data}
	_ = json.NewEncoder(w).Encode(env)
}

// WriteError writes an oRPC error envelope with the provided status and code.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	env := ErrorEnvelope{
		JSON: ErrorData{
			Defined: false,
			Code:    code,
			Status:  status,
			Message: message,
			Data:    nil,
		},
	}
	_ = json.NewEncoder(w).Encode(env)
}
