// Package agents finds agent CLIs and flags orphan config: a config directory
// whose binary is no longer installed (e.g. ~/.qwen with no `qwen` on PATH).
package agents

import (
	"context"
	"os"

	"github.com/Lordymine/cts/internal/configroots"
	"github.com/Lordymine/cts/internal/dirsize"
	"github.com/Lordymine/cts/internal/target"
)

// Lister reports whether an agent binary is installed. Injected so tests run
// without touching the real PATH.
type Lister interface {
	IsInstalled(bin string) bool
}

// Agent is a catalog entry: the binary to check, its config dir base names, and
// how to uninstall the package (when installed).
type Agent struct {
	Name    string
	Bin     string
	Dirs    []string // config dir base names, e.g. "qwen" (resolved per root)
	Manager string   // "npm" | "bun" | "uv" — empty means no uninstall command
	Package string   // package name in the manager
}

// Scanner cross-checks the catalog against what is installed and what is on disk
// across the OS config roots.
type Scanner struct {
	roots   []configroots.Root
	lister  Lister
	catalog []Agent
}

// New creates a Scanner. roots are the OS config base dirs (see configroots).
func New(roots []configroots.Root, lister Lister, catalog []Agent) Scanner {
	return Scanner{roots: roots, lister: lister, catalog: catalog}
}

// Category satisfies scan.Scanner.
func (s Scanner) Category() target.Category { return target.Agent }

// Scan walks the catalog. An agent becomes a target only if it has a presence on
// the machine (installed or with a config dir). Config without binary = dead.
func (s Scanner) Scan(ctx context.Context) ([]target.Target, error) {
	var targets []target.Target
	for _, a := range s.catalog {
		if err := ctx.Err(); err != nil {
			return targets, err
		}
		if t, present := s.inspect(a); present {
			targets = append(targets, t)
		}
	}
	return targets, nil
}

// inspect builds the target for an agent, looking across all config roots.
func (s Scanner) inspect(a Agent) (target.Target, bool) {
	var paths []string
	var size int64
	for _, root := range s.roots {
		for _, base := range a.Dirs {
			p := root.Entry(base)
			info, err := os.Stat(p)
			if err != nil {
				continue
			}
			paths = append(paths, p)
			if info.IsDir() {
				size += dirsize.Of(p)
			} else {
				size += info.Size()
			}
		}
	}

	installed := s.lister.IsInstalled(a.Bin)
	if !installed && len(paths) == 0 {
		return target.Target{}, false
	}

	t := target.Target{Name: a.Name, Category: target.Agent, Paths: paths, SizeBytes: size}
	if installed {
		t.Uninstall = uninstallCmd(a) // only uninstall what is installed
	} else if len(paths) > 0 {
		t.Dead = true
		t.Reason = "orphan config (binary not installed)"
	}
	return t, true
}

// uninstallCmd builds the uninstall command for the agent's package manager.
// With no known package/manager it returns nil (the target is removed by file only).
func uninstallCmd(a Agent) []string {
	if a.Package == "" {
		return nil
	}
	switch a.Manager {
	case "npm":
		return []string{"npm", "rm", "-g", a.Package}
	case "bun":
		return []string{"bun", "rm", "-g", a.Package}
	case "uv":
		return []string{"uv", "tool", "uninstall", a.Package}
	default:
		return nil
	}
}

// DefaultCatalog lists third-party agents that tend to leave orphan config.
// Extend as needed. The primary agents (claude, codex, pi, opencode) are left
// out on purpose: they are kept, not cleanup candidates.
func DefaultCatalog() []Agent {
	return []Agent{
		{Name: "qwen", Bin: "qwen", Dirs: []string{"qwen", "gqwen"}, Manager: "npm", Package: "@qwen-code/qwen-code"},
		{Name: "gemini", Bin: "gemini", Dirs: []string{"gemini"}, Manager: "npm", Package: "@google/gemini-cli"},
		{Name: "kimi", Bin: "kimi", Dirs: []string{"kimi"}, Manager: "uv", Package: "kimi-cli"},
		{Name: "verboo", Bin: "verboo", Dirs: []string{"verboo"}, Manager: "npm", Package: "@verboo/code"},
		{Name: "command-code", Bin: "command-code", Dirs: []string{"commandcode"}, Manager: "npm", Package: "command-code"},
		{Name: "mimo", Bin: "mimo", Dirs: []string{"mimo"}, Manager: "npm", Package: "@mimo-ai/cli"},
		// No known manager: removed by file only (orphan config).
		{Name: "autocodex", Bin: "autocodex", Dirs: []string{"autocodex"}},
		{Name: "goclaw", Bin: "goclaw", Dirs: []string{"goclaw"}},
		{Name: "hermes", Bin: "hermes", Dirs: []string{"hermes"}},
		{Name: "codebuddy", Bin: "codebuddy", Dirs: []string{"codebuddy"}},
		{Name: "iflow", Bin: "iflow", Dirs: []string{"iflow"}},
		{Name: "zencoder", Bin: "zencoder", Dirs: []string{"zencoder"}},
	}
}
