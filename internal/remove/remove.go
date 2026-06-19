// Package remove deletes targets safely: dry-run by default and a backup before
// deleting. If the backup fails, nothing is destroyed. When a target has an
// uninstall command, it runs before the files are deleted.
package remove

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"cts/internal/target"
)

// Result summarizes what the removal did (or would do, in dry-run).
type Result struct {
	Removed    []target.Target
	FreedBytes int64
	DryRun     bool
}

// Runner runs an external command (e.g. npm rm -g, claude mcp remove). Injected
// so tests run without executing anything for real.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) error
}

// Remover deletes targets. backupDir is where content is copied before it's gone.
type Remover struct {
	backupDir string
	dryRun    bool
	runner    Runner
}

// New creates a Remover. dryRun=true touches nothing — it only reports what it would do.
func New(backupDir string, dryRun bool, runner Runner) Remover {
	return Remover{backupDir: backupDir, dryRun: dryRun, runner: runner}
}

// Remove processes the targets. A target with no Paths and no Uninstall is skipped.
func (r Remover) Remove(ctx context.Context, targets []target.Target) (Result, error) {
	res := Result{DryRun: r.dryRun}
	for _, t := range targets {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		if len(t.Paths) == 0 && len(t.Uninstall) == 0 {
			continue // nothing to do with this target
		}
		if !r.dryRun {
			if err := r.removeOne(ctx, t); err != nil {
				return res, fmt.Errorf("remove %s: %w", t.Name, err)
			}
		}
		res.Removed = append(res.Removed, t)
		res.FreedBytes += t.SizeBytes
	}
	return res, nil
}

// removeOne: backup first (if it fails, nothing is destroyed); then uninstall the
// package (if any); finally delete the files.
func (r Remover) removeOne(ctx context.Context, t target.Target) error {
	for _, p := range t.Paths {
		if err := r.backup(t, p); err != nil {
			return fmt.Errorf("backup %s: %w", p, err)
		}
	}
	if len(t.Uninstall) > 0 {
		if err := r.runner.Run(ctx, t.Uninstall[0], t.Uninstall[1:]...); err != nil {
			return fmt.Errorf("uninstall (%v): %w", t.Uninstall, err)
		}
	}
	for _, p := range t.Paths {
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("delete %s: %w", p, err)
		}
	}
	return nil
}

// backup copies p into backupDir, preserving category/name. A symlink has no
// substantial content to keep — it is just removed.
func (r Remover) backup(t target.Target, p string) error {
	info, err := os.Lstat(p)
	if err != nil {
		return err
	}
	if info.Mode()&fs.ModeSymlink != 0 {
		return nil
	}

	dest := filepath.Join(r.backupDir, string(t.Category), t.Name, filepath.Base(p))
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	if info.IsDir() {
		return os.CopyFS(dest, os.DirFS(p))
	}
	return copyFile(p, dest)
}

func copyFile(src, dest string) error {
	b, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, b, 0o644)
}
