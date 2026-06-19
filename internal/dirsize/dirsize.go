// Package dirsize soma o tamanho dos arquivos sob um diretório.
package dirsize

import (
	"io/fs"
	"path/filepath"
)

// Of soma o tamanho (bytes) dos arquivos sob path. Best-effort: erro num
// caminho isolado não derruba a contagem. Resolve o symlink raiz antes de medir
// (uma skill symlinkada mediria 0 sem isso).
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
