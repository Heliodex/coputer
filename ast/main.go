package main

import (
	"fmt"
	"os"
	"time"

	. "github.com/Heliodex/coputer/ast/ast"
)

func main() {
	const filepath = AstDir + "/typeoptionals.luau"

	out, err := LuauAst(filepath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// no more standardisation, yayyy

	// pprof time
	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	fmt.Println("Error creating CPU profile file:", err)
	// 	return
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// write to ast.jsonc
	if err = os.WriteFile("ast.jsonc", out, 0o644); err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
	fmt.Println("AST written to ast.jsonc successfully.")

	st := time.Now()

	// encode as AST
	ast, err := DecodeAST(out)
	if err != nil {
		fmt.Println("Error decoding AST:", err)
		return
	}

	fmt.Println(ast)
	fmt.Printf("AST decoded in %s\n", time.Since(st))
}
