package audit_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/example/vaultpatch/internal/audit"
)

func TestLog_WritesJSONLine(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf)

	ev := audit.Event{
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Type:      audit.EventApply,
		Namespace: "prod",
		Path:      "secret/db",
		Keys:      []string{"password"},
		Success:   true,
	}

	if err := l.Log(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var got audit.Event
	if err := json.Unmarshal([]byte(line), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Namespace != "prod" {
		t.Errorf("namespace: got %q, want %q", got.Namespace, "prod")
	}
	if got.Type != audit.EventApply {
		t.Errorf("type: got %q, want %q", got.Type, audit.EventApply)
	}
}

func TestLog_SetsTimestampWhenZero(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf)

	ev := audit.Event{Type: audit.EventDiff, Namespace: "dev", Path: "secret/api"}
	if err := l.Log(ev); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got audit.Event
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Timestamp.IsZero() {
		t.Error("expected timestamp to be set automatically")
	}
}

func TestLogApply_DryRun(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf)

	if err := l.LogApply("staging", "secret/svc", []string{"token"}, true, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got audit.Event
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Type != audit.EventDryRun {
		t.Errorf("type: got %q, want %q", got.Type, audit.EventDryRun)
	}
	if !got.Success {
		t.Error("expected success=true for nil error")
	}
}

func TestLogApply_WithError(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf)

	applyErr := errors.New("permission denied")
	if err := l.LogApply("prod", "secret/db", nil, false, applyErr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got audit.Event
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &got); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if got.Success {
		t.Error("expected success=false when error is provided")
	}
	if got.Message != "permission denied" {
		t.Errorf("message: got %q, want %q", got.Message, "permission denied")
	}
}
