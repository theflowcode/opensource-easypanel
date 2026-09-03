package domain

import (
	"strings"
	"time"
)

// Session represents an authenticated user session or revocable API token.
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	TokenHash string    `json:"tokenHash"`
	ExpiresAt time.Time `json:"expiresAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// Validate ensures all required session fields are populated.
func (s *Session) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(s.UserID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(s.TokenHash) == "" {
		return ErrValidation
	}
	return nil
}

// IsExpired returns true if the session's expiry timestamp is in the past.
func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}
