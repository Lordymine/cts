// Command cts — Cut The Shit. Limpa skills, agentes, plugins e MCP mortos da máquina.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"cts/internal/remove"
	"cts/internal/scan"
	"cts/internal/scan/agents"
	"cts/internal/scan/plugins"
	"cts/internal/scan/skills"
	"cts/internal/target"
)

func main() {
	args := os.Args[1:]

	var err error
	switch {
	case len(args) == 0, args[0] == "clean":
		err = runInteractive(context.Background())
	case args[0] == "scan":
		err = runScan(context.Background())
	case args[0] == "purge":
		err = runPurge(context.Background(), len(args) > 1 && args[1] == "--yes")
	default:
		usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "cts: %v\n", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "uso: cts [clean] | cts scan | cts purge [--yes]")
}

// buildScanners monta os scanners com os caminhos reais. IO e wiring vivem aqui,
// na borda — os scanners em si são testáveis isolados.
func buildScanners(home string) []scan.Scanner {
	return []scan.Scanner{
		skills.New(filepath.Join(home, ".claude", "skills")),
		agents.New(home, pathLister{}, agents.DefaultCatalog()),
		plugins.New(filepath.Join(home, ".claude", "plugins")),
	}
}

func runScan(ctx context.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}
	targets, err := scan.Run(ctx, buildScanners(home)...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aviso: %v\n", err)
	}
	printReport(targets)
	return nil
}

// runPurge remove SÓ os mortos. Dry-run por padrão; --yes executa (com backup).
func runPurge(ctx context.Context, execute bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}
	targets, err := scan.Run(ctx, buildScanners(home)...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aviso: %v\n", err)
	}

	dead := onlyDead(targets)
	if len(dead) == 0 {
		fmt.Println("nada morto pra remover. Máquina limpa.")
		return nil
	}

	backupDir := filepath.Join(home, ".cts-backups", time.Now().Format("20060102-150405"))
	res, err := remove.New(backupDir, !execute).Remove(ctx, dead)
	if err != nil {
		return err
	}
	printPurge(res, backupDir)
	return nil
}

func onlyDead(targets []target.Target) []target.Target {
	var dead []target.Target
	for _, t := range targets {
		if t.Dead {
			dead = append(dead, t)
		}
	}
	return dead
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

func printPurge(res remove.Result, backupDir string) {
	for _, t := range res.Removed {
		fmt.Printf("✗ %-7s %-28s %9s\n", t.Category, t.Name, humanSize(t.SizeBytes))
	}
	verb := "removeu"
	if res.DryRun {
		verb = "removeria"
	}
	fmt.Printf("\n%s %d itens, libera %s.\n", verb, len(res.Removed), humanSize(res.FreedBytes))
	if res.DryRun {
		fmt.Println("(dry-run — nada foi apagado. Rode 'cts purge --yes' pra executar.)")
	} else {
		fmt.Printf("backup em %s\n", backupDir)
	}
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
