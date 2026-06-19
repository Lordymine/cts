package plugins

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"cts/internal/target"
)

func TestScanFlagsOrphanMarketplace(t *testing.T) {
	root := t.TempDir()
	// manifest: only claude-plugins-official has an installed plugin
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
			t.Errorf("%s: category %q, want %q", tg.Name, tg.Category, target.Plugin)
		}
		byName[tg.Name] = tg
	}

	if len(got) != 2 {
		t.Fatalf("want 2 marketplaces, got %d: %+v", len(got), got)
	}
	if byName["claude-plugins-official"].Dead {
		t.Error("claude-plugins-official has an installed plugin → not orphan")
	}
	td := byName["thedotmack"]
	if !td.Dead {
		t.Error("thedotmack has no installed plugin → orphan")
	}
	if len(td.Paths) != 2 {
		t.Errorf("thedotmack should join marketplaces/ + cache/ (2 paths), got %d", len(td.Paths))
	}
}

func TestScanNoManifestEverythingOrphan(t *testing.T) {
	root := t.TempDir()
	mkdir(t, filepath.Join(root, "marketplaces", "whatever"))

	got, err := New(root).Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(got) != 1 || !got[0].Dead {
		t.Fatalf("with no manifest, the marketplace should be orphan: %+v", got)
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
