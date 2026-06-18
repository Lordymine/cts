// Package remove apaga targets com segurança: dry-run por padrão e backup
// antes de apagar. Se o backup falha, o alvo NÃO é apagado.
package remove

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"cts/internal/target"
)

// Result resume o que a remoção fez (ou faria, em dry-run).
type Result struct {
	Removed    []target.Target
	FreedBytes int64
	DryRun     bool
}

// Remover apaga targets. backupDir é onde o conteúdo é copiado antes de sumir.
type Remover struct {
	backupDir string
	dryRun    bool
}

// New cria um Remover. dryRun=true não toca em nada — só reporta o que faria.
func New(backupDir string, dryRun bool) Remover {
	return Remover{backupDir: backupDir, dryRun: dryRun}
}

// Remove processa os targets. Em dry-run, nada é tocado. Em execução, faz backup
// das Paths antes de apagar; erro em um alvo aborta antes de apagá-lo.
func (r Remover) Remove(ctx context.Context, targets []target.Target) (Result, error) {
	res := Result{DryRun: r.dryRun}
	for _, t := range targets {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		if !r.dryRun {
			if err := r.removeOne(t); err != nil {
				return res, fmt.Errorf("remover %s: %w", t.Name, err)
			}
		}
		res.Removed = append(res.Removed, t)
		res.FreedBytes += t.SizeBytes
	}
	return res, nil
}

// removeOne faz backup e apaga cada path do alvo. Backup primeiro: se ele falha,
// retorna antes de apagar — nada é perdido sem cópia.
func (r Remover) removeOne(t target.Target) error {
	for _, p := range t.Paths {
		if err := r.backup(t, p); err != nil {
			return fmt.Errorf("backup %s: %w", p, err)
		}
		if err := os.RemoveAll(p); err != nil {
			return fmt.Errorf("apagar %s: %w", p, err)
		}
	}
	return nil
}

// backup copia p para dentro de backupDir, preservando categoria/nome. Symlink
// não tem conteúdo substancial a guardar — só será removido.
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
