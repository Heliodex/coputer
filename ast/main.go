package main

import (
	"encoding/json"
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

type Node struct {
	Type     string  `json:"type"`
	Location string  `json:"location"`
	HasEnd   *bool   `json:"hasEnd"`
	Body     *[]Node `json:"body"`
	Expr     *Node   `json:"expr"`
	Func     *Node   `json:"func"`
	Args     *[]Node `json:"args"`
}

func (n Node) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Type      %s\n", n.Type))
	b.WriteString(fmt.Sprintf("Location  %s\n", n.Location))
	if n.HasEnd != nil {
		b.WriteString(fmt.Sprintf("HasEnd    %t\n", *n.HasEnd))
	}

	if n.Body != nil {
		b.WriteString("Body:\n")
		for _, c := range *n.Body {
			b.WriteString(indentStart(c.String(), 4))
			b.WriteByte('\n')
		}
	}

	if n.Expr != nil {
		b.WriteString("Expr:\n")
		b.WriteString(indentStart(n.Expr.String(), 4))
		b.WriteByte('\n')
	}

	if n.Func != nil {
		b.WriteString("Func:\n")
		b.WriteString(indentStart(n.Func.String(), 4))
		b.WriteByte('\n')
	}

	if n.Args != nil {
		b.WriteString("Args:\n")
		for _, arg := range *n.Args {
			b.WriteString(indentStart(arg.String(), 4))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

type Comment struct {
	Type     string `json:"type"`
	Location string `json:"location"`
}

func (c Comment) String() string {
	return fmt.Sprintf("Type  %12s  Location  %s\n", c.Type, c.Location)
}

type AST struct {
	Root             Node      `json:"root"`
	CommentLocations []Comment `json:"commentLocations"`
}

func (ast AST) String() string {
	var b strings.Builder

	b.WriteString("Root:\n")
	b.WriteString(indentStart(ast.Root.String(), 4))
	b.WriteString("\n\n")

	b.WriteString("Comment Locations:\n")
	for _, c := range ast.CommentLocations {
		b.WriteString(indentStart(c.String(), 4))
		b.WriteByte('\n')
	}

	return b.String()
}

func main() {
	const filepath = "../test/ast/hello.luau"

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
	var ast AST
	err = json.Unmarshal(output, &ast)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return
	}

	fmt.Println(ast)
}
