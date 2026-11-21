package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestFile writes content to a file named name inside dir, creating it with the given permissions, and returns its path.
func writeTestFile(t *testing.T, dir, name, content string, mode os.FileMode) string {
	t.Helper()

	path := filepath.Join(dir, name)

	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal("error writing test file:", err)
	}

	if err := os.Chmod(path, mode); err != nil {
		t.Fatal("error changing test file mode:", err)
	}

	return path
}

// readTestFile reports the contents of the file at path.
func readTestFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("error reading test file:", err)
	}

	return string(content)
}

// TestFormatDir checks that every .luau file inside a directory, including those in its subdirectories, is formatted in place, leaving other files alone.
func TestFormatDir(t *testing.T) {
	dir := t.TempDir()

	unformatted := writeTestFile(t, dir, "a.luau", "local x=1\n", 0o644)
	formatted := writeTestFile(t, dir, "b.luau", "local y = 2\n", 0o644)
	other := writeTestFile(t, dir, "notes.txt", "local x=1\n", 0o644)

	nestedDir := filepath.Join(dir, "nested")
	if err := os.Mkdir(nestedDir, 0o755); err != nil {
		t.Fatal("error making test directory:", err)
	}

	nested := writeTestFile(t, nestedDir, "c.luau", "local z=3\n", 0o644)

	count, err := formatDir(dir)
	if err != nil {
		t.Fatal("error formatting directory:", err)
	}

	if count != 2 {
		t.Errorf("expected 2 files changed, got %d", count)
	}

	if got, want := readTestFile(t, unformatted), "local x = 1\n"; got != want {
		t.Errorf("unexpected formatted file, expected %q, got %q", want, got)
	}

	if got, want := readTestFile(t, formatted), "local y = 2\n"; got != want {
		t.Errorf("expected an already formatted file to be unchanged, expected %q, got %q", want, got)
	}

	if got, want := readTestFile(t, other), "local x=1\n"; got != want {
		t.Errorf("expected a file that isn't Luau to be unchanged, expected %q, got %q", want, got)
	}

	if got, want := readTestFile(t, nested), "local z = 3\n"; got != want {
		t.Errorf("expected a file in a subdirectory to be formatted, expected %q, got %q", want, got)
	}
}

// TestFormatDirKeepsMode checks that formatting a file in place preserves its permissions.
func TestFormatDirKeepsMode(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "private.luau", "local x=1\n", 0o600)

	if _, err := formatDir(dir); err != nil {
		t.Fatal("error formatting directory:", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal("error getting file info:", err)
	}

	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("expected permissions 0600, got %04o", mode)
	}
}

// TestFormatDirCarriesOnPastErrors checks that a file that fails to parse doesn't stop the rest of the directory from being formatted.
func TestFormatDirCarriesOnPastErrors(t *testing.T) {
	dir := t.TempDir()

	broken := writeTestFile(t, dir, "broken.luau", "local = 1\n", 0o644)
	unformatted := writeTestFile(t, dir, "a.luau", "local x=1\n", 0o644)

	count, err := formatDir(dir)
	if err == nil {
		t.Fatal("expected an error for the file that failed to parse")
	}

	if !strings.Contains(err.Error(), "broken.luau") {
		t.Errorf("expected the error to name the file, got %q", err)
	}

	if count != 1 {
		t.Errorf("expected 1 file changed, got %d", count)
	}

	if got, want := readTestFile(t, unformatted), "local x = 1\n"; got != want {
		t.Errorf("unexpected formatted file, expected %q, got %q", want, got)
	}

	if got, want := readTestFile(t, broken), "local = 1\n"; got != want {
		t.Errorf("expected the file that failed to parse to be unchanged, expected %q, got %q", want, got)
	}
}

// TestRunPaths checks the command line handling: a single path is accepted, more than one is rejected, and -w needs a path to write to.
func TestRunPaths(t *testing.T) {
	dir := t.TempDir()
	path := writeTestFile(t, dir, "a.luau", "local x=1\n", 0o644)

	if err := run([]string{path}, true); err != nil {
		t.Fatal("error formatting a file in place:", err)
	}

	if got, want := readTestFile(t, path), "local x = 1\n"; got != want {
		t.Errorf("unexpected formatted file, expected %q, got %q", want, got)
	}

	if err := run([]string{"a.luau", "b.luau"}, true); err == nil {
		t.Error("expected an error for more than one path")
	}

	if err := run(nil, true); err == nil {
		t.Error("expected an error for -w without a path")
	}

	if err := run([]string{dir}, false); err == nil {
		t.Error("expected an error for printing a directory")
	}
}
