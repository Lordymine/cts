package dirsize

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOfSomaArquivos(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a"), "12345")      // 5 bytes
	writeFile(t, filepath.Join(dir, "sub", "b"), "123") // 3 bytes
	if got := Of(dir); got != 8 {
		t.Fatalf("Of=%d, queria 8", got)
	}
}

func TestOfSegueSymlinkRaiz(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real")
	writeFile(t, filepath.Join(real, "f"), "1234") // 4 bytes

	link := filepath.Join(dir, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlink não suportado neste ambiente: %v", err)
	}
	if got := Of(link); got != 4 {
		t.Fatalf("Of(symlink)=%d, queria 4 — deve seguir o link", got)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
