package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGlobDoubleStarRecurses(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"z.txt":     "z\n",
		"a/x.txt":   "x\n",
		"a/b/y.txt": "y\n",
		"a/skip.md": "no\n",
	}

	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	got := run(t, "", "cat", filepath.ToSlash(dir)+"/**/*.txt").stdout

	if want := "y\nx\nz\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
