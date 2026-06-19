package configroots

import (
	"path/filepath"
	"testing"
)

func TestRootsForWindows(t *testing.T) {
	roots := rootsFor("windows", "C:/Users/u")
	if len(roots) != 4 {
		t.Fatalf("windows: want 4 roots (home, .config, Roaming, Local), got %d", len(roots))
	}
	if !roots[0].Dotted {
		t.Error("home root should be dotted")
	}
	if roots[1].Dotted {
		t.Error(".config root should not be dotted")
	}
}

func TestRootsForDarwin(t *testing.T) {
	roots := rootsFor("darwin", "/Users/u")
	if len(roots) != 3 {
		t.Fatalf("darwin: want 3 roots (home, .config, Library), got %d", len(roots))
	}
}

func TestRootsForLinux(t *testing.T) {
	roots := rootsFor("linux", "/home/u")
	if len(roots) != 2 {
		t.Fatalf("linux: want 2 roots (home, .config), got %d", len(roots))
	}
}

func TestEntry(t *testing.T) {
	dotted := Root{Path: "/home/u", Dotted: true}
	if got, want := dotted.Entry("qwen"), filepath.Join("/home/u", ".qwen"); got != want {
		t.Errorf("dotted Entry = %q, want %q", got, want)
	}
	plain := Root{Path: "/home/u/.config", Dotted: false}
	if got, want := plain.Entry("qwen"), filepath.Join("/home/u/.config", "qwen"); got != want {
		t.Errorf("plain Entry = %q, want %q", got, want)
	}
}
