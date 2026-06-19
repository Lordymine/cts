// Package skills finds installed skills and flags the broken/incomplete ones.
package skills

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Lordymine/cts/internal/dirsize"
	"github.com/Lordymine/cts/internal/target"
)

// Scanner sweeps a skills directory (e.g. ~/.claude/skills).
type Scanner struct {
	root string
}

// New creates a Scanner for the skills root directory.
func New(root string) Scanner {
	return Scanner{root: root}
}

// Category satisfies scan.Scanner.
func (s Scanner) Category() target.Category { return target.Skill }

// Scan lists each skill under root. It marks as dead any skill with a broken
// symlink or no SKILL.md (incomplete).
func (s Scanner) Scan(ctx context.Context) ([]target.Target, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // a missing skills dir is not an error — there's just nothing to clean
		}
		return nil, fmt.Errorf("read %s: %w", s.root, err)
	}

	var targets []target.Target
	for _, e := range entries {
		if err := ctx.Err(); err != nil {
			return targets, err
		}
		targets = append(targets, s.inspect(filepath.Join(s.root, e.Name()), e))
	}
	return targets, nil
}

// inspect classifies an entry. It hides the "is it dead?" rule behind a simple
// call — the caller doesn't need to know how it decides.
func (s Scanner) inspect(path string, e fs.DirEntry) target.Target {
	t := target.Target{Name: e.Name(), Category: target.Skill, Paths: []string{path}}

	// Broken symlink: Lstat (via DirEntry) sees the link, Stat fails because the target is gone.
	if e.Type()&fs.ModeSymlink != 0 {
		if _, err := os.Stat(path); err != nil {
			t.Dead, t.Reason = true, "broken symlink"
			return t
		}
	}

	if _, err := os.Stat(filepath.Join(path, "SKILL.md")); err != nil {
		t.Dead, t.Reason = true, "no SKILL.md"
		return t
	}

	t.SizeBytes = dirsize.Of(path)
	return t
}
