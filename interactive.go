package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/huh"

	"cts/internal/remove"
	"cts/internal/scan"
	"cts/internal/target"
)

// runInteractive: scan → lista interativa (mortos pré-marcados) → confirma → remove com backup.
// É a camada de UI; toda a lógica de remoção fica no core testado (internal/remove).
func runInteractive(ctx context.Context) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}
	targets, err := scan.Run(ctx, buildScanners(home)...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "aviso: %v\n", err)
	}
	if len(targets) == 0 {
		fmt.Println("nada encontrado. Máquina limpa.")
		return nil
	}

	chosen, err := selectTargets(targets)
	if err != nil {
		return ignoreAbort(err)
	}
	if len(chosen) == 0 {
		fmt.Println("nada selecionado.")
		return nil
	}

	ok, err := confirmRemove(chosen)
	if err != nil {
		return ignoreAbort(err)
	}
	if !ok {
		fmt.Println("cancelado.")
		return nil
	}

	backupDir := filepath.Join(home, ".cts-backups", time.Now().Format("20060102-150405"))
	res, err := remove.New(backupDir, false).Remove(ctx, chosen)
	if err != nil {
		return err
	}
	printPurge(res, backupDir)
	return nil
}

// selectTargets mostra a multiselect. O valor é o índice — Target tem slice
// (Paths), logo não é comparável, e huh exige valor comparável.
func selectTargets(targets []target.Target) ([]target.Target, error) {
	opts := make([]huh.Option[int], len(targets))
	for i, t := range targets {
		opts[i] = huh.NewOption(label(t), i).Selected(t.Dead)
	}

	var picked []int
	err := huh.NewMultiSelect[int]().
		Title("cts — marque o que remover (mortos já vêm marcados). Espaço seleciona, Enter confirma.").
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
		Title(fmt.Sprintf("Remover %d itens (libera %s)? Backup será feito antes.", len(chosen), humanSize(freed))).
		Affirmative("Remover").
		Negative("Cancelar").
		Value(&ok).
		Run()
	return ok, err
}

func label(t target.Target) string {
	mark := " "
	if t.Dead {
		mark = "✗"
	}
	reason := ""
	if t.Reason != "" {
		reason = "— " + t.Reason
	}
	return fmt.Sprintf("%s %-7s %-28s %9s %s", mark, t.Category, t.Name, humanSize(t.SizeBytes), reason)
}

// ignoreAbort transforma Ctrl+C / Esc do huh em saída limpa, não em erro.
func ignoreAbort(err error) error {
	if errors.Is(err, huh.ErrUserAborted) {
		fmt.Println("cancelado.")
		return nil
	}
	return err
}
