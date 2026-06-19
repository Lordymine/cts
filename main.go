// Command cts — Cut The Shit. Cleans dead skills, agents, plugins and MCP servers.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"

	"cts/internal/configroots"
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

// runMenu shows the logo + menu and dispatches the chosen action, looping until "Quit".
func runMenu(ctx context.Context) error {
	for {
		fmt.Println(ui.Logo())

		var action string
		err := huh.NewSelect[string]().
			Title("What do you want to do?").
			Options(
				huh.NewOption("Scan — see what's there (safe)", "scan"),
				huh.NewOption("Clean — pick items and remove", "clean"),
				huh.NewOption("Help", "help"),
				huh.NewOption("Quit", "quit"),
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

// buildScanners wires the scanners with real paths. IO and wiring live at the edge.
func buildScanners(home string) []scan.Scanner {
	return []scan.Scanner{
		skills.New(filepath.Join(home, ".claude", "skills")),
		agents.New(configroots.Roots(), pathLister{}, agents.DefaultCatalog()),
		plugins.New(filepath.Join(home, ".claude", "plugins")),
		mcp.New(filepath.Join(home, ".claude.json"), pathLister{}.IsInstalled),
	}
}

// scanAll runs every scanner against the user's real paths.
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
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}
	fmt.Print(ui.Report(targets))
	return nil
}

// runPurge removes ONLY dead targets. Dry-run by default; --yes executes (with backup).
func runPurge(ctx context.Context, execute bool) error {
	targets, err := scanAll(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	dead := onlyDead(targets)
	if len(dead) == 0 {
		fmt.Println("nothing dead to remove. Machine is clean.")
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

// pathLister checks installation via PATH. Real adapter of agents.Lister.
type pathLister struct{}

func (pathLister) IsInstalled(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

// cmdRunner runs real commands (npm rm -g, claude mcp remove). Adapter of
// remove.Runner, at the edge — the removal core knows nothing about exec.
type cmdRunner struct{}

func (cmdRunner) Run(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func printPurge(res remove.Result, backupDir string) {
	for _, t := range res.Removed {
		fmt.Printf("x %-7s %-28s %9s\n", t.Category, t.Name, ui.HumanSize(t.SizeBytes))
	}
	verb := "removed"
	if res.DryRun {
		verb = "would remove"
	}
	fmt.Printf("\n%s %d items, frees %s.\n", verb, len(res.Removed), ui.HumanSize(res.FreedBytes))
	if res.DryRun {
		fmt.Println("(dry-run — nothing was deleted. Run 'cts purge --yes' to execute.)")
	} else {
		fmt.Printf("backup at %s\n", backupDir)
	}
}
