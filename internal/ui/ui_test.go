package ui

import (
	"strings"
	"testing"

	"cts/internal/target"
)

func TestReportGroupsAndCountsDead(t *testing.T) {
	targets := []target.Target{
		{Name: "alive", Category: target.Skill, SizeBytes: 1024},
		{Name: "gone", Category: target.Agent, Dead: true, Reason: "orphan config"},
	}
	out := Report(targets)
	for _, want := range []string{"alive", "gone", "SKILL", "AGENT", "2 targets", "1 dead"} {
		if !strings.Contains(out, want) {
			t.Errorf("Report should contain %q; output:\n%s", want, out)
		}
	}
}

func TestReportEmpty(t *testing.T) {
	if out := Report(nil); !strings.Contains(out, "clean") {
		t.Errorf("Report(nil) should say the machine is clean, got %q", out)
	}
}

func TestHumanSize(t *testing.T) {
	cases := map[int64]string{0: "0B", 512: "512B", 1024: "1.0KB", 1048576: "1.0MB"}
	for in, want := range cases {
		if got := HumanSize(in); got != want {
			t.Errorf("HumanSize(%d)=%q, want %q", in, got, want)
		}
	}
}
