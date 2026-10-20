package adminworker

import (
	"encoding/json"
	"testing"

	"github.com/AppeiYA/requisition-system/internal/admin/tasks"
)

func Test_InvitationEmailTask_RoundTrip(t *testing.T) {
	t.Parallel()

	payload := tasks.InvitationEmailPayload{
		To:        "invitee@example.com",
		InvitedBy: "usr-1234",
		Token:     "raw-invitation-token",
		ExpiresAt: "2026-01-01T00:00:00Z",
	}

	task, err := invitationEmailTask(payload)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}

	if task.Type() != tasks.SendInvitationEmail {
		t.Errorf("expected task type %q, got %q", tasks.SendInvitationEmail, task.Type())
	}

	var decoded tasks.InvitationEmailPayload
	if err := json.Unmarshal(task.Payload(), &decoded); err != nil {
		t.Fatalf("failed to decode payload: %v", err)
	}
	if decoded != payload {
		t.Errorf("payload did not round-trip: got %+v, want %+v", decoded, payload)
	}
}
