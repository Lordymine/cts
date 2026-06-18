package remove

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"cts/internal/target"
)

func TestDryRunNaoApaga(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, "skill")
	writeFile(t, filepath.Join(skill, "SKILL.md"), "x")

	tg := target.Target{Name: "skill", Category: target.Skill, Paths: []string{skill}, SizeBytes: 1}
	res, err := New(filepath.Join(dir, "backup"), true).Remove(context.Background(), []target.Target{tg})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if !res.DryRun {
		t.Error("Result.DryRun deveria ser true")
	}
	if _, err := os.Stat(skill); err != nil {
		t.Errorf("dry-run não pode apagar nada; mas sumiu: %v", err)
	}
	if len(res.Removed) != 1 || res.FreedBytes != 1 {
		t.Errorf("Removed=%d FreedBytes=%d, queria 1 e 1", len(res.Removed), res.FreedBytes)
	}
}

func TestExecutaApagaEFazBackup(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, "skill")
	writeFile(t, filepath.Join(skill, "SKILL.md"), "conteudo")
	backup := filepath.Join(dir, "backup")

	tg := target.Target{Name: "skill", Category: target.Skill, Paths: []string{skill}}
	_, err := New(backup, false).Remove(context.Background(), []target.Target{tg})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if _, err := os.Stat(skill); !os.IsNotExist(err) {
		t.Errorf("execução deveria ter apagado %s (err=%v)", skill, err)
	}
	if !backupContemArquivo(backup, "SKILL.md") {
		t.Error("backup deveria conter o arquivo apagado")
	}
}

func TestBackupFalhandoNaoApaga(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, "skill")
	writeFile(t, filepath.Join(skill, "SKILL.md"), "x")

	// backupDir aponta para um ARQUIVO existente → MkdirAll do backup falha → não pode apagar
	badBackup := filepath.Join(dir, "arquivo")
	writeFile(t, badBackup, "ocupa o nome")

	tg := target.Target{Name: "skill", Category: target.Skill, Paths: []string{skill}}
	_, err := New(badBackup, false).Remove(context.Background(), []target.Target{tg})
	if err == nil {
		t.Fatal("backup falho deveria devolver erro")
	}
	if _, statErr := os.Stat(skill); statErr != nil {
		t.Error("se o backup falha, o alvo NÃO pode ser apagado")
	}
}

func TestSemPathsEhPulado(t *testing.T) {
	tg := target.Target{Name: "mcp-x", Category: target.MCP} // sem Paths (remoção é config-edit)
	res, err := New(t.TempDir(), false).Remove(context.Background(), []target.Target{tg})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(res.Removed) != 0 || res.FreedBytes != 0 {
		t.Errorf("alvo sem Paths não pode ser removido pelo file-remover; veio Removed=%d Freed=%d", len(res.Removed), res.FreedBytes)
	}
}

func backupContemArquivo(root, name string) bool {
	found := false
	_ = filepath.WalkDir(root, func(_ string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && d.Name() == name {
			found = true
		}
		return nil
	})
	return found
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
