// Package stamp writes a metadata key into Vault secret paths to record
// when and by whom a secret was last modified by vaultpatch.
package stamp

import (
	"context"
	"fmt"
	"time"
)

const DefaultKey = "_vaultpatch_stamped_at"

// Writer can read and write a map of key/value pairs at a Vault path.
type Writer interface {
	Read(ctx context.Context, path string) (map[string]string, error)
	Write(ctx map[string]string) error
}

// Options controls stamping behaviour.
type Options struct {
	Key    string
	Actor  string
	DryRun bool
}

// Result records the outcome for a single path.
type Result struct {
	Path  string
	Error error
}

// Stamper applies a timestamp metadata key to one or more secret paths.
type Stamper struct {
	w   Writer
	opt Options
}

// New creates a Stamper. If opt.Key is empty, DefaultKey is used.
func New(w Writer, opt Options) *Stamper {
	if opt.Key == "" {
		opt.Key = DefaultKey
	}
	return &Stamper{w: w, opt: opt}
}

// Apply stamps each path with the current UTC timestamp.
func (s *Stamper) Apply(ctx context.Context, paths []string) []Result {
	ts := time.Now().UTC().Format(time.RFC3339)
	results := make([]Result, 0, len(paths))

	for _, p := range paths {
		results = append(results, s.stamp(ctx, p, ts))
	}
	return results
}

func (s *Stamper) stamp(ctx context.Context, path, ts string) Result {
	data, err := s.w.Read(ctx, path)
	if err != nil {
		return Result{Path: path, Error: fmt.Errorf("read: %w", err)}
	}

	if data == nil {
		data = make(map[string]string)
	}

	value := ts
	if s.opt.Actor != "" {
		value = ts + " by " + s.opt.Actor
	}
	data[s.opt.Key] = value

	if s.opt.DryRun {
		return Result{Path: path}
	}

	if err := s.w.Write(ctx, path, data); err != nil {
		return Result{Path: path, Error: fmt.Errorf("write: %w", err)}
	}
	return Result{Path: path}
}
