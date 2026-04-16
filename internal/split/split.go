package split

import (
	"fmt"
	"strings"
)

// Reader reads secrets from a path.
type Reader interface {
	List(path string) ([]string, error)
	Read(path string) (map[string]string, error)
}

// Writer writes secrets to a path.
type Writer interface {
	Write(path string, data map[string]string) error
}

// Result holds the outcome of a split operation.
type Result struct {
	Written []string
	Skipped []string
	DryRun  bool
}

// Splitter splits a single secret path into multiple paths keyed by a field.
type Splitter struct {
	rw       interface{ Reader
		Writer }
	keyField string
	destDir  string
	dryRun   bool
}

// NewSplitter creates a Splitter that partitions secrets by keyField into destDir.
func NewSplitter(rw interface {
	Reader
	Writer
}, keyField, destDir string, dryRun bool) *Splitter {
	return &Splitter{rw: rw, keyField: keyField, destDir: destDir, dryRun: dryRun}
}

// Apply reads src, groups remaining keys under destDir/<keyField-value>, and writes each group.
func (s *Splitter) Apply(src string) (Result, error) {
	data, err := s.rw.Read(src)
	if err != nil {
		return Result{}, fmt.Errorf("split: read %s: %w", src, err)
	}

	groupVal, ok := data[s.keyField]
	if !ok || strings.TrimSpace(groupVal) == "" {
		return Result{Skipped: []string{src}, DryRun: s.dryRun}, nil
	}

	destPath := strings.TrimRight(s.destDir, "/") + "/" + groupVal
	payload := make(map[string]string, len(data)-1)
	for k, v := range data {
		if k != s.keyField {
			payload[k] = v
		}
	}

	if s.dryRun {
		return Result{Written: []string{destPath}, DryRun: true}, nil
	}

	if err := s.rw.Write(destPath, payload); err != nil {
		return Result{}, fmt.Errorf("split: write %s: %w", destPath, err)
	}
	return Result{Written: []string{destPath}, DryRun: false}, nil
}
