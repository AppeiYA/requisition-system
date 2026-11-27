package config

import (
	"testing"
	"time"
)

var attachmentEnvVars = []string{
	"ATTACHMENT_STORAGE_DRIVER",
	"ATTACHMENT_LOCAL_ROOT",
	"ATTACHMENT_MAX_SIZE_BYTES",
	"ATTACHMENT_MAX_PER_REF",
	"ATTACHMENT_SIGNED_URL_TTL",
	"STAGED_UPLOAD_TTL",
	"ATTACHMENT_STAGING_TTL", // legacy fallback for STAGED_UPLOAD_TTL
	"STAGED_UPLOAD_MAX_PENDING_PER_USER",
	"STAGED_UPLOAD_MAX_BYTES_PER_USER",
	"ATTACHMENT_RECONCILE_BATCH_SIZE",
	"ATTACHMENT_RECONCILE_MAX_ATTEMPTS",
	"ATTACHMENT_RECONCILE_BACKOFF",
	"ATTACHMENT_RECONCILE_SCHEDULE",
	"ATTACHMENT_RECONCILE_STUCK_AFTER",
	"ATTACHMENT_CLEANUP_SCHEDULE",
	"ATTACHMENT_S3_BUCKET",
	"ATTACHMENT_S3_REGION",
	"ATTACHMENT_S3_ENDPOINT",
	"ATTACHMENT_S3_ACCESS_KEY_ID",
	"ATTACHMENT_S3_SECRET_ACCESS_KEY",
	"ATTACHMENT_S3_FORCE_PATH_STYLE",
	"ATTACHMENT_S3_USE_SSL",
}

func resetAttachmentEnv(t *testing.T) {
	t.Helper()
	for _, key := range attachmentEnvVars {
		unsetEnv(t, key)
	}
}

func TestLoadAttachmentConfigDefaults(t *testing.T) {
	resetAttachmentEnv(t)

	cfg, err := LoadAttachmentConfig()
	if err != nil {
		t.Fatalf("LoadAttachmentConfig() error = %v", err)
	}
	if cfg.StorageDriver != "local" {
		t.Errorf("StorageDriver = %q, want local", cfg.StorageDriver)
	}
	if cfg.MaxSizeBytes != 10<<20 {
		t.Errorf("MaxSizeBytes = %d, want %d", cfg.MaxSizeBytes, 10<<20)
	}
	if cfg.MaxPerRef != 10 {
		t.Errorf("MaxPerRef = %d, want 10", cfg.MaxPerRef)
	}
	if cfg.SignedURLTTL != 5*time.Minute {
		t.Errorf("SignedURLTTL = %v, want 5m", cfg.SignedURLTTL)
	}
	if cfg.StagingTTL != 24*time.Hour {
		t.Errorf("StagingTTL = %v, want 24h", cfg.StagingTTL)
	}
	if cfg.StagingMaxPendingPerUser != 20 {
		t.Errorf("StagingMaxPendingPerUser = %d, want 20", cfg.StagingMaxPendingPerUser)
	}
	if cfg.StagingMaxBytesPerUser != 0 {
		t.Errorf("StagingMaxBytesPerUser = %d, want 0", cfg.StagingMaxBytesPerUser)
	}
	if cfg.ReconcileBatchSize != 50 {
		t.Errorf("ReconcileBatchSize = %d, want 50", cfg.ReconcileBatchSize)
	}
	if cfg.ReconcileMaxAttempts != 5 {
		t.Errorf("ReconcileMaxAttempts = %d, want 5", cfg.ReconcileMaxAttempts)
	}
	if cfg.ReconcileBackoff != time.Minute {
		t.Errorf("ReconcileBackoff = %v, want 1m", cfg.ReconcileBackoff)
	}
	if cfg.ReconcileSchedule != "@every 15s" {
		t.Errorf("ReconcileSchedule = %q, want @every 15s", cfg.ReconcileSchedule)
	}
	if cfg.ReconcileStuckAfter != 15*time.Minute {
		t.Errorf("ReconcileStuckAfter = %v, want 15m", cfg.ReconcileStuckAfter)
	}
	if cfg.CleanupSchedule != "@every 5m" {
		t.Errorf("CleanupSchedule = %q, want @every 5m", cfg.CleanupSchedule)
	}
}

func TestLoadAttachmentConfigStagingOverrides(t *testing.T) {
	resetAttachmentEnv(t)
	t.Setenv("STAGED_UPLOAD_TTL", "12h")
	t.Setenv("ATTACHMENT_STAGING_TTL", "6h") // must lose to the new name
	t.Setenv("STAGED_UPLOAD_MAX_PENDING_PER_USER", "7")
	t.Setenv("STAGED_UPLOAD_MAX_BYTES_PER_USER", "8192")

	cfg, err := LoadAttachmentConfig()
	if err != nil {
		t.Fatalf("LoadAttachmentConfig() error = %v", err)
	}
	if cfg.StagingTTL != 12*time.Hour {
		t.Errorf("StagingTTL = %v, want 12h", cfg.StagingTTL)
	}
	if cfg.StagingMaxPendingPerUser != 7 {
		t.Errorf("StagingMaxPendingPerUser = %d, want 7", cfg.StagingMaxPendingPerUser)
	}
	if cfg.StagingMaxBytesPerUser != 8192 {
		t.Errorf("StagingMaxBytesPerUser = %d, want 8192", cfg.StagingMaxBytesPerUser)
	}
}

func TestLoadAttachmentConfigLegacyStagingTTL(t *testing.T) {
	resetAttachmentEnv(t)
	t.Setenv("ATTACHMENT_STAGING_TTL", "6h")

	cfg, err := LoadAttachmentConfig()
	if err != nil {
		t.Fatalf("LoadAttachmentConfig() error = %v", err)
	}
	if cfg.StagingTTL != 6*time.Hour {
		t.Errorf("StagingTTL = %v, want 6h", cfg.StagingTTL)
	}
}

func TestLoadAttachmentConfigValidation(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
	}{
		{name: "unknown driver", env: map[string]string{"ATTACHMENT_STORAGE_DRIVER": "ftp"}},
		{name: "zero size", env: map[string]string{"ATTACHMENT_MAX_SIZE_BYTES": "0"}},
		{name: "zero count", env: map[string]string{"ATTACHMENT_MAX_PER_REF": "0"}},
		{name: "zero ttl", env: map[string]string{"ATTACHMENT_SIGNED_URL_TTL": "0s"}},
		{name: "zero staged upload ttl", env: map[string]string{"STAGED_UPLOAD_TTL": "0s"}},
		{name: "zero legacy staging ttl", env: map[string]string{"ATTACHMENT_STAGING_TTL": "0s"}},
		{name: "zero staged pending quota", env: map[string]string{"STAGED_UPLOAD_MAX_PENDING_PER_USER": "0"}},
		{name: "negative staged byte quota", env: map[string]string{"STAGED_UPLOAD_MAX_BYTES_PER_USER": "-1"}},
		{name: "zero reconcile batch", env: map[string]string{"ATTACHMENT_RECONCILE_BATCH_SIZE": "0"}},
		{name: "zero reconcile attempts", env: map[string]string{"ATTACHMENT_RECONCILE_MAX_ATTEMPTS": "0"}},
		{name: "zero reconcile backoff", env: map[string]string{"ATTACHMENT_RECONCILE_BACKOFF": "0s"}},
		{name: "empty reconcile schedule", env: map[string]string{"ATTACHMENT_RECONCILE_SCHEDULE": "  "}},
		{name: "zero stuck after", env: map[string]string{"ATTACHMENT_RECONCILE_STUCK_AFTER": "0s"}},
		{name: "empty cleanup schedule", env: map[string]string{"ATTACHMENT_CLEANUP_SCHEDULE": "  "}},
		{name: "s3 without bucket", env: map[string]string{"ATTACHMENT_STORAGE_DRIVER": "s3"}},
		{
			name: "s3 without region or endpoint",
			env: map[string]string{
				"ATTACHMENT_STORAGE_DRIVER": "s3",
				"ATTACHMENT_S3_BUCKET":      "attachments",
			},
		},
		{
			name: "half credentials",
			env:  map[string]string{"ATTACHMENT_S3_ACCESS_KEY_ID": "key"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetAttachmentEnv(t)
			for key, value := range tt.env {
				t.Setenv(key, value)
			}
			if _, err := LoadAttachmentConfig(); err == nil {
				t.Error("expected configuration error")
			}
		})
	}
}

func TestLoadAttachmentConfigS3(t *testing.T) {
	resetAttachmentEnv(t)
	t.Setenv("ATTACHMENT_STORAGE_DRIVER", "S3")
	t.Setenv("ATTACHMENT_S3_BUCKET", "attachments")
	t.Setenv("ATTACHMENT_S3_REGION", "eu-west-1")
	t.Setenv("ATTACHMENT_S3_FORCE_PATH_STYLE", "true")

	cfg, err := LoadAttachmentConfig()
	if err != nil {
		t.Fatalf("LoadAttachmentConfig() error = %v", err)
	}
	if cfg.StorageDriver != "s3" {
		t.Errorf("StorageDriver = %q, want s3", cfg.StorageDriver)
	}
	if !cfg.S3ForcePathStyle {
		t.Error("S3ForcePathStyle = false, want true")
	}
}

func TestLoadAttachmentConfigS3CustomEndpointWithoutRegion(t *testing.T) {
	resetAttachmentEnv(t)
	t.Setenv("ATTACHMENT_STORAGE_DRIVER", "s3")
	t.Setenv("ATTACHMENT_S3_BUCKET", "attachments")
	t.Setenv("ATTACHMENT_S3_ENDPOINT", "http://localhost:9000")
	t.Setenv("ATTACHMENT_S3_USE_SSL", "false")

	cfg, err := LoadAttachmentConfig()
	if err != nil {
		t.Fatalf("LoadAttachmentConfig() error = %v", err)
	}
	if cfg.S3UseSSL {
		t.Error("S3UseSSL = true, want false")
	}
	if cfg.S3Endpoint != "http://localhost:9000" {
		t.Errorf("S3Endpoint = %q", cfg.S3Endpoint)
	}
}
