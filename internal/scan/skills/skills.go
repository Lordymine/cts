// Package skills acha skills instaladas e marca as quebradas/incompletas.
package skills

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"cts/internal/target"
)

// Scanner varre um diretório de skills (ex.: ~/.claude/skills).
type Scanner struct {
	root string
}

// New cria um Scanner para o diretório raiz de skills.
func New(root string) Scanner {
	return Scanner{root: root}
}

// Category satisfaz scan.Scanner.
func (s Scanner) Category() target.Category { return target.Skill }

// Scan lista cada skill no root. Marca como morta a que tem symlink
// quebrado ou não tem SKILL.md (incompleta).
func (s Scanner) Scan(ctx context.Context) ([]target.Target, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // sem dir de skills não é erro — só não há o que limpar
		}
		return nil, fmt.Errorf("ler %s: %w", s.root, err)
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

// inspect classifica uma entrada. Esconde a regra de "está morta?" atrás de
// uma chamada simples — o chamador não precisa saber como ela decide.
func (s Scanner) inspect(path string, e fs.DirEntry) target.Target {
	t := target.Target{Name: e.Name(), Category: target.Skill, Paths: []string{path}}

	// Symlink quebrado: Lstat (via DirEntry) vê o link, Stat falha porque o alvo sumiu.
	if e.Type()&fs.ModeSymlink != 0 {
		if _, err := os.Stat(path); err != nil {
			t.Dead, t.Reason = true, "symlink quebrado"
			return t
		}
	}

	if _, err := os.Stat(filepath.Join(path, "SKILL.md")); err != nil {
		t.Dead, t.Reason = true, "sem SKILL.md"
		return t
	}

	t.SizeBytes = dirSize(path)
	return t
}

// dirSize soma o tamanho dos arquivos sob path. Best-effort: erro num caminho
// isolado não derruba a contagem inteira.
func dirSize(path string) int64 {
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
