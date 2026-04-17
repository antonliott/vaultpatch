package truncate

import "fmt"

// Options controls truncation behaviour.
type Options struct {
	MaxLen  int
	Suffix  string
	DryRun  bool
}

// Result holds the outcome of a truncation run.
type Result struct {
	Path     string
	Key      string
	Original string
	Truncated string
	DryRun   bool
}

func (r Result) String() string {
	if r.DryRun {
		return fmt.Sprintf("[dry-run] %s %s: %q -> %q", r.Path, r.Key, r.Original, r.Truncated)
	}
	return fmt.Sprintf("%s %s: %q -> %q", r.Path, r.Key, r.Original, r.Truncated)
}

//go:generate mockgen -destination=mock_test.go -package=truncate_test . ReadWriter

// ReadWriter is the Vault KV interface required by Truncator.
type ReadWriter interface {
	List(path string) ([]string, error)
	Read(path string) (map[string]string, error)
	Write(path string, data map[string]string) error
}

// Truncator shortens secret values that exceed MaxLen.
type Truncator struct {
	rw   ReadWriter
	opts Options
}

// NewTruncator returns a configured Truncator.
func NewTruncator(rw ReadWriter, opts Options) *Truncator {
	if opts.Suffix == "" {
		opts.Suffix = "..."
	}
	return &Truncator{rw: rw, opts: opts}
}

// Apply lists secrets under path and truncates any values longer than MaxLen.
func (t *Truncator) Apply(path string) ([]Result, error) {
	keys, err := t.rw.List(path)
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", path, err)
	}

	var results []Result
	for _, k := range keys {
		full := path + "/" + k
		secrets, err := t.rw.Read(full)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", full, err)
		}

		updated := make(map[string]string, len(secrets))
		changed := false
		for sk, sv := range secrets {
			if len(sv) > t.opts.MaxLen {
				truncVal := sv[:t.opts.MaxLen] + t.opts.Suffix
				results = append(results, Result{
					Path:      full,
					Key:       sk,
					Original:  sv,
					Truncated: truncVal,
					DryRun:    t.opts.DryRun,
				})
				updated[sk] = truncVal
				changed = true
			} else {
				updated[sk] = sv
			}
		}

		if changed && !t.opts.DryRun {
			if err := t.rw.Write(full, updated); err != nil {
				return nil, fmt.Errorf("write %s: %w", full, err)
			}
		}
	}
	return results, nil
}
