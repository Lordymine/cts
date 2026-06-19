// Package dirsize sums the size of the files under a directory.
package dirsize

import (
	"io/fs"
	"path/filepath"
)

// Of sums the size (bytes) of the files under path. Best-effort: an error on a
// single path does not break the count. It resolves the root symlink before
// walking (a symlinked skill would otherwise measure 0).
func Of(path string) int64 {
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	var total int64
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}
