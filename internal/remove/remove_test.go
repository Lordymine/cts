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

// recordingRunner records commands without executing them.
type recordingRunner struct{ calls [][]string }

func (r *recordingRunner) Run(_ context.Context, name string, args ...string) error {
	r.calls = append(r.calls, append([]string{name}, args...))
	return nil
}

func TestDryRunDoesNotDelete(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, "skill")
	writeFile(t, filepath.Join(skill, "SKILL.md"), "x")

	tg := target.Target{Name: "skill", Category: target.Skill, Paths: []string{skill}, SizeBytes: 1}
	res, err := New(filepath.Join(dir, "backup"), true, &recordingRunner{}).Remove(context.Background(), []target.Target{tg})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if !res.DryRun {
		t.Error("Result.DryRun should be true")
	}
	if _, err := os.Stat(skill); err != nil {
		t.Errorf("dry-run must not delete anything, but it's gone: %v", err)
	}
	if len(res.Removed) != 1 || res.FreedBytes != 1 {
		t.Errorf("Removed=%d FreedBytes=%d, want 1 and 1", len(res.Removed), res.FreedBytes)
	}
}

func TestExecuteDeletesAndBacksUp(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, "skill")
	writeFile(t, filepath.Join(skill, "SKILL.md"), "content")
	backup := filepath.Join(dir, "backup")

	tg := target.Target{Name: "skill", Category: target.Skill, Paths: []string{skill}}
	_, err := New(backup, false, &recordingRunner{}).Remove(context.Background(), []target.Target{tg})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}

	if _, err := os.Stat(skill); !os.IsNotExist(err) {
		t.Errorf("execution should have deleted %s (err=%v)", skill, err)
	}
	if !backupHasFile(backup, "SKILL.md") {
		t.Error("backup should contain the deleted file")
	}
}

func TestBackupFailureDoesNotDelete(t *testing.T) {
	dir := t.TempDir()
	skill := filepath.Join(dir, "skill")
	writeFile(t, filepath.Join(skill, "SKILL.md"), "x")

	badBackup := filepath.Join(dir, "file")
	writeFile(t, badBackup, "occupies the name") // backupDir is a file → MkdirAll fails

	tg := target.Target{Name: "skill", Category: target.Skill, Paths: []string{skill}}
	_, err := New(badBackup, false, &recordingRunner{}).Remove(context.Background(), []target.Target{tg})
	if err == nil {
		t.Fatal("a failed backup should return an error")
	}
	if _, statErr := os.Stat(skill); statErr != nil {
		t.Error("if the backup fails, the target must NOT be deleted")
	}
}

func TestNoPathsNoUninstallIsSkipped(t *testing.T) {
	tg := target.Target{Name: "empty", Category: target.MCP} // nothing to do
	res, err := New(t.TempDir(), false, &recordingRunner{}).Remove(context.Background(), []target.Target{tg})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(res.Removed) != 0 {
		t.Errorf("a target with no Paths and no Uninstall should be skipped, got Removed=%d", len(res.Removed))
	}
}

func TestUninstallRunsCommandThenDeletes(t *testing.T) {
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
		t.Fatalf("want 1 command run, got %d", len(rr.calls))
	}
	if got, want := strings.Join(rr.calls[0], " "), "npm rm -g @qwen-code/qwen-code"; got != want {
		t.Errorf("command = %q, want %q", got, want)
	}
	if _, err := os.Stat(cfg); !os.IsNotExist(err) {
		t.Error("config should have been deleted after the uninstall")
	}
}

func TestDryRunDoesNotRunCommand(t *testing.T) {
	rr := &recordingRunner{}
	tg := target.Target{Name: "x", Category: target.Agent, Uninstall: []string{"npm", "rm", "-g", "x"}}
	res, err := New(t.TempDir(), true, rr).Remove(context.Background(), []target.Target{tg})
	if err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(rr.calls) != 0 {
		t.Error("dry-run must not run any command")
	}
	if len(res.Removed) != 1 {
		t.Error("dry-run should list the target as 'would remove'")
	}
}

func backupHasFile(root, name string) bool {
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
