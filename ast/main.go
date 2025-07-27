package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tailscale/hujson"
)

const (
	Ext            = ".luau"
	astDir         = "../test/ast"
	benchmarkDir   = "../test/benchmark"
	conformanceDir = "../test/conformance"
)

func luauAst(path string) (output []byte, err error) {
	cmd := exec.Command("luau-ast", path)
	return cmd.Output()
}

func indentStart(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", n) + line
	}
	return strings.Join(lines, "\n")
}

// remember, luau-ast outputs JSONC, not JSON
func standardise(in []byte) []byte {
	v, err := hujson.Parse(in)
	if err != nil {
		return in
	}
	v.Standardize()
	v.Format()
	return v.Pack()
}

func main() {
	const filepath = astDir + "/functiongeneric.luau"

	out, err := luauAst(filepath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	s := standardise(out)

	// pprof time
	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	fmt.Println("Error creating CPU profile file:", err)
	// 	return
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// write to ast.jsonc
	if err = os.WriteFile("ast.jsonc", s, 0o644); err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	fmt.Println("AST written to ast.jsonc successfully.")

	st := time.Now()

	// encode as AST
	ast, err := DecodeAST(s)
	if err != nil {
		fmt.Println("Error decoding AST:", err)
		return
	}

	fmt.Println(ast)
	fmt.Printf("AST decoded in %s\n", time.Since(st))
}
