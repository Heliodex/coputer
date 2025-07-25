package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

const (
	Ext    = ".luau"
	astDir = "../test/ast"
)

func trimext(s string) string {
	return strings.TrimSuffix(s, Ext)
}

func TestAST(t *testing.T) {
	files, err := os.ReadDir(astDir)
	if err != nil {
		t.Fatal("error reading AST tests directory:", err)
	}

	for _, f := range files {
		fn := f.Name()
		if !strings.HasSuffix(fn, Ext) {
			continue
		}
		name := trimext(fn)

		t.Log(" -- Testing", name, "--")
		filename := fmt.Sprintf("%s/%s", astDir, name)

		output, err := luauAst(filename + Ext)
		if err != nil {
			t.Fatal("error running luau-ast:", err)
		}

		// Decode the AST
		ast, err := DecodeAST(output)
		if err != nil {
			t.Fatal("error decoding AST:", err)
		}
		o := ast.String()

		ogb, err := os.ReadFile(filename + ".txt")
		if err != nil {
			t.Fatal("error reading expected output:", err)
		}
		og := string(ogb)

		if o != og {
			t.Errorf("output mismatch:\n-- Expected\n%s\n-- Got\n%s\n", og, o)
			fmt.Println()

			// print mismatch
			oLines, ogLines := strings.Split(o, "\n"), strings.Split(og, "\n")
			olen, oglen := len(oLines), len(ogLines)

			if olen != oglen {
				t.Errorf("line count mismatch: expected %d, got %d", oglen, olen)
			}

			for i := range max(olen, oglen) {
				if i >= olen || i >= oglen {
					continue
				}

				if oline, ogline := oLines[i], ogLines[i]; oline != ogline {
					t.Errorf("mismatched line: \n%s\n%v\n%s\n%v\n", oline, []byte(oline), ogline, []byte(ogline))
				}
			}

			os.Exit(1)
		}
	}
}
