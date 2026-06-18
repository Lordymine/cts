package agents

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"cts/internal/target"
)

// fakeLister implementa Lister sem tocar no PATH real.
type fakeLister map[string]bool

func (f fakeLister) IsInstalled(bin string) bool { return f[bin] }

func TestScan(t *testing.T) {
	home := t.TempDir()
	mkdir(t, filepath.Join(home, ".qwen"))  // config existe, binário NÃO instalado → morto
	mkdir(t, filepath.Join(home, ".codex")) // config existe, binário instalado → vivo
	// "ghost": no catálogo, mas sem config em disco e não instalado → ignorado

	catalog := []Agent{
		{Name: "qwen", Bin: "qwen", Dirs: []string{".qwen", ".gqwen"}},
		{Name: "codex", Bin: "codex", Dirs: []string{".codex"}},
		{Name: "ghost", Bin: "ghost", Dirs: []string{".ghost"}},
	}
	lister := fakeLister{"codex": true} // só codex instalado

	got, err := New(home, lister, catalog).Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	byName := make(map[string]target.Target, len(got))
	for _, tg := range got {
		if tg.Category != target.Agent {
			t.Errorf("%s: categoria %q, queria %q", tg.Name, tg.Category, target.Agent)
		}
		byName[tg.Name] = tg
	}

	if len(got) != 2 {
		t.Fatalf("queria 2 alvos (qwen, codex), veio %d: %+v", len(got), got)
	}
	if !byName["qwen"].Dead {
		t.Errorf("qwen: config existe e binário não instalado → deveria ser morto (reason=%q)", byName["qwen"].Reason)
	}
	if byName["codex"].Dead {
		t.Errorf("codex: instalado → não deveria ser morto")
	}
	if _, ok := byName["ghost"]; ok {
		t.Errorf("ghost: sem config e não instalado → não deveria aparecer")
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}
