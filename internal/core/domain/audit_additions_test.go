package domain_test

import (
	"testing"
	"time"

	"github.com/opensource-easypanel/openpanel/internal/core/domain"
)

func TestBackupScheduleValidation(t *testing.T) {
	valid := &domain.BackupSchedule{
		ID:                  "bs-1",
		ProjectName:         "proj-a",
		ServiceName:         "svc-db",
		Type:                "database",
		TargetName:          "prod_db",
		Schedule:            "0 2 * * *",
		Enabled:             true,
		StorageProviderID:   "sp-s3",
		StorageProviderPath: "/backups/db",
		Retention:           7,
		CreatedAt:           time.Now().UTC(),
		UpdatedAt:           time.Now().UTC(),
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("expected valid backup schedule, got: %v", err)
	}

	invalidNoID := &domain.BackupSchedule{
		ProjectName:       "proj-a",
		ServiceName:       "svc-db",
		Schedule:          "0 2 * * *",
		StorageProviderID: "sp-s3",
	}
	if err := invalidNoID.Validate(); err != domain.ErrValidation {
		t.Errorf("expected ErrValidation for missing ID, got %v", err)
	}

	invalidNoSchedule := &domain.BackupSchedule{
		ID:                "bs-2",
		ProjectName:       "proj-a",
		ServiceName:       "svc-db",
		StorageProviderID: "sp-s3",
	}
	if err := invalidNoSchedule.Validate(); err != domain.ErrValidation {
		t.Errorf("expected ErrValidation for missing Schedule, got %v", err)
	}
}

func TestStorageProviderMultiProtocol(t *testing.T) {
	sftp := &domain.StorageProvider{
		ID:       "sp-sftp",
		Name:     "Hetzner SFTP",
		Type:     domain.StorageProviderTypeSFTP,
		Host:     "sftp.hetzner.com",
		Port:     22,
		Username: "backup_user",
		Password: "secretPassword",
	}
	if err := sftp.Validate(); err != nil {
		t.Fatalf("expected valid SFTP provider, got: %v", err)
	}

	s3R2 := &domain.StorageProvider{
		ID:           "sp-r2",
		Name:         "Cloudflare R2",
		Type:         domain.StorageProviderTypeS3,
		Subtype:      "cloudflare-r2",
		Endpoint:     "https://account.r2.cloudflarestorage.com",
		Bucket:       "db-backups",
		AccessKey:    "r2-key",
		SecretKey:    "r2-secret",
		StorageClass: "STANDARD",
	}
	if err := s3R2.Validate(); err != nil {
		t.Fatalf("expected valid S3 R2 provider, got: %v", err)
	}

	invalidSFTP := &domain.StorageProvider{
		ID:   "sp-sftp-bad",
		Name: "Bad SFTP",
		Type: domain.StorageProviderTypeSFTP,
	}
	if err := invalidSFTP.Validate(); err != domain.ErrValidation {
		t.Errorf("expected ErrValidation for missing host/user, got %v", err)
	}
}

func TestMiddlewareValidation(t *testing.T) {
	mw := &domain.Middleware{
		ID:   "mw-auth",
		Name: "basic-auth-admin",
		Type: "basicAuth",
		Config: map[string]interface{}{
			"users": []string{"admin:$apr1$..."},
		},
	}
	if err := mw.Validate(); err != nil {
		t.Fatalf("expected valid middleware, got: %v", err)
	}

	invalidMW := &domain.Middleware{
		ID: "mw-bad",
	}
	if err := invalidMW.Validate(); err != domain.ErrValidation {
		t.Errorf("expected ErrValidation for missing Name/Type, got %v", err)
	}
}

func TestTemplateSchemaComposition(t *testing.T) {
	schema := domain.TemplateSchema{
		Services: []domain.TemplateServiceEntry{
			{
				Type: domain.ServiceTypeApp,
				Data: map[string]interface{}{
					"serviceName": "web",
					"image":       "nginx:alpine",
				},
			},
			{
				Type: domain.ServiceTypeRedis,
				Data: map[string]interface{}{
					"serviceName": "cache",
					"image":       "redis:7",
				},
			},
		},
	}
	if len(schema.Services) != 2 {
		t.Fatalf("expected 2 services in template schema, got %d", len(schema.Services))
	}
	if schema.Services[0].Type != domain.ServiceTypeApp {
		t.Errorf("expected first service to be app, got %s", schema.Services[0].Type)
	}
}
