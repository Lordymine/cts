package remove

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cts/internal/target"
)

// recordingRunner registra os comandos sem executá-los.
type recordingRunner struct{ calls [][]string }

func (r *recordingRunner) Run(_ context.Context, name string, args ...string) error {
	r.calls = append(r.calls, append([]string{name}, args...))
	return nil
}

func TestDryRunNaoApaga(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, "skill")
	writeFile(t, filepath.Join(skill, "SKILL.md"), "x")

	tg := target.Target{Name: "skill", Category: target.Skill, Paths: []string{skill}, SizeBytes: 1}
	res, err := New(filepath.Join(dir, "backup"), true, &recordingRunner{}).Remove(context.Background(), []target.Target{tg})
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
	_, err := New(backup, false, &recordingRunner{}).Remove(context.Background(), []target.Target{tg})
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

	badBackup := filepath.Join(dir, "arquivo")
	writeFile(t, badBackup, "ocupa o nome") // backupDir é um arquivo → MkdirAll falha

	tg := target.Target{Name: "skill", Category: target.Skill, Paths: []string{skill}}
	_, err := New(badBackup, false, &recordingRunner{}).Remove(context.Background(), []target.Target{tg})
	if err == nil {
		t.Fatal("backup falho deveria devolver erro")
	}
	if _, statErr := os.Stat(skill); statErr != nil {
		t.Error("se o backup falha, o alvo NÃO pode ser apagado")
	}
}

func TestSemPathsNemUninstallEhPulado(t *testing.T) {
	tg := target.Target{Name: "vazio", Category: target.MCP} // nada a fazer
	res, err := New(t.TempDir(), false, &recordingRunner{}).Remove(context.Background(), []target.Target{tg})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(res.Removed) != 0 {
		t.Errorf("alvo sem Paths nem Uninstall deveria ser pulado, veio Removed=%d", len(res.Removed))
	}
}

func TestUninstallRodaComandoEntaoApaga(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "cfg")
	writeFile(t, filepath.Join(cfg, "f"), "x")
	rr := &recordingRunner{}

	tg := target.Target{
		Name:      "qwen",
		Category:  target.Agent,
		Paths:     []string{cfg},
		Uninstall: []string{"npm", "rm", "-g", "@qwen-code/qwen-code"},
	}
	_, err := New(filepath.Join(dir, "backup"), false, rr).Remove(context.Background(), []target.Target{tg})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if len(rr.calls) != 1 {
		t.Fatalf("queria 1 comando rodado, veio %d", len(rr.calls))
	}
	if got, want := strings.Join(rr.calls[0], " "), "npm rm -g @qwen-code/qwen-code"; got != want {
		t.Errorf("comando = %q, queria %q", got, want)
	}
	if _, err := os.Stat(cfg); !os.IsNotExist(err) {
		t.Error("config deveria ter sido apagada após o uninstall")
	}
}

func TestDryRunNaoRodaComando(t *testing.T) {
	rr := &recordingRunner{}
	tg := target.Target{Name: "x", Category: target.Agent, Uninstall: []string{"npm", "rm", "-g", "x"}}
	res, err := New(t.TempDir(), true, rr).Remove(context.Background(), []target.Target{tg})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(rr.calls) != 0 {
		t.Error("dry-run não pode rodar comando")
	}
	if len(res.Removed) != 1 {
		t.Error("dry-run deveria listar o alvo como 'removeria'")
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
