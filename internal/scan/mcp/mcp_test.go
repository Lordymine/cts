package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cts/internal/target"
)

func TestScanInventariaEFlagaComandoQuebrado(t *testing.T) {
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

	installed := map[string]bool{"npx": true, "node": true} // uvx NÃO instalado
	got, err := New(cfgPath, func(c string) bool { return installed[c] }).Scan(context.Background())
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	byName := make(map[string]target.Target, len(got))
	for _, tg := range got {
		if tg.Category != target.MCP {
			t.Errorf("%s: categoria %q, queria %q", tg.Name, tg.Category, target.MCP)
		}
		if len(tg.Paths) != 0 {
			t.Errorf("%s: MCP não tem Paths (remoção é config-edit), veio %v", tg.Name, tg.Paths)
		}
		byName[tg.Name] = tg
	}

	if len(got) != 3 {
		t.Fatalf("queria 3 servers (context7, plane, notion), veio %d", len(got))
	}
	if !strings.Contains(byName["plane"].Reason, "comando não encontrado") {
		t.Errorf("plane (uvx ausente) deveria marcar comando não encontrado, veio %q", byName["plane"].Reason)
	}
	if strings.Contains(byName["context7"].Reason, "não encontrado") {
		t.Errorf("context7 (npx ok) não deveria marcar quebrado, veio %q", byName["context7"].Reason)
	}
	if !strings.Contains(byName["notion"].Reason, "projeto") {
		t.Errorf("notion deveria indicar escopo de projeto, veio %q", byName["notion"].Reason)
	}
	if len(byName["context7"].Uninstall) == 0 {
		t.Error("context7 (user scope) deveria ter comando de remoção (claude mcp remove)")
	}
	if len(byName["notion"].Uninstall) != 0 {
		t.Errorf("notion (projeto) não deve ter comando automático, veio %v", byName["notion"].Uninstall)
	}
}

func TestScanSemConfigNaoEhErro(t *testing.T) {
	got, err := New(filepath.Join(t.TempDir(), "nao-existe.json"), func(string) bool { return true }).Scan(context.Background())
	if err != nil {
		t.Fatalf("config inexistente deveria ser silencioso: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("queria 0, veio %d", len(got))
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
