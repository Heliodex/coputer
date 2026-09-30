package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestFormatting parses each {name}.luau file in test/formatting and checks
// that its Source() output matches the expected {name}_out.luau file.
func TestFormatting(t *testing.T) {
	files, err := os.ReadDir("../" + FormattingDir)
	if err != nil {
		t.Fatal("error reading formatting tests directory:", err)
	}

	for _, f := range files {
		fn := f.Name()
		if !strings.HasSuffix(fn, Ext) || strings.HasSuffix(fn, FormattingOutSuffix+Ext) {
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
		o := res.Root.Source() + "\n"

		ogb, err := os.ReadFile(filename + FormattingOutSuffix + Ext)
		if err != nil {
			t.Fatal("error reading expected output:", err)
		}
		og := strings.ReplaceAll(string(ogb), "\r\n", "\n")

		if o != og {
			t.Errorf("%s: output mismatch:\n-- Expected\n%s\n-- Got\n%s\n", name, og, o)

			oLines, ogLines := strings.Split(o, "\n"), strings.Split(og, "\n")
			olen, oglen := len(oLines), len(ogLines)

			if olen != oglen {
				t.Errorf("%s: line count mismatch: expected %d, got %d", name, oglen, olen)
			}

			for i := range max(olen, oglen) {
				if i >= olen || i >= oglen {
					continue
				}

				if oline, ogline := oLines[i], ogLines[i]; oline != ogline {
					t.Errorf("%s: mismatched line %d, expected:\n%s\n%v\ngot:\n%s\n%v\n", name, i+1, ogline, []byte(ogline), oline, []byte(oline))
				}
			}

			continue
		}

		// the expected output should itself be valid Luau
		if ok, res := Parse(o, Options{}); !ok {
			t.Errorf("%s: formatted output failed to parse: %v", name, res.Errors)
		}
	}
}
