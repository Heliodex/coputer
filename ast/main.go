package main

import (
	"fmt"
	"os"

	. "github.com/Heliodex/coputer/ast/ast"
)

func main() {
	const filepath = AstDir + "/typepack.luau"

	content, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	source := string(content)

	out, err := LuauAst(filepath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// no more standardisation, yayyy

	// write out to ast.json
	// err = os.WriteFile("ast.json", out, 0o644)
	// if err != nil {
	// 	fmt.Println("Error writing AST to file:", err)
	// 	return
	// }

	// pprof time
	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	fmt.Println("Error creating CPU profile file:", err)
	// 	return
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// st := time.Now()

	fmt.Println(source)

	// encode as AST
	ast, err := DecodeAST(out)
	if err != nil {
		fmt.Println("Error decoding AST:", err)
		return
	}

	fmt.Println(ast)

	new, err := ast.Source(source)
	if err != nil {
		fmt.Println("Error getting source from AST:", err)
		return
	}
	fmt.Println(new)
	// fmt.Printf("AST decoded in %s\n", time.Since(st))
}
