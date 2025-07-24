package main

import (
	"fmt"
	"os/exec"
	"strings"
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

func main() {
	const filepath = "../test/ast/tableaccess.luau"

	output, err := luauAst(filepath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// write to ast.json
	// if err = os.WriteFile("ast.json", output, 0o644); err != nil {
	// 	fmt.Println("Error writing to file:", err)
	// 	return
	// }
	// fmt.Println("AST written to ast.json successfully.")

	// encode as AST
	ast, err := DecodeAST(output)
	if err != nil {
		fmt.Println("Error decoding AST:", err)
		return
	}

	fmt.Println(ast)
}
