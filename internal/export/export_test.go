package export_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/export"
)

func TestNewExporter_ValidFormats(t *testing.T) {
	formats := []export.Format{export.FormatJSON, export.FormatYAML, export.FormatDotenv}
	for _, f := range formats {
		_, err := export.NewExporter(f, &bytes.Buffer{})
		if err != nil {
			t.Errorf("expected no error for format %q, got: %v", f, err)
		}
	}
}

func TestNewExporter_InvalidFormat(t *testing.T) {
	_, err := export.NewExporter("xml", &bytes.Buffer{})
	if err == nil {
		t.Fatal("expected error for unsupported format, got nil")
	}
}

func TestWrite_JSON(t *testing.T) {
	var buf bytes.Buffer
	e, _ := export.NewExporter(export.FormatJSON, &buf)
	secrets := map[string]string{"KEY": "value"}

	if err := e.Write(secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var out map[string]string
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if out["KEY"] != "value" {
		t.Errorf("expected KEY=value, got %q", out["KEY"])
	}
}

func TestWrite_Dotenv(t *testing.T) {
	var buf bytes.Buffer
	e, _ := export.NewExporter(export.FormatDotenv, &buf)
	secrets := map[string]string{
		"DB_PASS": "s3cr3t",
		"API_KEY": "abc123",
	}

	if err := e.Write(secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `API_KEY="abc123"`) {
		t.Errorf("expected API_KEY line, got:\n%s", out)
	}
	if !strings.Contains(out, `DB_PASS="s3cr3t"`) {
		t.Errorf("expected DB_PASS line, got:\n%s", out)
	}
}

func TestWrite_Dotenv_EscapesQuotes(t *testing.T) {
	var buf bytes.Buffer
	e, _ := export.NewExporter(export.FormatDotenv, &buf)
	secrets := map[string]string{"VAL": `say "hello"`}

	_ = e.Write(secrets)

	if !strings.Contains(buf.String(), `\"hello\"`) {
		t.Errorf("expected escaped quotes in output, got: %s", buf.String())
	}
}

func TestWrite_YAML(t *testing.T) {
	var buf bytes.Buffer
	e, _ := export.NewExporter(export.FormatYAML, &buf)
	secrets := map[string]string{"TOKEN": "xyz"}

	if err := e.Write(secrets); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "TOKEN") {
		t.Errorf("expected TOKEN in YAML output, got: %s", buf.String())
	}
}
