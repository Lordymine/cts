// Package mcp inventaria os MCP servers configurados e marca os de stdio cujo
// comando não está instalado. Remover um MCP é editar o JSON (config-edit), não
// apagar arquivo — por isso os alvos saem sem Paths.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cts/internal/target"
)

// Scanner lê um arquivo de config (ex.: ~/.claude.json). hasCommand é injetado
// para testar sem depender do PATH real.
type Scanner struct {
	configPath string
	hasCommand func(string) bool
}

// New cria um Scanner. hasCommand diz se um binário existe (ex.: exec.LookPath).
func New(configPath string, hasCommand func(string) bool) Scanner {
	return Scanner{configPath: configPath, hasCommand: hasCommand}
}

// Category satisfaz scan.Scanner.
func (s Scanner) Category() target.Category { return target.MCP }

type config struct {
	McpServers map[string]server `json:"mcpServers"`
	Projects   map[string]struct {
		McpServers map[string]server `json:"mcpServers"`
	} `json:"projects"`
}

// server: só o command importa — vazio significa transporte http/sse, sem
// binário pra checar.
type server struct {
	Command string `json:"command"`
}

// Scan inventaria os MCP servers do escopo user e de cada projeto.
func (s Scanner) Scan(ctx context.Context) ([]target.Target, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // sem config: nada a inventariar
		}
		return nil, fmt.Errorf("ler %s: %w", s.configPath, err)
	}

	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", s.configPath, err)
	}

	var targets []target.Target
	for _, name := range sortedKeys(cfg.McpServers) {
		if err := ctx.Err(); err != nil {
			return targets, err
		}
		targets = append(targets, s.inspect(name, cfg.McpServers[name], "user"))
	}
	for _, proj := range sortedKeys(cfg.Projects) {
		servers := cfg.Projects[proj].McpServers
		for _, name := range sortedKeys(servers) {
			targets = append(targets, s.inspect(name, servers[name], "projeto: "+filepath.Base(proj)))
		}
	}
	return targets, nil
}

// inspect monta o Target de um MCP. Paths fica vazio de propósito.
func (s Scanner) inspect(name string, srv server, scope string) target.Target {
	reason := scope
	if bin := firstWord(srv.Command); bin != "" && !s.hasCommand(bin) {
		reason = scope + " — comando não encontrado: " + bin
	}
	return target.Target{Name: name, Category: target.MCP, Reason: reason}
}

func firstWord(cmd string) string {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// sortedKeys devolve as chaves de um mapa em ordem — saída determinística.
func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
