// Package plugins acha marketplaces de plugin órfãos: clones/caches em disco de
// um marketplace que não tem mais nenhum plugin instalado.
package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cts/internal/dirsize"
	"cts/internal/target"
)

// subdirs do diretório de plugins onde cada marketplace deixa rastro em disco.
var subdirs = []string{"marketplaces", "cache"}

// Scanner varre o diretório de plugins (ex.: ~/.claude/plugins).
type Scanner struct {
	root string
}

// New cria um Scanner para o diretório raiz de plugins.
func New(root string) Scanner {
	return Scanner{root: root}
}

// Category satisfaz scan.Scanner.
func (s Scanner) Category() target.Category { return target.Plugin }

// Scan cruza os marketplaces em disco com os que têm plugin instalado
// (lido de installed_plugins.json). Sem plugin instalado = órfão.
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

// inspect junta os caminhos em disco (marketplaces/<name>, cache/<name>) de um
// marketplace e decide se é órfão.
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
		t.Reason = "marketplace órfão (sem plugin instalado)"
	}
	return t
}

// activeMarketplaces lê installed_plugins.json e devolve o conjunto de
// marketplaces que ainda têm plugin instalado. Chave é "<plugin>@<marketplace>".
func (s Scanner) activeMarketplaces() (map[string]bool, error) {
	data, err := os.ReadFile(filepath.Join(s.root, "installed_plugins.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]bool{}, nil // sem manifest: nada instalado, tudo é órfão
		}
		return nil, fmt.Errorf("ler manifest: %w", err)
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

// marketplaceNames devolve, ordenado, os nomes de marketplace presentes em
// disco (união de marketplaces/ e cache/).
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
	sort.Strings(names) // determinístico — mapa não tem ordem
	return names
}
