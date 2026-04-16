package squash

import "fmt"

// Writer writes secrets to a path.
type Writer interface {
	Write(path string, data map[string]string) error
}

// Reader reads secrets from a path.
type Reader interface {
	Read(path string) (map[string]string, error)
}

// ReadWriter can read and write secrets.
type ReadWriter interface {
	Reader
	Writer
}

// Result holds the outcome of a squash operation.
type Result struct {
	Destination string
	Merged      int
	DryRun      bool
	Err         error
}

func (r Result) String() string {
	if r.DryRun {
		return fmt.Sprintf("[dry-run] would squash %d paths into %s", r.Merged, r.Destination)
	}
	if r.Err != nil {
		return fmt.Sprintf("squash failed: %v", r.Err)
	}
	return fmt.Sprintf("squashed %d paths into %s", r.Merged, r.Destination)
}

// Squasher merges multiple secret paths into one destination path.
type Squasher struct {
	rw     ReadWriter
	dryRun bool
}

// New returns a new Squasher.
func New(rw ReadWriter, dryRun bool) *Squasher {
	return &Squasher{rw: rw, dryRun: dryRun}
}

// Apply reads all sources and writes merged secrets to dest.
// Later sources overwrite earlier ones on key conflict.
func (s *Squasher) Apply(dest string, sources []string) Result {
	merged := make(map[string]string)
	for _, src := range sources {
		data, err := s.rw.Read(src)
		if err != nil {
			return Result{Destination: dest, Err: fmt.Errorf("read %s: %w", src, err)}
		}
		for k, v := range data {
			merged[k] = v
		}
	}
	if s.dryRun {
		return Result{Destination: dest, Merged: len(sources), DryRun: true}
	}
	if err := s.rw.Write(dest, merged); err != nil {
		return Result{Destination: dest, Err: fmt.Errorf("write %s: %w", dest, err)}
	}
	return Result{Destination: dest, Merged: len(sources)}
}
