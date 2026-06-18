package plugins

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"cts/internal/target"
)

func TestScanMarcaMarketplaceOrfao(t *testing.T) {
	root := t.TempDir()
	// manifest: só claude-plugins-official tem plugin instalado
	writeFile(t, filepath.Join(root, "installed_plugins.json"),
		`{"plugins":{"github@claude-plugins-official":[{}],"context7@claude-plugins-official":[{}]}}`)
	mkdir(t, filepath.Join(root, "marketplaces", "claude-plugins-official"))
	mkdir(t, filepath.Join(root, "marketplaces", "thedotmack"))
	mkdir(t, filepath.Join(root, "cache", "thedotmack"))

	got, err := New(root).Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	byName := make(map[string]target.Target, len(got))
	for _, tg := range got {
		if tg.Category != target.Plugin {
			t.Errorf("%s: categoria %q, queria %q", tg.Name, tg.Category, target.Plugin)
		}
		byName[tg.Name] = tg
	}

	if len(got) != 2 {
		t.Fatalf("queria 2 marketplaces, veio %d: %+v", len(got), got)
	}
	if byName["claude-plugins-official"].Dead {
		t.Error("claude-plugins-official tem plugin instalado → não é órfão")
	}
	td := byName["thedotmack"]
	if !td.Dead {
		t.Error("thedotmack sem plugin instalado → órfão")
	}
	if len(td.Paths) != 2 {
		t.Errorf("thedotmack deveria juntar marketplaces/ + cache/ (2 paths), veio %d", len(td.Paths))
	}
}

func TestScanSemManifestTudoOrfao(t *testing.T) {
	root := t.TempDir()
	mkdir(t, filepath.Join(root, "marketplaces", "qualquer"))

	got, err := New(root).Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(got) != 1 || !got[0].Dead {
		t.Fatalf("sem manifest, marketplace deveria ser órfão: %+v", got)
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
