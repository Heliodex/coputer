package main

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func filename(f string) string {
	return fmt.Sprintf("test/%s", f)
}

func luauPrint(a any) string {
	switch v := a.(type) {
	case string:
		return v
	case float64:
		if math.Pow10(19) <= v && v < math.Pow10(21) {
			return fmt.Sprintf("%.0f", v)
		} else if float64(int(v)) == v && v < math.Pow10(21) {
			return fmt.Sprintf("%d", int(v))
		} else if v == math.Inf(1) {
			return "inf"
		} else if v == math.Inf(-1) {
			return "-inf"
		} else if v > math.Pow10(50) {
			return fmt.Sprintf("%.0e", v)
		}
		return fmt.Sprint(v)
	}
	return fmt.Sprint(a)
}

func litecode(t *testing.T, f string, o int) string {
	cmd := exec.Command("luau-compile", "--binary", fmt.Sprintf("-O%d", o), filename(f))
	bytecode, err := cmd.Output()
	if err != nil {
		t.Error("error running luau-compile:", err)
		return ""
	}

	deserialised := luau_deserialise(bytecode)

	b := strings.Builder{}
	luau_print := func(args ...any) (ret []any) {
		// b.WriteString(fmt.Sprint(args...))
		for i, arg := range args {
			b.WriteString(luauPrint(arg))
			if i < len(args)-1 {
				b.WriteString("\t")
			}
		}
		b.WriteString("\r\n") // yeah
		return
	}

	fn, _ := luau_load(deserialised, map[any]any{
		"print": &luau_print,
	})
	fn()

	return b.String()
}

func luau(f string) (string, error) {
	cmd := exec.Command("luau", filename(f))
	o, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(o), nil
}

func TestConformance(t *testing.T) {
	files, err := os.ReadDir("test")
	if err != nil {
		t.Error("error reading test directory:", err)
		return
	}

	for _, f := range files {
		name := f.Name()
		fmt.Println(" -- Testing", name, "--")

		og, err := luau(name)
		if err != nil {
			t.Error("error running luau:", err)
			return
		}

		outputs := []string{
			litecode(t, name, 0),
			litecode(t, name, 1),
			litecode(t, name, 2),
		}
		fmt.Println()

		for _, o := range outputs {
			if o != og {
				t.Errorf("output mismatch:\n-- Expected\n%s\n-- Got\n%s", og, o)
				return
			}
		}

		fmt.Println(og)
	}

	fmt.Println("-- Done! --")
	fmt.Println()
}
