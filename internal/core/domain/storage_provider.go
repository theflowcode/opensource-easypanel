package domain

import (
	"strings"
	"time"
)

// StorageProvider types for database backups.
const (
	StorageProviderTypeLocal   = "local"
	StorageProviderTypeS3      = "s3"
	StorageProviderTypeSFTP    = "sftp"
	StorageProviderTypeFTP     = "ftp"
	StorageProviderTypeDropbox = "dropbox"
	StorageProviderTypeGoogle  = "google"
)

// StorageProvider represents a storage target for database backups and snapshots.
type StorageProvider struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"` // "local", "s3", "sftp", "ftp", "dropbox", "google"
	Subtype      string    `json:"subtype,omitempty"` // for S3: "aws", "backblaze", "cloudflare-r2", "digital-ocean", "wasabi", "other"
	Path         string    `json:"path,omitempty"`
	Endpoint     string    `json:"endpoint,omitempty"`
	Bucket       string    `json:"bucket,omitempty"`
	Region       string    `json:"region,omitempty"`
	AccessKey    string    `json:"accessKey,omitempty"`
	SecretKey    string    `json:"secretKey,omitempty"`
	Host         string    `json:"host,omitempty"`
	Port         int       `json:"port,omitempty"`
	Username     string    `json:"username,omitempty"`
	Password     string    `json:"password,omitempty"`
	StorageClass string    `json:"storageClass,omitempty"`
	RefreshToken string    `json:"refreshToken,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
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
	if sp.Type == StorageProviderTypeSFTP || sp.Type == StorageProviderTypeFTP {
		if strings.TrimSpace(sp.Host) == "" || strings.TrimSpace(sp.Username) == "" {
			return ErrValidation
		}
	}
	return nil
}
