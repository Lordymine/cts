// Package agents acha CLIs de agente e marca config órfã: pasta de config
// cujo binário não está mais instalado (ex.: ~/.qwen sem o `qwen` no PATH).
package agents

import (
	"context"
	"os"
	"path/filepath"

	"cts/internal/dirsize"
	"cts/internal/target"
)

// Lister diz se o binário de um agente está instalado. É injetado de propósito:
// o teste passa um fake e não toca no PATH real da máquina.
type Lister interface {
	IsInstalled(bin string) bool
}

// Agent é uma entrada do catálogo: o binário a checar, as pastas de config, e
// como desinstalar o pacote (quando instalado).
type Agent struct {
	Name    string
	Bin     string
	Dirs    []string // nomes de pasta sob home (ex.: ".qwen", ".gqwen")
	Manager string   // "npm" | "bun" | "uv" — vazio = sem comando de uninstall
	Package string   // nome do pacote no manager
}

// Scanner cruza o catálogo com o que está instalado e o que há em disco.
type Scanner struct {
	home    string
	lister  Lister
	catalog []Agent
}

// New cria um Scanner. home é onde vivem as pastas de config (ex.: o diretório do usuário).
func New(home string, lister Lister, catalog []Agent) Scanner {
	return Scanner{home: home, lister: lister, catalog: catalog}
}

// Category satisfaz scan.Scanner.
func (s Scanner) Category() target.Category { return target.Agent }

// Scan percorre o catálogo. Um agente só vira alvo se tiver presença na máquina
// (instalado ou com pasta de config). Config sem binário = morto.
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

// inspect monta o Target de um agente. present=false quando não há nada dele na
// máquina (não instalado e sem config) — aí não é alvo de nada.
func (s Scanner) inspect(a Agent) (target.Target, bool) {
	var paths []string
	var size int64
	for _, d := range a.Dirs {
		p := filepath.Join(s.home, d)
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

	installed := s.lister.IsInstalled(a.Bin)
	if !installed && len(paths) == 0 {
		return target.Target{}, false
	}

	t := target.Target{Name: a.Name, Category: target.Agent, Paths: paths, SizeBytes: size}
	if installed {
		t.Uninstall = uninstallCmd(a) // só desinstala o que está instalado
	} else if len(paths) > 0 {
		t.Dead = true
		t.Reason = "config órfã (binário não instalado)"
	}
	return t, true
}

// uninstallCmd monta o comando de desinstalação conforme o gerenciador. Sem
// pacote/manager conhecido, devolve nil (o alvo será removido só por arquivo).
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

// DefaultCatalog lista agentes de terceiros que costumam deixar config órfã.
// Extensível — adicione conforme aparecerem. Os agentes principais (claude,
// codex, pi, opencode) ficam de fora: são mantidos, não candidatos a limpeza.
func DefaultCatalog() []Agent {
	return []Agent{
		{Name: "qwen", Bin: "qwen", Dirs: []string{".qwen", ".gqwen"}, Manager: "npm", Package: "@qwen-code/qwen-code"},
		{Name: "gemini", Bin: "gemini", Dirs: []string{".gemini"}, Manager: "npm", Package: "@google/gemini-cli"},
		{Name: "kimi", Bin: "kimi", Dirs: []string{".kimi"}, Manager: "uv", Package: "kimi-cli"},
		{Name: "verboo", Bin: "verboo", Dirs: []string{".verboo"}, Manager: "npm", Package: "@verboo/code"},
		{Name: "command-code", Bin: "command-code", Dirs: []string{".commandcode"}, Manager: "npm", Package: "command-code"},
		{Name: "mimo", Bin: "mimo", Dirs: []string{".mimo"}, Manager: "npm", Package: "@mimo-ai/cli"},
		// Sem manager conhecido: removidos só por arquivo (config órfã).
		{Name: "autocodex", Bin: "autocodex", Dirs: []string{".autocodex"}},
		{Name: "goclaw", Bin: "goclaw", Dirs: []string{".goclaw"}},
		{Name: "hermes", Bin: "hermes", Dirs: []string{".hermes"}},
		{Name: "codebuddy", Bin: "codebuddy", Dirs: []string{".codebuddy"}},
		{Name: "iflow", Bin: "iflow", Dirs: []string{".iflow"}},
		{Name: "zencoder", Bin: "zencoder", Dirs: []string{".zencoder"}},
	}
}
