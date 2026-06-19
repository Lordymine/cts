// Package ui handles presentation: logo, help and the formatted report.
// It only renders — it holds no scan or removal logic.
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

// Logo returns the cts ASCII banner.
func Logo() string {
	art := accent.Render(logoArt)
	tag := dim.Render("Cut The Shit · clean dead skills, agents, plugins and MCP servers")
	return "\n" + art + "\n\n" + tag + "\n"
}

// Help returns the formatted help text.
func Help() string {
	var b strings.Builder
	b.WriteString(accent.Render("cts") + dim.Render(" — commands") + "\n\n")
	rows := [][2]string{
		{"cts", "interactive menu"},
		{"cts scan", "list what's on the machine (read-only)"},
		{"cts clean", "pick items from a list and remove"},
		{"cts purge", "show what it would remove (dead only, dry-run)"},
		{"cts purge --yes", "actually remove the dead ones (with backup)"},
		{"cts version", "print the version"},
		{"cts help", "this help"},
	}
	for _, r := range rows {
		b.WriteString("  " + accent.Render(fmt.Sprintf("%-17s", r[0])) + dim.Render(r[1]) + "\n")
	}
	b.WriteString("\n" + dim.Render("Safe: dry-run by default, confirms before deleting, backup in .cts-backups/."))
	return b.String()
}

// Report formats the targets grouped by category.
func Report(targets []target.Target) string {
	if len(targets) == 0 {
		return dim.Render("nothing found. Machine is clean. ✨") + "\n"
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
				b.WriteString(deadSt.Render("  x " + cells))
			} else {
				b.WriteString("    " + cells)
			}
			if t.Reason != "" {
				b.WriteString("  " + dim.Render(t.Reason))
			}
			b.WriteString("\n")
		}
	}
	b.WriteString("\n" + dim.Render(fmt.Sprintf("%d targets · %d dead", len(targets), deadN)) + "\n")
	return b.String()
}

// HumanSize formats bytes into something readable (1.5KB, 279.0MB).
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
