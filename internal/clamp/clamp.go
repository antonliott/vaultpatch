// Package clamp truncates secret values that exceed a maximum byte length.
package clamp

import "fmt"

// Options configures the Clamper.
type Options struct {
	MaxBytes int
	DryRun   bool
}

// Result holds the outcome of a clamp operation.
type Result struct {
	Path    string
	Key     string
	OrigLen int
	NewLen  int
	DryRun  bool
}

// Writer writes secrets to a path.
type Writer interface {
	Write(path string, data map[string]string) error
}

// Reader reads secrets from a path.
type Reader interface {
	Read(path string) (map[string]string, error)
}

// ReadWriter combines Reader and Writer.
type ReadWriter interface {
	Reader
	Writer
}

// Clamper truncates long secret values.
type Clamper struct {
	rw   ReadWriter
	opts Options
}

// New creates a new Clamper.
func New(rw ReadWriter, opts Options) *Clamper {
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = 256
	}
	return &Clamper{rw: rw, opts: opts}
}

// Apply reads secrets at path, clamps values exceeding MaxBytes, and writes
// them back unless DryRun is set. Returns the list of clamped keys.
func (c *Clamper) Apply(path string) ([]Result, error) {
	data, err := c.rw.Read(path)
	if err != nil {
		return nil, fmt.Errorf("clamp: read %s: %w", path, err)
	}

	var results []Result
	updated := make(map[string]string, len(data))
	for k, v := range data {
		if len(v) > c.opts.MaxBytes {
			results = append(results, Result{
				Path:    path,
				Key:     k,
				OrigLen: len(v),
				NewLen:  c.opts.MaxBytes,
				DryRun:  c.opts.DryRun,
			})
			updated[k] = v[:c.opts.MaxBytes]
		} else {
			updated[k] = v
		}
	}

	if len(results) > 0 && !c.opts.DryRun {
		if err := c.rw.Write(path, updated); err != nil {
			return nil, fmt.Errorf("clamp: write %s: %w", path, err)
		}
	}
	return results, nil
}
