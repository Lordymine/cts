// Package configroots resolves the OS-specific base directories where tool
// configuration lives, so scanners find configs on Linux, macOS and Windows.
package configroots

import (
	"os"
	"path/filepath"
	"runtime"
)

// Root is a base directory where tool configs live. Dotted is true when tools
// store config as a hidden ".<name>" entry (the home directory); false when the
// entry is a plain "<name>" (XDG ~/.config, Windows AppData, ~/Library).
type Root struct {
	Path   string
	Dotted bool
}

// Roots returns the config roots for the current OS.
func Roots() []Root {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	return rootsFor(runtime.GOOS, home)
}

// rootsFor is the pure logic, testable per OS without running on that OS.
func rootsFor(goos, home string) []Root {
	roots := []Root{
		{Path: home, Dotted: true},
		{Path: filepath.Join(home, ".config"), Dotted: false},
	}
	switch goos {
	case "windows":
		roots = append(roots,
			Root{Path: filepath.Join(home, "AppData", "Roaming"), Dotted: false},
			Root{Path: filepath.Join(home, "AppData", "Local"), Dotted: false},
		)
	case "darwin":
		roots = append(roots, Root{Path: filepath.Join(home, "Library", "Application Support"), Dotted: false})
	}
	return roots
}

// Entry builds the candidate path for a tool's config dir under this root,
// adding the "." prefix on the home root.
func (r Root) Entry(name string) string {
	if r.Dotted {
		name = "." + name
	}
	return filepath.Join(r.Path, name)
}
