package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/Lordymine/cts/internal/remove"
	"github.com/Lordymine/cts/internal/scan"
	"github.com/Lordymine/cts/internal/target"
	"github.com/Lordymine/cts/internal/ui"
)

// runInteractive: scan → interactive list (dead pre-checked) → confirm → remove
// with backup. This is the UI layer; all removal logic lives in the tested core
// (internal/remove).
func runInteractive(ctx context.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}
	targets, err := scan.Run(ctx, buildScanners(home)...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}
	if len(targets) == 0 {
		fmt.Println("nothing found. Machine is clean.")
		return nil
	}

	chosen, err := selectTargets(targets)
	if err != nil {
		return ignoreAbort(err)
	}
	if len(chosen) == 0 {
		fmt.Println("nothing selected.")
		return nil
	}

	ok, err := confirmRemove(chosen)
	if err != nil {
		return ignoreAbort(err)
	}
	if !ok {
		fmt.Println("cancelled.")
		return nil
	}

	backupDir := filepath.Join(home, ".cts-backups", time.Now().Format("20060102-150405"))
	res, err := remove.New(backupDir, false, cmdRunner{}).Remove(ctx, chosen)
	if err != nil {
		return err
	}
	printPurge(res, backupDir)
	return nil
}

// selectTargets shows the multi-select. The value is the index — Target has a
// slice (Paths), so it is not comparable, and huh requires a comparable value.
func selectTargets(targets []target.Target) ([]target.Target, error) {
	opts := make([]huh.Option[int], len(targets))
	for i, t := range targets {
		opts[i] = huh.NewOption(label(t), i).Selected(t.Dead)
	}

	var picked []int
	err := huh.NewMultiSelect[int]().
		Title("cts — mark what to remove (dead is pre-checked). Space selects, Enter confirms.").
		Options(opts...).
		Value(&picked).
		Run()
	if err != nil {
		return nil, err
	}

	chosen := make([]target.Target, 0, len(picked))
	for _, i := range picked {
		chosen = append(chosen, targets[i])
	}
	return chosen, nil
}

func confirmRemove(chosen []target.Target) (bool, error) {
	var freed int64
	for _, t := range chosen {
		freed += t.SizeBytes
	}
	var ok bool
	err := huh.NewConfirm().
		Title(fmt.Sprintf("Remove %d items (frees %s)? A backup is made first.", len(chosen), ui.HumanSize(freed))).
		Affirmative("Remove").
		Negative("Cancel").
		Value(&ok).
		Run()
	return ok, err
}

func label(t target.Target) string {
	mark := " "
	if t.Dead {
		mark = "x"
	}
	reason := ""
	if t.Reason != "" {
		reason = "- " + t.Reason
	}
	return fmt.Sprintf("%s %-7s %-28s %9s %s", mark, t.Category, t.Name, ui.HumanSize(t.SizeBytes), reason)
}

// ignoreAbort turns huh's Ctrl+C / Esc into a clean exit, not an error.
func ignoreAbort(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		fmt.Println("cancelled.")
		return nil
	}
	return err
}
