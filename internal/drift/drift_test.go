package drift_test

import (
	"context"
	"errors"
	"testing"

	"github.com/your-org/vaultpatch/internal/drift"
)

type staticReader struct {
	data map[string]string
	err  error
}

func (s *staticReader) Read(_ context.Context, _ string) (map[string]string, error) {
	return s.data, s.err
}

func TestDetect_NoDrift(t *testing.T) {
	r := &staticReader{data: map[string]string{"key": "val"}}
	d := drift.NewDetector(r)

	report, err := d.Detect(context.Background(), "secret/app", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if report.HasDrift() {
		t.Errorf("expected no drift, got %d delta(s)", len(report.Deltas))
	}
}

func TestDetect_Changed(t *testing.T) {
	r := &staticReader{data: map[string]string{"key": "new"}}
	d := drift.NewDetector(r)

	report, err := d.Detect(context.Background(), "secret/app", map[string]string{"key": "old"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Deltas) != 1 || report.Deltas[0].Status != "changed" {
		t.Errorf("expected 1 changed delta, got %+v", report.Deltas)
	}
}

func TestDetect_Added(t *testing.T) {
	r := &staticReader{data: map[string]string{"key": "val", "extra": "new"}}
	d := drift.NewDetector(r)

	report, err := d.Detect(context.Background(), "secret/app", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Deltas) != 1 || report.Deltas[0].Status != "added" {
		t.Errorf("expected 1 added delta, got %+v", report.Deltas)
	}
}

func TestDetect_Removed(t *testing.T) {
	r := &staticReader{data: map[string]string{}}
	d := drift.NewDetector(r)

	report, err := d.Detect(context.Background(), "secret/app", map[string]string{"key": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Deltas) != 1 || report.Deltas[0].Status != "removed" {
		t.Errorf("expected 1 removed delta, got %+v", report.Deltas)
	}
}

func TestDetect_ReadError(t *testing.T) {
	r := &staticReader{err: errors.New("vault unavailable")}
	d := drift.NewDetector(r)

	_, err := d.Detect(context.Background(), "secret/app", map[string]string{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReport_Summary_NoDrift(t *testing.T) {
	rep := drift.Report{Path: "secret/app", Deltas: nil}
	got := rep.Summary()
	if got != "secret/app: no drift detected" {
		t.Errorf("unexpected summary: %q", got)
	}
}

func TestReport_Summary_WithDrift(t *testing.T) {
	rep := drift.Report{
		Path:   "secret/app",
		Deltas: []drift.Delta{{Key: "k", Status: "changed"}},
	}
	got := rep.Summary()
	if got != "secret/app: 1 drift delta(s) detected" {
		t.Errorf("unexpected summary: %q", got)
	}
}
