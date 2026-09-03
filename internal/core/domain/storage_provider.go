package domain

import (
	"strings"
	"time"
)

// StorageProvider types for database backups.
const (
	StorageProviderTypeLocal = "local"
	StorageProviderTypeS3    = "s3"
)

// StorageProvider represents a storage target for database backups and snapshots.
type StorageProvider struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // "local" or "s3"
	Path      string    `json:"path,omitempty"`
	Endpoint  string    `json:"endpoint,omitempty"`
	Bucket    string    `json:"bucket,omitempty"`
	Region    string    `json:"region,omitempty"`
	AccessKey string    `json:"accessKey,omitempty"`
	SecretKey string    `json:"secretKey,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Validate ensures required storage provider fields are provided.
func (sp *StorageProvider) Validate() error {
	if strings.TrimSpace(sp.ID) == "" {
		return ErrValidation
	}
	if strings.TrimSpace(sp.Name) == "" {
		return ErrValidation
	}
	if sp.Type == "" {
		sp.Type = StorageProviderTypeLocal
	}
	if sp.Type == StorageProviderTypeLocal && strings.TrimSpace(sp.Path) == "" {
		sp.Path = "/etc/easypanel/backups"
	}
	if sp.Type == StorageProviderTypeS3 {
		if strings.TrimSpace(sp.Bucket) == "" || strings.TrimSpace(sp.Endpoint) == "" {
			return ErrValidation
		}
	}
	return nil
}
