package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cts/internal/target"
)

func TestScanInventoriesAndFlagsBrokenCommand(t *testing.T) {
	cfgPath := filepath.Join(t.TempDir(), ".claude.json")
	writeFile(t, cfgPath, `{
		"mcpServers": {
			"context7": {"command": "npx"},
			"plane": {"command": "uvx"}
		},
		"projects": {
			"D:/proj/x": {"mcpServers": {"notion": {"command": "node"}}}
		}
	}`)

	installed := map[string]bool{"npx": true, "node": true} // uvx NOT installed
	got, err := New(cfgPath, func(c string) bool { return installed[c] }).Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	byName := make(map[string]target.Target, len(got))
	for _, tg := range got {
		if tg.Category != target.MCP {
			t.Errorf("%s: category %q, want %q", tg.Name, tg.Category, target.MCP)
		}
		if len(tg.Paths) != 0 {
			t.Errorf("%s: MCP has no Paths (removal is a command), got %v", tg.Name, tg.Paths)
		}
		byName[tg.Name] = tg
	}

	if len(got) != 3 {
		t.Fatalf("want 3 servers (context7, plane, notion), got %d", len(got))
	}
	if !strings.Contains(byName["plane"].Reason, "command not found") {
		t.Errorf("plane (uvx missing) should flag command not found, got %q", byName["plane"].Reason)
	}
	if strings.Contains(byName["context7"].Reason, "not found") {
		t.Errorf("context7 (npx ok) should not be flagged, got %q", byName["context7"].Reason)
	}
	if !strings.Contains(byName["notion"].Reason, "project") {
		t.Errorf("notion should indicate project scope, got %q", byName["notion"].Reason)
	}
	if len(byName["context7"].Uninstall) == 0 {
		t.Error("context7 (user scope) should have a removal command (claude mcp remove)")
	}
	if len(byName["notion"].Uninstall) != 0 {
		t.Errorf("notion (project) should not have an automatic command, got %v", byName["notion"].Uninstall)
	}
}

func TestScanNoConfigIsNotError(t *testing.T) {
	got, err := New(filepath.Join(t.TempDir(), "does-not-exist.json"), func(string) bool { return true }).Scan(context.Background())
	if err != nil {
		t.Fatalf("a missing config should be silent: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("want 0, got %d", len(got))
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
