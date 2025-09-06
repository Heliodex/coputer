package main

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/Heliodex/coputer/ast/ast"
)

const LuauExt = ".luau"

func processFile(filepath string) error {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	out, err := LuauAst(filepath)
	if err != nil {
		return fmt.Errorf("error converting to Luau AST: %w", err)
	}

	// encode as AST
	ast, err := DecodeAST(out)
	if err != nil {
		return fmt.Errorf("error decoding AST: %w", err)
	}

	fmt.Printf("AST: %+v\n", ast)

	newsource, err := ast.Source(string(content))
	if err != nil {
		return fmt.Errorf("error encoding AST: %w", err)
	}
	// fmt.Println(new)

	err = os.WriteFile(filepath, []byte(newsource), 0o644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: <path>")
		return
	}

	filearg := os.Args[1]

	info, err := os.Stat(filearg)
	if err != nil {
		fmt.Println("Error getting file stat:", err)
		return
	}

	var files []string
	if info.IsDir() {
		dirEntries, err := os.ReadDir(filearg)
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}

		for _, entry := range dirEntries {
			if entry.IsDir() {
				continue
			}

			ext := filepath.Ext(entry.Name())
			if ext != LuauExt {
				continue
			}

			files = append(files, filearg+"/"+entry.Name())
		}
	} else {
		if filepath.Ext(filearg) != LuauExt {
			fmt.Println("File is not a .luau file")
			return
		}
		files = append(files, filearg)
	}

	for _, file := range files {
		fmt.Println("Processing file", file)
		err := processFile(file)
		if err != nil {
			fmt.Println("Error processing file:", err)
		}
	}
}
