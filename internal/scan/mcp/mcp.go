// Package mcp inventories the configured MCP servers and flags the stdio ones
// whose command is not installed. Removing an MCP server is a config edit (via
// `claude mcp remove`), not a file delete — so targets come out without Paths.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lordymine/cts/internal/target"
)

// Scanner reads a config file (e.g. ~/.claude.json). hasCommand is injected so
// tests don't depend on the real PATH.
type Scanner struct {
	configPath string
	hasCommand func(string) bool
}

// New creates a Scanner. hasCommand reports whether a binary exists (e.g. exec.LookPath).
func New(configPath string, hasCommand func(string) bool) Scanner {
	return Scanner{configPath: configPath, hasCommand: hasCommand}
}

// Category satisfies scan.Scanner.
func (s Scanner) Category() target.Category { return target.MCP }

type config struct {
	McpServers map[string]server `json:"mcpServers"`
	Projects   map[string]struct {
		McpServers map[string]server `json:"mcpServers"`
	} `json:"projects"`
}

// server: only command matters — empty means an http/sse transport with no
// binary to check.
type server struct {
	Command string `json:"command"`
}

// Scan inventories the MCP servers from the user scope and from each project.
func (s Scanner) Scan(ctx context.Context) ([]target.Target, error) {
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // no config: nothing to inventory
		}
		return nil, fmt.Errorf("read %s: %w", s.configPath, err)
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
		targets = append(targets, s.inspect(name, cfg.McpServers[name], "user", true))
	}
	for _, proj := range sortedKeys(cfg.Projects) {
		servers := cfg.Projects[proj].McpServers
		for _, name := range sortedKeys(servers) {
			targets = append(targets, s.inspect(name, servers[name], "project: "+filepath.Base(proj), false))
		}
	}
	return targets, nil
}

// inspect builds the target for an MCP server. Paths is intentionally empty;
// removal is via command. Only the user scope gets an automatic command
// (`claude mcp remove -s user`); project servers stay inventory-only (the user
// removes them inside the project).
func (s Scanner) inspect(name string, srv server, scope string, userScope bool) target.Target {
	reason := scope
	if bin := firstWord(srv.Command); bin != "" && !s.hasCommand(bin) {
		reason = scope + " — command not found: " + bin
	}
	t := target.Target{Name: name, Category: target.MCP, Reason: reason}
	if userScope {
		t.Uninstall = []string{"claude", "mcp", "remove", name, "-s", "user"}
	}
	return t
}

func firstWord(cmd string) string {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

// sortedKeys returns a map's keys in order — deterministic output.
func sortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
