package domain

import (
	"strings"
	"time"
)

// Project represents a logical grouping of services, domains, and resources.
type Project struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Validate ensures the project attributes meet domain constraints.
func (p *Project) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(p.Name) == "" {
		return ErrValidation
	}
	return nil
}
