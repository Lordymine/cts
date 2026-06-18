// Package remove apaga targets com segurança: dry-run por padrão e backup antes
// de apagar. Se o backup falha, nada é destruído. Quando o alvo tem comando de
// desinstalação (Uninstall), ele roda antes de apagar os arquivos.
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

// Runner roda um comando externo (ex.: npm rm -g, claude mcp remove). Injetado
// para testar sem executar nada de verdade.
type Runner interface {
	Run(ctx context.Context, name string, args ...string) error
}

// Remover apaga targets. backupDir é onde o conteúdo é copiado antes de sumir.
type Remover struct {
	backupDir string
	dryRun    bool
	runner    Runner
}

// New cria um Remover. dryRun=true não toca em nada — só reporta o que faria.
func New(backupDir string, dryRun bool, runner Runner) Remover {
	return Remover{backupDir: backupDir, dryRun: dryRun, runner: runner}
}

// Remove processa os targets. Alvo sem Paths e sem Uninstall é pulado.
func (r Remover) Remove(ctx context.Context, targets []target.Target) (Result, error) {
	res := Result{DryRun: r.dryRun}
	for _, t := range targets {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		if len(t.Paths) == 0 && len(t.Uninstall) == 0 {
			continue // nada a fazer com este alvo
		}
		if !r.dryRun {
			if err := r.removeOne(ctx, t); err != nil {
				return res, fmt.Errorf("remover %s: %w", t.Name, err)
			}
		}
		res.Removed = append(res.Removed, t)
		res.FreedBytes += t.SizeBytes
	}
	return res, nil
}

// removeOne: backup primeiro (se falha, nada é destruído); depois desinstala o
// pacote (se houver); por fim apaga os arquivos.
func (r Remover) removeOne(ctx context.Context, t target.Target) error {
	for _, p := range t.Paths {
		if err := r.backup(t, p); err != nil {
			return fmt.Errorf("backup %s: %w", p, err)
		}
	}
	if len(t.Uninstall) > 0 {
		if err := r.runner.Run(ctx, t.Uninstall[0], t.Uninstall[1:]...); err != nil {
			return fmt.Errorf("desinstalar (%v): %w", t.Uninstall, err)
		}
	}
	for _, p := range t.Paths {
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
