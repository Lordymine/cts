package agents

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"cts/internal/configroots"
	"cts/internal/target"
)

// fakeLister implements Lister without touching the real PATH.
type fakeLister map[string]bool

func (f fakeLister) IsInstalled(bin string) bool { return f[bin] }

func TestScanFlagsOrphanConfigAcrossRoots(t *testing.T) {
	home := t.TempDir()
	xdg := t.TempDir()
	mkdir(t, filepath.Join(home, ".qwen"))  // dotted home config, binary missing → dead
	mkdir(t, filepath.Join(xdg, "gemini"))  // plain (XDG-style) config, binary missing → dead
	mkdir(t, filepath.Join(home, ".codex")) // config + binary installed → alive

	roots := []configroots.Root{
		{Path: home, Dotted: true},
		{Path: xdg, Dotted: false},
	}
	catalog := []Agent{
		{Name: "qwen", Bin: "qwen", Dirs: []string{"qwen"}},
		{Name: "gemini", Bin: "gemini", Dirs: []string{"gemini"}},
		{Name: "codex", Bin: "codex", Dirs: []string{"codex"}},
		{Name: "ghost", Bin: "ghost", Dirs: []string{"ghost"}}, // no config, not installed → skipped
	}
	lister := fakeLister{"codex": true}

	got, err := New(roots, lister, catalog).Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	byName := make(map[string]target.Target, len(got))
	for _, tg := range got {
		if tg.Category != target.Agent {
			t.Errorf("%s: category %q, want %q", tg.Name, tg.Category, target.Agent)
		}
		byName[tg.Name] = tg
	}

	if !byName["qwen"].Dead {
		t.Error("qwen: dotted home config + no binary should be dead")
	}
	if !byName["gemini"].Dead {
		t.Error("gemini: plain-root config + no binary should be dead (cross-platform coverage)")
	}
	if byName["codex"].Dead {
		t.Error("codex: installed should not be dead")
	}
	if _, ok := byName["ghost"]; ok {
		t.Error("ghost: no config and not installed should be skipped")
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}
