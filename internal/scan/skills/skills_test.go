package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Lordymine/cts/internal/target"
)

func TestScanFlagsSkillWithoutSkillMdAsDead(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "healthy", "SKILL.md"), "# ok") // healthy
	mkdir(t, filepath.Join(root, "broken"))                          // no SKILL.md → dead

	got, err := New(root).Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	wantDead := map[string]bool{"healthy": false, "broken": true}
	if len(got) != len(wantDead) {
		t.Fatalf("found %d skills, want %d", len(got), len(wantDead))
	}
	for _, tg := range got {
		if tg.Category != target.Skill {
			t.Errorf("%s: category %q, want %q", tg.Name, tg.Category, target.Skill)
		}
		want, ok := wantDead[tg.Name]
		if !ok {
			t.Errorf("unexpected skill: %s", tg.Name)
			continue
		}
		if tg.Dead != want {
			t.Errorf("%s: dead=%v, want %v (reason=%q)", tg.Name, tg.Dead, want, tg.Reason)
		}
	}
}

func TestScanMissingDirIsNotError(t *testing.T) {
	got, err := New(filepath.Join(t.TempDir(), "does-not-exist")).Scan(context.Background())
	if err != nil {
		t.Fatalf("a missing dir should be silent, got: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0 targets, got %d", len(got))
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	mkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
