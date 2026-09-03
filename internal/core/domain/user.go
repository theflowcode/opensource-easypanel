package domain

import (
	"strings"
	"time"
)

// User represents an authorized user in OpenSource Easypanel.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // never expose hash in JSON responses
	Role         string    `json:"role"` // "admin", "viewer"
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// Validate validates user fields.
func (u *User) Validate() error {
	if strings.TrimSpace(u.ID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(u.Email) == "" || !strings.Contains(u.Email, "@") {
		return ErrValidation
	}
	if strings.TrimSpace(u.PasswordHash) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(u.Role) == "" {
		return ErrValidation
	}
	return nil
}
