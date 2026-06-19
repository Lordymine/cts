// Command cts — Cut The Shit. Limpa skills, agentes, plugins e MCP mortos da máquina.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"

	"cts/internal/remove"
	"cts/internal/scan"
	"cts/internal/scan/agents"
	"cts/internal/scan/mcp"
	"cts/internal/scan/plugins"
	"cts/internal/scan/skills"
	"cts/internal/target"
	"cts/internal/ui"
)

func main() {
	args := os.Args[1:]

	var err error
	switch {
	case len(args) == 0:
		err = runMenu(context.Background())
	case args[0] == "clean":
		err = runInteractive(context.Background())
	case args[0] == "scan":
		err = runScan(context.Background())
	case args[0] == "purge":
		err = runPurge(context.Background(), len(args) > 1 && args[1] == "--yes")
	case args[0] == "help", args[0] == "-h", args[0] == "--help":
		fmt.Println(ui.Help())
	default:
		fmt.Println(ui.Help())
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "cts: %v\n", err)
		os.Exit(1)
	}
}

// runMenu mostra logo + menu e despacha a ação, em loop até "Sair".
func runMenu(ctx context.Context) error {
	for {
		fmt.Println(ui.Logo())

		var action string
		err := huh.NewSelect[string]().
			Title("O que vamos fazer?").
			Options(
				huh.NewOption("Escanear — ver o que tem (seguro)", "scan"),
				huh.NewOption("Limpar — escolher e remover", "clean"),
				huh.NewOption("Ajuda", "help"),
				huh.NewOption("Sair", "quit"),
			).
			Value(&action).
			Run()
		if err != nil {
			return ignoreAbort(err)
		}

		switch action {
		case "scan":
			if err := runScan(ctx); err != nil {
				return err
			}
		case "clean":
			if err := runInteractive(ctx); err != nil {
				return err
			}
		case "help":
			fmt.Println(ui.Help())
		case "quit":
			return nil
		}
	}
}

// buildScanners monta os scanners com os caminhos reais. IO e wiring na borda.
func buildScanners(home string) []scan.Scanner {
	return []scan.Scanner{
		skills.New(filepath.Join(home, ".claude", "skills")),
		agents.New(home, pathLister{}, agents.DefaultCatalog()),
		plugins.New(filepath.Join(home, ".claude", "plugins")),
		mcp.New(filepath.Join(home, ".claude.json"), pathLister{}.IsInstalled),
	}
}

// scanAll roda todos os scanners com os caminhos reais do usuário.
func scanAll(ctx context.Context) ([]target.Target, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	return scan.Run(ctx, buildScanners(home)...)
}

func runScan(ctx context.Context) error {
	targets, err := scanAll(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aviso: %v\n", err)
	}
	fmt.Print(ui.Report(targets))
	return nil
}

// runPurge remove SÓ os mortos. Dry-run por padrão; --yes executa (com backup).
func runPurge(ctx context.Context, execute bool) error {
	targets, err := scanAll(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aviso: %v\n", err)
	}

	dead := onlyDead(targets)
	if len(dead) == 0 {
		fmt.Println("nada morto pra remover. Máquina limpa.")
		return nil
	}

	home, _ := os.UserHomeDir()
	backupDir := filepath.Join(home, ".cts-backups", time.Now().Format("20060102-150405"))
	res, err := remove.New(backupDir, !execute, cmdRunner{}).Remove(ctx, dead)
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

// pathLister checa instalação via PATH. Adapter real do agents.Lister.
type pathLister struct{}

func (pathLister) IsInstalled(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

// cmdRunner roda comandos de verdade (npm rm -g, claude mcp remove). Adapter do
// remove.Runner, na borda — o core de remoção não sabe de exec.
type cmdRunner struct{}

func (cmdRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func printPurge(res remove.Result, backupDir string) {
	for _, t := range res.Removed {
		fmt.Printf("✗ %-7s %-28s %9s\n", t.Category, t.Name, ui.HumanSize(t.SizeBytes))
	}
	verb := "removeu"
	if res.DryRun {
		verb = "removeria"
	}
	fmt.Printf("\n%s %d itens, libera %s.\n", verb, len(res.Removed), ui.HumanSize(res.FreedBytes))
	if res.DryRun {
		fmt.Println("(dry-run — nada foi apagado. Rode 'cts purge --yes' pra executar.)")
	} else {
		fmt.Printf("backup em %s\n", backupDir)
	}
}
