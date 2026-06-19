// Package ui cuida da apresentação: logo, ajuda e relatório formatado.
// Só renderiza — não contém lógica de scan ou remoção.
package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"cts/internal/target"
)

var (
	accent   = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	dim      = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	deadSt   = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
	catSt    = lipgloss.NewStyle().Foreground(lipgloss.Color("111")).Bold(true)
	catOrder = []target.Category{target.Skill, target.Agent, target.Plugin, target.MCP}
)

const logoArt = ` ██████╗████████╗███████╗
██╔════╝╚══██╔══╝██╔════╝
██║        ██║   ███████╗
██║        ██║   ╚════██║
╚██████╗   ██║   ███████║
 ╚═════╝   ╚═╝   ╚══════╝`

// Logo devolve o banner ASCII do cts.
func Logo() string {
	art := accent.Render(logoArt)
	tag := dim.Render("Cut The Shit · limpa skills, agentes, plugins e MCP mortos")
	return "\n" + art + "\n\n" + tag + "\n"
}

// Help devolve a ajuda formatada.
func Help() string {
	var b strings.Builder
	b.WriteString(accent.Render("cts") + dim.Render(" — comandos") + "\n\n")
	rows := [][2]string{
		{"cts", "menu interativo"},
		{"cts scan", "lista o que há na máquina (read-only)"},
		{"cts clean", "escolher numa lista e remover"},
		{"cts purge", "mostra o que removeria (só mortos, dry-run)"},
		{"cts purge --yes", "remove os mortos de verdade (com backup)"},
		{"cts help", "esta ajuda"},
	}
	for _, r := range rows {
		b.WriteString("  " + accent.Render(fmt.Sprintf("%-17s", r[0])) + dim.Render(r[1]) + "\n")
	}
	b.WriteString("\n" + dim.Render("Seguro: dry-run por padrão, confirma antes de apagar, backup em .cts-backups/."))
	return b.String()
}

// Report formata os alvos agrupados por categoria.
func Report(targets []target.Target) string {
	if len(targets) == 0 {
		return dim.Render("nada encontrado. Máquina limpa. ✨") + "\n"
	}

	groups := make(map[target.Category][]target.Target)
	for _, t := range targets {
		groups[t.Category] = append(groups[t.Category], t)
	}

	var b strings.Builder
	deadN := 0
	for _, c := range catOrder {
		ts := groups[c]
		if len(ts) == 0 {
			continue
		}
		sort.Slice(ts, func(i, j int) bool { return ts[i].Name < ts[j].Name })
		b.WriteString("\n" + catSt.Render(strings.ToUpper(string(c))) + "\n")
		for _, t := range ts {
			cells := fmt.Sprintf("%-28s %10s", t.Name, HumanSize(t.SizeBytes))
			if t.Dead {
				deadN++
				b.WriteString(deadSt.Render("  ✗ " + cells))
			} else {
				b.WriteString("    " + cells)
			}
			if t.Reason != "" {
				b.WriteString("  " + dim.Render(t.Reason))
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n" + dim.Render(fmt.Sprintf("%d alvos · %d mortos", len(targets), deadN)) + "\n")
	return b.String()
}

// HumanSize formata bytes em algo legível (1.5KB, 279.0MB).
func HumanSize(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "KMGTPE"[exp])
}
