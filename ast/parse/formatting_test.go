package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// compareSource reports a mismatch between a formatted source and its expected file, including a line-by-line diff.
func compareSource(t *testing.T, name, got, expected string) bool {
	t.Helper()

	if got == expected {
		return true
	}

	t.Errorf("%s: output mismatch:\n-- Expected\n%s\n-- Got\n%s\n", name, expected, got)

	gotLines, expectedLines := strings.Split(got, "\n"), strings.Split(expected, "\n")
	gotLen, expectedLen := len(gotLines), len(expectedLines)

	if gotLen != expectedLen {
		t.Errorf("%s: line count mismatch: expected %d, got %d", name, expectedLen, gotLen)
	}

	for i := range max(gotLen, expectedLen) {
		if i >= gotLen || i >= expectedLen {
			continue
		}

		if gotLine, expectedLine := gotLines[i], expectedLines[i]; gotLine != expectedLine {
			t.Errorf("%s: mismatched line %d, expected:\n%s\n%v\ngot:\n%s\n%v\n", name, i+1, expectedLine, []byte(expectedLine), gotLine, []byte(gotLine))
		}
	}

	return false
}

// TestFormatting parses each {name}.luau file in test/formatting and checks that its Source() output matches the expected {name}_out.luau file.
func TestFormatting(t *testing.T) {
	files, err := os.ReadDir("../" + FormattingDir)
	if err != nil {
		t.Fatal("error reading formatting tests directory:", err)
	}

	for _, f := range files {
		fn := f.Name()
		if !strings.HasSuffix(fn, Ext) || strings.HasSuffix(fn, FormattingOutSuffix+Ext) || strings.HasSuffix(fn, MinifiedOutSuffix+Ext) {
			continue
		}
		name := trimext(fn)

		t.Log(" -- Testing", name, "--")
		filename := fmt.Sprintf("../%s/%s", FormattingDir, name)

		content, err := os.ReadFile(filename + Ext)
		if err != nil {
			t.Fatal("error reading test file:", err)
		}

		ok, res := Parse(string(content), Options{})
		if !ok {
			t.Errorf("%s: error parsing file: %v", name, res.Errors)
			continue
		}

		// formatted files are expected to end with a newline
		got := res.Root.Source() + "\n"

		expectedBytes, err := os.ReadFile(filename + FormattingOutSuffix + Ext)
		if err != nil {
			t.Fatal("error reading expected output:", err)
		}
		expected := strings.ReplaceAll(string(expectedBytes), "\r\n", "\n")

		if !compareSource(t, name, got, expected) {
			continue
		}

		// the expected output should itself be valid Luau
		if ok, res := Parse(got, Options{}); !ok {
			t.Errorf("%s: formatted output failed to parse: %v", name, res.Errors)
		}
	}
}

// TestFormattingIdempotent checks that parsing already-formatted files and rendering them again reproduces them exactly.
// If you want this except for all existing test files, see [TestSourceRoundTrip].
func TestFormattingIdempotent(t *testing.T) {
	files, err := os.ReadDir("../" + FormattingDir)
	if err != nil {
		t.Fatal("error reading formatting tests directory:", err)
	}

	for _, f := range files {
		fn := f.Name()
		if !strings.HasSuffix(fn, FormattingOutSuffix+Ext) {
			continue
		}
		name := strings.TrimSuffix(trimext(fn), FormattingOutSuffix)

		t.Log(" -- Testing", name, "--")
		filename := fmt.Sprintf("../%s/%s%s", FormattingDir, name, FormattingOutSuffix)

		content, err := os.ReadFile(filename + Ext)
		if err != nil {
			t.Fatal("error reading formatted file:", err)
		}
		expected := strings.ReplaceAll(string(content), "\r\n", "\n")

		ok, res := Parse(expected, Options{})
		if !ok {
			t.Errorf("%s: error parsing formatted file: %v", name, res.Errors)
			continue
		}

		compareSource(t, name, res.Root.Source()+"\n", expected)
	}
}
