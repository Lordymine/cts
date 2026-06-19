// Package plugins finds orphan plugin marketplaces: clones/caches on disk for a
// marketplace that no longer has any installed plugin.
package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Lordymine/cts/internal/dirsize"
	"github.com/Lordymine/cts/internal/target"
)

// subdirs of the plugins directory where each marketplace leaves a trace on disk.
var subdirs = []string{"marketplaces", "cache"}

// Scanner sweeps the plugins directory (e.g. ~/.claude/plugins).
type Scanner struct {
	root string
}

// New creates a Scanner for the plugins root directory.
func New(root string) Scanner {
	return Scanner{root: root}
}

// Category satisfies scan.Scanner.
func (s Scanner) Category() target.Category { return target.Plugin }

// Scan cross-checks the marketplaces on disk against those with an installed
// plugin (read from installed_plugins.json). No installed plugin = orphan.
func (s Scanner) Scan(ctx context.Context) ([]target.Target, error) {
	active, err := s.activeMarketplaces()
	if err != nil {
		return nil, err
	}

	var targets []target.Target
	for _, name := range s.marketplaceNames() {
		if err := ctx.Err(); err != nil {
			return targets, err
		}
		targets = append(targets, s.inspect(name, active))
	}
	return targets, nil
}

// inspect joins a marketplace's on-disk paths (marketplaces/<name>, cache/<name>)
// and decides whether it is orphan.
func (s Scanner) inspect(name string, active map[string]bool) target.Target {
	var paths []string
	var size int64
	for _, sub := range subdirs {
		p := filepath.Join(s.root, sub, name)
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			paths = append(paths, p)
			size += dirsize.Of(p)
		}
	}

	t := target.Target{Name: name, Category: target.Plugin, Paths: paths, SizeBytes: size}
	if !active[name] {
		t.Dead = true
		t.Reason = "orphan marketplace (no installed plugin)"
	}
	return t
}

// activeMarketplaces reads installed_plugins.json and returns the set of
// marketplaces that still have an installed plugin. Keys are "<plugin>@<marketplace>".
func (s Scanner) activeMarketplaces() (map[string]bool, error) {
	data, err := os.ReadFile(filepath.Join(s.root, "installed_plugins.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil // no manifest: nothing installed, everything is orphan
		}
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var m struct {
		Plugins map[string]json.RawMessage `json:"plugins"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	active := make(map[string]bool)
	for key := range m.Plugins {
		if _, mkt, ok := strings.Cut(key, "@"); ok {
			active[mkt] = true
		}
	}
	return active, nil
}

// marketplaceNames returns, sorted, the marketplace names present on disk
// (union of marketplaces/ and cache/).
func (s Scanner) marketplaceNames() []string {
	seen := make(map[string]bool)
	for _, sub := range subdirs {
		entries, err := os.ReadDir(filepath.Join(s.root, sub))
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				seen[e.Name()] = true
			}
		}
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	sort.Strings(names) // deterministic — maps have no order
	return names
}
