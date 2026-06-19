package ui

import (
	"strings"
	"testing"

	"cts/internal/target"
)

func TestReportAgrupaEContaMortos(t *testing.T) {
	targets := []target.Target{
		{Name: "viva", Category: target.Skill, SizeBytes: 1024},
		{Name: "morta", Category: target.Agent, Dead: true, Reason: "config órfã"},
	}
	out := Report(targets)
	for _, want := range []string{"viva", "morta", "SKILL", "AGENT", "2 alvos", "1 mortos"} {
		if !strings.Contains(out, want) {
			t.Errorf("Report deveria conter %q; saída:\n%s", want, out)
		}
	}
}

func TestReportVazio(t *testing.T) {
	if out := Report(nil); !strings.Contains(out, "limpa") {
		t.Errorf("Report(nil) deveria dizer que está limpa, veio %q", out)
	}
}

func TestHumanSize(t *testing.T) {
	cases := map[int64]string{0: "0B", 512: "512B", 1024: "1.0KB", 1048576: "1.0MB"}
	for in, want := range cases {
		if got := HumanSize(in); got != want {
			t.Errorf("HumanSize(%d)=%q, queria %q", in, got, want)
		}
	}
}
