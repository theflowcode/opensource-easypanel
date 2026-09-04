package http

import (
	"strings"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/adapter/http/orpc"
	"github.com/opensource-easypanel/openpanel/internal/core/domain"
	"github.com/opensource-easypanel/openpanel/internal/core/port"
)

type createStorageProviderInput struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Subtype   string `json:"subtype"`
	Path      string `json:"path"`
	Endpoint  string `json:"endpoint"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
}

type destroyStorageProviderInput struct {
	ID string `json:"id"`
}

func toStorageProviderDTO(sp *domain.StorageProvider) map[string]any {
	secretMasked := ""
	if sp.SecretKey != "" {
		secretMasked = "********"
	}
	return map[string]any{
		"id":        sp.ID,
		"name":      sp.Name,
		"type":      sp.Type,
		"subtype":   sp.Subtype,
		"path":      sp.Path,
		"endpoint":  sp.Endpoint,
		"bucket":    sp.Bucket,
		"region":    sp.Region,
		"accessKey": sp.AccessKey,
		"secretKey": secretMasked,
		"createdAt": sp.CreatedAt.Format(time.RFC3339),
		"updatedAt": sp.UpdatedAt.Format(time.RFC3339),
	}
}

// registerServerStorageRoutes binds storage providers and database/volume backups.
func registerServerStorageRoutes(d *orpc.Dispatcher, db port.DatabasePort) {
	// Storage Providers
	listStorageProvidersHandler := func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		providers, err := db.ListStorageProviders(c.Context)
		if err != nil {
			return nil, err
		}
		dtos := make([]map[string]any, 0, len(providers))
		for _, sp := range providers {
			dtos = append(dtos, toStorageProviderDTO(sp))
		}
		return dtos, nil
	}
	d.Register("storageProviders/common/listStorageProviders", listStorageProvidersHandler)
	d.Register("storageProviders/common/list", listStorageProvidersHandler)

	d.Register("storageProviders/common/listOptions", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return []map[string]any{
			{"label": "Local", "value": "local"},
			{"label": "Amazon S3", "value": "s3"},
			{"label": "SFTP", "value": "sftp"},
		}, nil
	})

	createStorageProviderHandler := func(defaultType string) orpc.HandlerFunc {
		return func(c *orpc.Context) (any, error) {
			if err := c.RequireAdmin(); err != nil {
				return nil, err
			}
			in, err := orpc.Bind[createStorageProviderInput](c)
			if err != nil {
				return nil, err
			}
			spType := strings.TrimSpace(in.Type)
			if spType == "" {
				spType = defaultType
			}
			sp := &domain.StorageProvider{
				ID:        domain.NewID(),
				Name:      strings.TrimSpace(in.Name),
				Type:      spType,
				Subtype:   strings.TrimSpace(in.Subtype),
				Path:      in.Path,
				Endpoint:  in.Endpoint,
				Bucket:    in.Bucket,
				Region:    in.Region,
				AccessKey: in.AccessKey,
				SecretKey: in.SecretKey,
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}
			if err := db.CreateStorageProvider(c.Context, sp); err != nil {
				return nil, err
			}
			return toStorageProviderDTO(sp), nil
		}
	}

	d.Register("storageProviders/common/createStorageProvider", createStorageProviderHandler(""))
	d.Register("storageProviders/s3/createProvider", createStorageProviderHandler("s3"))
	d.Register("storageProviders/local/createProvider", createStorageProviderHandler("local"))
	d.Register("storageProviders/sftp/createProvider", createStorageProviderHandler("sftp"))

	destroyStorageProviderHandler := func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		in, err := orpc.Bind[destroyStorageProviderInput](c)
		if err != nil {
			return nil, err
		}
		if err := db.DeleteStorageProvider(c.Context, in.ID); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	}
	d.Register("storageProviders/common/destroyStorageProvider", destroyStorageProviderHandler)
	d.Register("storageProviders/s3/destroyProvider", destroyStorageProviderHandler)
	d.Register("storageProviders/local/destroyProvider", destroyStorageProviderHandler)
	d.Register("storageProviders/sftp/destroyProvider", destroyStorageProviderHandler)

	// Backups
	d.Register("databaseBackups/listDatabaseBackups", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return []any{}, nil
	})

	d.Register("databaseBackups/runDatabaseBackup", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})

	d.Register("volumeBackups/listVolumeBackups", func(c *orpc.Context) (any, error) {
		if err := c.RequireAuth(); err != nil {
			return nil, err
		}
		return []any{}, nil
	})

	d.Register("volumeBackups/runVolumeBackup", func(c *orpc.Context) (any, error) {
		if err := c.RequireAdmin(); err != nil {
			return nil, err
		}
		return map[string]bool{"success": true}, nil
	})
}
