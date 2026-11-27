package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AttachmentConfig controls file storage for requisition attachments. The
// driver can be switched to an S3-compatible store with an environment change
// alone; no domain or use case code depends on the choice.
type AttachmentConfig struct {
	StorageDriver            string
	LocalRoot                string
	MaxSizeBytes             int64
	MaxPerRef                int
	SignedURLTTL             time.Duration
	StagingTTL               time.Duration
	StagingMaxPendingPerUser int
	StagingMaxBytesPerUser   int64

	// Reconciliation controls the outbox worker that moves claimed uploads to
	// their permanent keys. Schedule is a cronspec accepted by asynq.
	ReconcileBatchSize   int
	ReconcileMaxAttempts int
	ReconcileBackoff     time.Duration
	ReconcileSchedule    string
	ReconcileStuckAfter  time.Duration
	CleanupSchedule      string

	S3Bucket          string
	S3Region          string
	S3Endpoint        string
	S3AccessKeyID     string
	S3SecretAccessKey string
	S3ForcePathStyle  bool
	S3UseSSL          bool
}

func LoadAttachmentConfig() (*AttachmentConfig, error) {
	driver := strings.ToLower(strings.TrimSpace(getEnv("ATTACHMENT_STORAGE_DRIVER", "local")))
	if driver != "local" && driver != "s3" {
		return nil, fmt.Errorf("ATTACHMENT_STORAGE_DRIVER must be local or s3")
	}

	root := getEnv("ATTACHMENT_LOCAL_ROOT", "./var/attachments")
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("ATTACHMENT_LOCAL_ROOT must not be empty")
	}
	if info, err := os.Stat(root); err == nil && !info.IsDir() {
		return nil, fmt.Errorf("ATTACHMENT_LOCAL_ROOT must be a directory")
	}

	maxSize, err := getIntEnv("ATTACHMENT_MAX_SIZE_BYTES", 3<<20)
	if err != nil {
		return nil, err
	}
	if maxSize <= 0 {
		return nil, fmt.Errorf("ATTACHMENT_MAX_SIZE_BYTES must be positive")
	}

	maxPerRef, err := getIntEnv("ATTACHMENT_MAX_PER_REF", 10)
	if err != nil {
		return nil, err
	}
	if maxPerRef <= 0 {
		return nil, fmt.Errorf("ATTACHMENT_MAX_PER_REF must be positive")
	}

	ttl, err := getTimeDurationEnv("ATTACHMENT_SIGNED_URL_TTL", 5*time.Minute)
	if err != nil {
		return nil, err
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("ATTACHMENT_SIGNED_URL_TTL must be positive")
	}

	stagingTTLStr := getEnv("STAGED_UPLOAD_TTL", "")
	if stagingTTLStr == "" {
		stagingTTLStr = getEnv("ATTACHMENT_STAGING_TTL", "24h")
	}
	stagingTTL, err := time.ParseDuration(strings.TrimSpace(stagingTTLStr))
	if err != nil {
		return nil, fmt.Errorf("invalid STAGED_UPLOAD_TTL: %w", err)
	}
	if stagingTTL <= 0 {
		return nil, fmt.Errorf("STAGED_UPLOAD_TTL must be positive")
	}

	maxPendingPerUser, err := getIntEnv("STAGED_UPLOAD_MAX_PENDING_PER_USER", 20)
	if err != nil {
		return nil, err
	}
	if maxPendingPerUser <= 0 {
		return nil, fmt.Errorf("STAGED_UPLOAD_MAX_PENDING_PER_USER must be positive")
	}

	maxBytesPerUser, err := getIntEnv("STAGED_UPLOAD_MAX_BYTES_PER_USER", 0)
	if err != nil {
		return nil, err
	}
	if maxBytesPerUser < 0 {
		return nil, fmt.Errorf("STAGED_UPLOAD_MAX_BYTES_PER_USER must be non-negative")
	}

	reconcileBatchSize, err := getIntEnv("ATTACHMENT_RECONCILE_BATCH_SIZE", 50)
	if err != nil {
		return nil, err
	}
	if reconcileBatchSize <= 0 {
		return nil, fmt.Errorf("ATTACHMENT_RECONCILE_BATCH_SIZE must be positive")
	}
	reconcileMaxAttempts, err := getIntEnv("ATTACHMENT_RECONCILE_MAX_ATTEMPTS", 5)
	if err != nil {
		return nil, err
	}
	if reconcileMaxAttempts <= 0 {
		return nil, fmt.Errorf("ATTACHMENT_RECONCILE_MAX_ATTEMPTS must be positive")
	}
	reconcileBackoff, err := getTimeDurationEnv("ATTACHMENT_RECONCILE_BACKOFF", time.Minute)
	if err != nil {
		return nil, err
	}
	if reconcileBackoff <= 0 {
		return nil, fmt.Errorf("ATTACHMENT_RECONCILE_BACKOFF must be positive")
	}
	reconcileSchedule := strings.TrimSpace(getEnv("ATTACHMENT_RECONCILE_SCHEDULE", "@every 15s"))
	if reconcileSchedule == "" {
		return nil, fmt.Errorf("ATTACHMENT_RECONCILE_SCHEDULE must not be empty")
	}
	reconcileStuckAfter, err := getTimeDurationEnv("ATTACHMENT_RECONCILE_STUCK_AFTER", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	if reconcileStuckAfter <= 0 {
		return nil, fmt.Errorf("ATTACHMENT_RECONCILE_STUCK_AFTER must be positive")
	}
	cleanupSchedule := strings.TrimSpace(getEnv("ATTACHMENT_CLEANUP_SCHEDULE", "@every 5m"))
	if cleanupSchedule == "" {
		return nil, fmt.Errorf("ATTACHMENT_CLEANUP_SCHEDULE must not be empty")
	}

	bucket := getEnv("ATTACHMENT_S3_BUCKET", "")
	region := getEnv("ATTACHMENT_S3_REGION", "")
	endpoint := getEnv("ATTACHMENT_S3_ENDPOINT", "")
	accessKey := getEnv("ATTACHMENT_S3_ACCESS_KEY_ID", "")
	secretKey := getEnv("ATTACHMENT_S3_SECRET_ACCESS_KEY", "")
	forcePathStyle, err := getBoolEnv("ATTACHMENT_S3_FORCE_PATH_STYLE", false)
	if err != nil {
		return nil, err
	}
	useSSL, err := getBoolEnv("ATTACHMENT_S3_USE_SSL", true)
	if err != nil {
		return nil, err
	}

	if driver == "s3" && strings.TrimSpace(bucket) == "" {
		return nil, fmt.Errorf("ATTACHMENT_S3_BUCKET is required when the storage driver is s3")
	}
	if driver == "s3" && strings.TrimSpace(endpoint) == "" && strings.TrimSpace(region) == "" {
		return nil, fmt.Errorf("ATTACHMENT_S3_REGION is required when no custom endpoint is set")
	}
	if (accessKey == "") != (secretKey == "") {
		return nil, fmt.Errorf("ATTACHMENT_S3_ACCESS_KEY_ID and ATTACHMENT_S3_SECRET_ACCESS_KEY must be set together")
	}

	return &AttachmentConfig{
		StorageDriver:            driver,
		LocalRoot:                filepath.Clean(root),
		MaxSizeBytes:             int64(maxSize),
		MaxPerRef:                maxPerRef,
		SignedURLTTL:             ttl,
		StagingTTL:               stagingTTL,
		StagingMaxPendingPerUser: maxPendingPerUser,
		StagingMaxBytesPerUser:   int64(maxBytesPerUser),
		ReconcileBatchSize:       reconcileBatchSize,
		ReconcileMaxAttempts:     reconcileMaxAttempts,
		ReconcileBackoff:         reconcileBackoff,
		ReconcileSchedule:        reconcileSchedule,
		ReconcileStuckAfter:      reconcileStuckAfter,
		CleanupSchedule:          cleanupSchedule,
		S3Bucket:                 bucket,
		S3Region:                 region,
		S3Endpoint:               endpoint,
		S3AccessKeyID:            accessKey,
		S3SecretAccessKey:        secretKey,
		S3ForcePathStyle:         forcePathStyle,
		S3UseSSL:                 useSSL,
	}, nil
}
