package orpc

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

// Context encapsulates request metadata, auth state, and payload for an oRPC call.
type Context struct {
	context.Context
	Request        *http.Request
	ResponseWriter http.ResponseWriter
	Token          string
	User           *domain.User
	Session        *domain.Session
	Input          json.RawMessage
}

// Bind decodes the raw oRPC JSON input into the target struct.
// If input is empty or null, a zero-initialized value is returned.
func Bind[T any](c *Context) (*T, error) {
	var target T
	raw := strings.TrimSpace(string(c.Input))
	if raw == "" || raw == "null" || raw == "undefined" {
		return &target, nil
	}

	if err := json.Unmarshal(c.Input, &target); err != nil {
		return nil, NewBadRequest("invalid json input: " + err.Error())
	}
	return &target, nil
}

// RequireAuth ensures that an authenticated user is attached to the context.
func (c *Context) RequireAuth() error {
	if c.User == nil || c.Session == nil {
		return NewUnauthorized("unauthorized: valid session required")
	}
	if c.Session.IsExpired() {
		return NewUnauthorized("session expired")
	}
	return nil
}

// RequireAdmin ensures that the authenticated user possesses admin privileges.
func (c *Context) RequireAdmin() error {
	if err := c.RequireAuth(); err != nil {
		return err
	}
	if c.User.Role != domain.RoleAdmin {
		return NewForbidden("forbidden: administrative privileges required")
	}
	return nil
}
