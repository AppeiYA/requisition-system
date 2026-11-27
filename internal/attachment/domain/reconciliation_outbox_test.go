package domain

import (
	"strings"
	"testing"
	"time"
)

func validReconciliationEvent(t *testing.T) *ReconciliationEvent {
	t.Helper()
	event, err := NewReconciliationEvent(
		"aevt-0123456789abcdef",
		"att-0123456789abcdef",
		"upl-0123456789abcdef",
		"staging/upl-0123456789abcdef",
		"requisition/req-0123456789abcdef/att-0123456789abcdef",
		strings.Repeat("a", ChecksumLength),
		time.Date(2026, 11, 27, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("NewReconciliationEvent() error = %v", err)
	}
	return event
}

func TestReconciliationEventFailureSchedule(t *testing.T) {
	now := time.Date(2026, 11, 27, 12, 30, 0, 0, time.UTC)
	tests := []struct {
		name          string
		attempts      int
		maxAttempts   int
		wantStatus    ReconciliationStatus
		wantAttempts  int
		wantNextDelay time.Duration
	}{
		{"first retry", 0, 5, ReconciliationStatusFailed, 1, time.Minute},
		{"second retry", 1, 5, ReconciliationStatusFailed, 2, 2 * time.Minute},
		{"third retry", 2, 5, ReconciliationStatusFailed, 3, 4 * time.Minute},
		{"last attempt is terminal", 4, 5, ReconciliationStatusDead, 5, 0},
		{"single attempt is terminal", 0, 1, ReconciliationStatusDead, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := validReconciliationEvent(t)
			event.attempts = tt.attempts

			status, attempts, next := event.FailureSchedule(tt.maxAttempts, time.Minute, now)
			if status != tt.wantStatus || attempts != tt.wantAttempts {
				t.Fatalf("FailureSchedule() = %q/%d, want %q/%d", status, attempts, tt.wantStatus, tt.wantAttempts)
			}
			if tt.wantStatus == ReconciliationStatusDead {
				if !next.Equal(now) {
					t.Errorf("terminal next attempt = %v, want %v", next, now)
				}
				return
			}
			minNext := now.Add(tt.wantNextDelay)
			maxNext := minNext.Add(tt.wantNextDelay / 4)
			if next.Before(minNext) || next.After(maxNext) {
				t.Errorf("next attempt = %v, want within [%v, %v]", next, minNext, maxNext)
			}
		})
	}
}
