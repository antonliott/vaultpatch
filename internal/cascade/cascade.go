// Package cascade propagates secret values from a parent path to one or more
// child paths, writing only keys that do not already exist in the child
// (unless overwrite is enabled).
package cascade

import (
	"context"
	"fmt"
)

// Reader reads secrets from a Vault path.
type Reader interface {
	Read(ctx context.Context, path string) (map[string]string, error)
}

// Writer writes secrets to a Vault path.
type Writer interface {
	Write(ctx context.Context, path string, data map[string]string) error
}

// ReadWriter combines Reader and Writer.
type ReadWriter interface {
	Reader
	Writer
}

// Options controls cascade behaviour.
type Options struct {
	DryRun    bool
	Overwrite bool
}

// Result summarises what happened during a cascade operation.
type Result struct {
	Path    string
	Written int
	Skipped int
	Err     error
}

// Cascader propagates secrets from a parent path to child paths.
type Cascader struct {
	rw   ReadWriter
	opts Options
}

// NewCascader returns a new Cascader.
func NewCascader(rw ReadWriter, opts Options) *Cascader {
	return &Cascader{rw: rw, opts: opts}
}

// Apply reads the parent path and propagates its key/value pairs to each
// child path. Keys already present in a child are skipped unless Overwrite
// is set.
func (c *Cascader) Apply(ctx context.Context, parent string, children []string) []Result {
	results := make([]Result, 0, len(children))

	parentData, err := c.rw.Read(ctx, parent)
	if err != nil {
		for _, ch := range children {
			results = append(results, Result{Path: ch, Err: fmt.Errorf("read parent: %w", err)})
		}
		return results
	}

	for _, child := range children {
		res := c.propagate(ctx, parentData, child)
		results = append(results, res)
	}
	return results
}

func (c *Cascader) propagate(ctx context.Context, parentData map[string]string, child string) Result {
	res := Result{Path: child}

	childData, err := c.rw.Read(ctx, child)
	if err != nil {
		res.Err = fmt.Errorf("read child %s: %w", child, err)
		return res
	}

	merged := make(map[string]string, len(childData))
	for k, v := range childData {
		merged[k] = v
	}

	for k, v := range parentData {
		if _, exists := merged[k]; exists && !c.opts.Overwrite {
			res.Skipped++
			continue
		}
		merged[k] = v
		res.Written++
	}

	if res.Written == 0 {
		return res
	}

	if !c.opts.DryRun {
		if err := c.rw.Write(ctx, child, merged); err != nil {
			res.Err = fmt.Errorf("write child %s: %w", child, err)
		}
	}
	return res
}
