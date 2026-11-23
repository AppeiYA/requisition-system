// Package audit records administrative mutations in an append-only log.
//
// The Recorder is called from inside the use case that performs the mutation,
// on the same context, so that the audit entry is written in the same
// transaction: either both the change and its audit trail commit, or neither
// does.
package audit

import (
	"context"
	"time"
)

// Entry is one administrative action. Before and After are serialised to JSON
// by the recorder and must never contain password hashes or token hashes.
type Entry struct {
	ID         string
	ActorID    string
	Action     string
	TargetType string
	TargetID   string
	Before     any
	After      any
	Reason     string
	RequestID  string
	IPAddress  string
	UserAgent  string
	CreatedAt  time.Time
}

// Recorder persists audit entries.
type Recorder interface {
	Record(ctx context.Context, entry Entry) error
}

// IDGenerator produces prefixed identifiers such as "aud-...".
type IDGenerator interface {
	GenerateID(prefix string) (string, error)
}
