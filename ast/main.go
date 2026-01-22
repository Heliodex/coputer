package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	. "github.com/Heliodex/coputer/ast/ast"
)

const LuauExt = ".luau"

func processFile(filepath string, stdout bool) error {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	out, err := LuauAst(filepath)
	if err != nil {
		return fmt.Errorf("convert to Luau AST: %w", err)
	}

	// write to ast.json
	// err = os.WriteFile("ast.json", out, 0o644)
	// if err != nil {
	// 	return fmt.Errorf("write ast.json: %w", err)
	// }

	// encode as AST
	ast, err := DecodeAST(out)
	if err != nil {
		return fmt.Errorf("decode AST: %w", err)
	}

	// fmt.Println(ast)

	newsource, err := ast.Source(string(content))
	if err != nil {
		return fmt.Errorf("encode AST: %w", err)
	}
	// fmt.Println(new)

	if stdout {
		fmt.Print(newsource)
		return nil
	}

	if err = os.WriteFile(filepath, []byte(newsource), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func cmdFile(filearg string) error {
	info, err := os.Stat(filearg)
	if err != nil {
		// fmt.Println("Error getting file stat:", err)
		return fmt.Errorf("get file stat: %w", err)
	}

	if info.IsDir() {
		// fmt.Println("Provided path is a directory, use 'dir' command instead")
		return fmt.Errorf("provided path is a directory, use 'dir' command instead")
	}

	if filepath.Ext(filearg) != LuauExt {
		// fmt.Println("File is not a .luau file")
		return fmt.Errorf("file is not a .luau file")
	}

	// fmt.Println("Processing file", filearg)

	if err = processFile(filearg, true); err != nil {
		// fmt.Println("Error processing file:", err)
		return fmt.Errorf("process file: %w", err)
	}
	return nil
}

func cmdDir(dirarg string) error {
	info, err := os.Stat(dirarg)
	if err != nil {
		// fmt.Println("Error getting file stat:", err)
		return fmt.Errorf("get file stat: %w", err)
	}

	if !info.IsDir() {
		// fmt.Println("Provided path is not a directory, use 'file' command instead")
		return fmt.Errorf("provided path is not a directory, use 'file' command instead")
	}

	dirEntries, err := os.ReadDir(dirarg)
	if err != nil {
		// fmt.Println("Error reading directory:", err)
		return fmt.Errorf("read directory: %w", err)
	}

	var files []string
	for _, entry := range dirEntries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != LuauExt {
			continue
		}

		files = append(files, dirarg+"/"+entry.Name())
	}

	for _, file := range files {
		fmt.Println("Processing file", file)
		err := processFile(file, false)
		if err != nil {
			fmt.Println("Error processing file:", err)
		}
	}
	return nil
}

func cmdInput(content []byte) error {
	out, err := LuauAstInput(content)
	if err != nil {
		return fmt.Errorf("convert to Luau AST: %w", err)
	}

	// encode as AST
	ast, err := DecodeAST(out)
	if err != nil {
		return fmt.Errorf("decode AST: %w", err)
	}

	newsource, err := ast.Source(string(content))
	if err != nil {
		return fmt.Errorf("encode AST: %w", err)
	}

	fmt.Print(newsource)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: <command>")
		fmt.Println("Available commands: file, dir, input")
		return
	}

	switch command := os.Args[1]; command {
	case "file":
		if len(os.Args) < 3 {
			fmt.Println("Usage: file <path>")
			return
		}
		filearg := os.Args[2]
		if err := cmdFile(filearg); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "dir":
		if len(os.Args) < 3 {
			fmt.Println("Usage: dir <path>")
			return
		}
		dirarg := os.Args[2]
		if err := cmdDir(dirarg); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "input":
		content, err := io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("Error reading from stdin:", err)
			os.Exit(1)
		}
		if err := cmdInput(content); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Unknown command:", command)
		fmt.Println("Available commands: file, dir, input")
		os.Exit(1)
	}
}
