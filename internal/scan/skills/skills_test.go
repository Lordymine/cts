package skills

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"cts/internal/target"
)

func TestScanMarcaSkillSemSkillMdComoMorta(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "healthy", "SKILL.md"), "# ok") // saudável
	mkdir(t, filepath.Join(root, "broken"))                          // sem SKILL.md → morta

	got, err := New(root).Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	wantDead := map[string]bool{"healthy": false, "broken": true}
	if len(got) != len(wantDead) {
		t.Fatalf("achou %d skills, queria %d", len(got), len(wantDead))
	}
	for _, tg := range got {
		if tg.Category != target.Skill {
			t.Errorf("%s: categoria %q, queria %q", tg.Name, tg.Category, target.Skill)
		}
		want, ok := wantDead[tg.Name]
		if !ok {
			t.Errorf("skill inesperada: %s", tg.Name)
			continue
		}
		if tg.Dead != want {
			t.Errorf("%s: dead=%v, queria %v (reason=%q)", tg.Name, tg.Dead, want, tg.Reason)
		}
	}
}

func TestScanDirInexistenteNaoEhErro(t *testing.T) {
	got, err := New(filepath.Join(t.TempDir(), "nao-existe")).Scan(context.Background())
	if err != nil {
		t.Fatalf("dir inexistente deveria ser silencioso, veio: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("queria 0 alvos, veio %d", len(got))
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
