package main

import (
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestHello(t *testing.T) {
	fmt.Println("Compiling")

	// execute luau-compile
	cmd := exec.Command("luau-compile", "--binary", "-O0", "main.luau")
	// get the output
	bytecode, err := cmd.Output()
	if err != nil {
		t.Error("error running luau-compile:", err)
	}

	deserialised := luau_deserialise(bytecode)

	output1 := strings.Builder{}
	luau_print := func(args ...any) (ret []any) {
		output1.WriteString(fmt.Sprint(args...))
		output1.WriteString("\r\n")
		return
	}

	fn, _ := luau_load(deserialised, map[any]any{
		"print": &luau_print,
	})
	fn()

	fmt.Println()

	cmd2 := exec.Command("luau", "main.luau")
	output2, err := cmd2.Output()
	if err != nil {
		t.Error("error running luau:", err)
	}

	result1 :=  output1.String()
	result2 := string(output2)

	if result1 != result2 {
		t.Errorf("output mismatch:\n%s\n%s", result1, result2)
	}
}
