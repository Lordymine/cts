// Command cts — Cut The Shit. Limpa skills, agentes, plugins e MCP mortos da máquina.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"cts/internal/scan"
	"cts/internal/scan/agents"
	"cts/internal/scan/skills"
	"cts/internal/target"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "scan" {
		fmt.Fprintln(os.Stderr, "uso: cts scan")
		os.Exit(2)
	}
	if err := runScan(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "cts: %v\n", err)
		os.Exit(1)
	}
}

func runScan(ctx context.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}

	scanners := []scan.Scanner{
		skills.New(filepath.Join(home, ".claude", "skills")),
		agents.New(home, pathLister{}, agents.DefaultCatalog()),
	}

	targets, err := scan.Run(ctx, scanners...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aviso: %v\n", err) // erro parcial: ainda mostramos o que achou
	}
	printReport(targets)
	return nil
}

// pathLister checa instalação via PATH. É o adapter real do agents.Lister;
// fica na borda (IO), longe do domínio testável.
type pathLister struct{}

func (pathLister) IsInstalled(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

func printReport(targets []target.Target) {
	if len(targets) == 0 {
		fmt.Println("nada encontrado.")
		return
	}
	var dead int
	for _, t := range targets {
		mark := "  "
		if t.Dead {
			mark, dead = "✗ ", dead+1
		}
		fmt.Printf("%s%-7s %-28s %9s  %s\n", mark, t.Category, t.Name, humanSize(t.SizeBytes), t.Reason)
	}
	fmt.Printf("\n%d alvos, %d mortos.\n", len(targets), dead)
}

func humanSize(b int64) string {
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
