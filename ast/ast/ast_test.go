package ast

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func trimext(s string) string {
	return strings.TrimSuffix(s, Ext)
}

func TestAST(t *testing.T) {
	files, err := os.ReadDir("../" + AstDir)
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
		filename := fmt.Sprintf("../%s/%s", AstDir, name)

		out, err := LuauAst(filename + Ext)
		if err != nil {
			t.Fatal("error running luau-ast:", err)
		}

		// Decode the AST
		ast, err := DecodeAST(out)
		if err != nil {
			t.Fatal("error decoding AST:", err)
		}
		o := ast.String()

		// write to file
		// if err = os.WriteFile(filename+".txt", []byte(o), 0o644); err != nil {
		// 	t.Fatal("error writing output file:", err)
		// }

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
					t.Errorf("mismatched line, expected:\n%s\n%v\ngot:\n%s\n%v\n", ogline, []byte(ogline), oline, []byte(oline))
				}
			}

			os.Exit(1)
		}
	}
}

func parseFile(t *testing.T, f os.DirEntry, dir string) {
	fn := f.Name()
	if !strings.HasSuffix(fn, Ext) {
		return
	}
	name := trimext(fn)

	t.Log(" -- Testing", name, "--")
	filename := fmt.Sprintf("../%s/%s", dir, name)

	if name == "luauception" {
		fmt.Println("⚠️ WARNING! ⚠️ This test takes about a minute to run. It will also eat all of your RAM.")
	}

	out, err := LuauAst(filename + Ext)
	if err != nil {
		t.Fatal("error running luau-ast:", err)
	}

	fmt.Println("luau-ast completed")
	st := time.Now()

	// Decode the AST
	if _, err = DecodeAST(out); err != nil {
		t.Fatal("error decoding AST:", err)
	}

	fmt.Println("decoded in", time.Since(st))
}

func TestParsing(t *testing.T) {
	files1, err := os.ReadDir("../" + BenchmarkDir)
	if err != nil {
		t.Fatal("error reading benchmark tests directory:", err)
	}

	files2, err := os.ReadDir("../" + ConformanceDir)
	if err != nil {
		t.Fatal("error reading conformance tests directory:", err)
	}

	for _, f := range files1 {
		parseFile(t, f, BenchmarkDir)
	}

	for _, f := range files2 {
		parseFile(t, f, ConformanceDir)
	}
}

func TestSource(t *testing.T) {
	files, err := os.ReadDir("../" + AstDir)
	if err != nil {
		t.Fatal("error reading benchmark tests directory:", err)
	}

	for _, f := range files {
		fn := f.Name()
		if !strings.HasSuffix(fn, Ext) {
			continue
		}
		name := trimext(fn)

		t.Log(" -- Testing", name, "--")
		filename := fmt.Sprintf("../%s/%s", AstDir, name)

		content, err := os.ReadFile(filename + Ext)
		if err != nil {
			t.Fatal("error reading test file:", err)
		}

		out, err := LuauAst(filename + Ext)
		if err != nil {
			t.Fatal("error running luau-ast:", err)
		}

		// Decode the AST
		ast, err := DecodeAST(out)
		if err != nil {
			t.Fatal("error decoding AST:", err)
		}
		src, err := ast.Source(string(content))
		if err != nil {
			t.Fatal("error getting source from AST:", err)
		}

		fmt.Println(src)
	}
}
