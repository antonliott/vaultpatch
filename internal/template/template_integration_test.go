package template_test

import (
	"strings"
	"testing"

	"github.com/your-org/vaultpatch/internal/template"
)

// Integration-style tests that exercise Renderer through its public API
// using realistic multi-key secret maps.

func TestRenderer_DatabaseDSN(t *testing.T) {
	r := template.NewRenderer()
	tmpl := `postgres://{{index . "db_user"}}:{{index . "db_pass"}}@{{default "db_host" "localhost" .}}:{{default "db_port" "5432" .}}/{{required "db_name" .}}`
	data := map[string]string{
		"db_user": "admin",
		"db_pass": "hunter2",
		"db_name": "myapp",
	}
	got, err := r.Render("dsn", tmpl, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "postgres://admin:hunter2@localhost:5432/myapp"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRenderer_MultipleTemplates(t *testing.T) {
	r := template.NewRenderer()
	templates := map[string]string{
		"app.env": `API_KEY={{index . "api_key"}}\nDEBUG={{default "debug" "false" .}}`,
		"db.env":  `DB_URL={{index . "db_url"}}`,
	}
	data := map[string]string{
		"api_key": "abc123",
		"db_url":  "postgres://localhost/prod",
	}
	out, err := r.RenderAll(templates, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out["app.env"], "abc123") {
		t.Errorf("app.env missing api_key: %q", out["app.env"])
	}
	if !strings.Contains(out["db.env"], "postgres://") {
		t.Errorf("db.env missing db_url: %q", out["db.env"])
	}
}

func TestRenderer_RequiredMissingInMulti(t *testing.T) {
	r := template.NewRenderer()
	templates := map[string]string{
		"ok":  `{{index . "k"}}`,
		"bad": `{{required "missing" .}}`,
	}
	_, err := r.RenderAll(templates, map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected error for missing required key in RenderAll")
	}
}
