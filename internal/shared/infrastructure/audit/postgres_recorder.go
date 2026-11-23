package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/AppeiYA/requisition-system/internal/shared/infrastructure/db"
)

const insertEntryQuery = `
	INSERT INTO admin_audit_log (
		id, actor_id, action, target_type, target_id,
		before_state, after_state, reason, request_id,
		ip_address, user_agent, created_at
	) VALUES (
		:id, :actor_id, :action, :target_type, :target_id,
		:before_state, :after_state, :reason, :request_id,
		:ip_address, :user_agent, :created_at
	)
`

// PostgresRecorder writes entries to admin_audit_log. It resolves the executor
// from the context, so when the caller runs inside
// db.TransactionManager.WithinTransaction the entry joins that transaction.
type PostgresRecorder struct {
	repository  db.Repository
	idGenerator IDGenerator
}

func NewPostgresRecorder(repository db.Repository, idGenerator IDGenerator) *PostgresRecorder {
	return &PostgresRecorder{
		repository:  repository,
		idGenerator: idGenerator,
	}
}

var _ Recorder = (*PostgresRecorder)(nil)

func (r *PostgresRecorder) Record(ctx context.Context, entry Entry) error {
	if entry.ActorID == "" {
		return fmt.Errorf("audit: actor id is required")
	}
	if entry.Action == "" {
		return fmt.Errorf("audit: action is required")
	}
	if entry.TargetType == "" || entry.TargetID == "" {
		return fmt.Errorf("audit: target type and id are required")
	}

	if entry.ID == "" {
		id, err := r.idGenerator.GenerateID("aud")
		if err != nil {
			return fmt.Errorf("audit: generate id: %w", err)
		}
		entry.ID = id
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}

	if meta, ok := RequestMetaFromContext(ctx); ok {
		if entry.RequestID == "" {
			entry.RequestID = meta.RequestID
		}
		if entry.IPAddress == "" {
			entry.IPAddress = meta.IPAddress
		}
		if entry.UserAgent == "" {
			entry.UserAgent = meta.UserAgent
		}
	}

	before, err := marshalState(entry.Before)
	if err != nil {
		return err
	}
	after, err := marshalState(entry.After)
	if err != nil {
		return err
	}

	var ip any
	if entry.IPAddress != "" {
		if net.ParseIP(entry.IPAddress) == nil {
			return fmt.Errorf("audit: invalid ip address %q", entry.IPAddress)
		}
		ip = entry.IPAddress
	}

	executor := r.repository.Executor(ctx)
	_, err = executor.NamedExecContext(ctx, insertEntryQuery, map[string]any{
		"id":           entry.ID,
		"actor_id":     entry.ActorID,
		"action":       entry.Action,
		"target_type":  entry.TargetType,
		"target_id":    entry.TargetID,
		"before_state": before,
		"after_state":  after,
		"reason":       nullableString(entry.Reason),
		"request_id":   nullableString(entry.RequestID),
		"ip_address":   ip,
		"user_agent":   nullableString(entry.UserAgent),
		"created_at":   entry.CreatedAt,
	})
	if err != nil {
		return fmt.Errorf("audit: insert entry: %w", err)
	}

	return nil
}

// marshalState returns a JSON string, or nil for an absent state. It returns a
// string rather than []byte so lib/pq sends it as text and PostgreSQL parses it
// as jsonb.
func marshalState(state any) (any, error) {
	if state == nil {
		return nil, nil
	}

	raw, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("audit: marshal state: %w", err)
	}

	return string(raw), nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
