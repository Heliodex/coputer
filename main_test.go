package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime/pprof"
	"strings"
	"testing"
	"time"
)

func litecode(t *testing.T, f string, o uint8) (string, time.Duration) {
	cmd := exec.Command("luau-compile", "--binary", fmt.Sprintf("-O%d", o), f)
	bytecode, err := cmd.Output()
	if err != nil {
		t.Error("error running luau-compile:", err, cmd.Args)
		return "", 0
	}

	deserialised := Deserialise(bytecode)

	b := strings.Builder{}
	luau_print := Function(func(co *Coroutine, args ...any) (ret []any) {
		// b.WriteString(fmt.Sprint(args...))
		for i, arg := range args {
			b.WriteString(tostring(arg))

			if i < len(args)-1 {
				b.WriteString("\t")
			}
		}
		b.WriteString("\r\n") // yeah
		return
	})

	co, _ := Load(deserialised, f, o, map[any]any{
		"print": &luau_print,
	}, map[string]Rets{})

	startTime := time.Now()
	co.Resume()
	endTime := time.Now()

	return b.String(), endTime.Sub(startTime)
}

func luau(f string) (string, error) {
	cmd := exec.Command("luau", f)
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

	// onlyTest := "require.luau"

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		name := f.Name()
		// if name != onlyTest {
		// 	continue
		// }

		fmt.Println(" -- Testing", name, "--")
		filename := fmt.Sprintf("test/%s", name)

		og, err := luau(filename)
		if err != nil {
			t.Error("error running luau:", err)
			return
		}

		o0, _ := litecode(t, filename, 0)
		o1, _ := litecode(t, filename, 1)
		o2, _ := litecode(t, filename, 2)
		fmt.Println()

		for i, o := range []string{o0, o1, o2} {
			if o != og {
				t.Errorf("%d output mismatch:\n-- Expected\n%s\n-- Got\n%s", i, og, o)
				fmt.Println()

				// print mismatch
				oLines := strings.Split(o, "\n")
				ogLines := strings.Split(og, "\n")
				for i, line := range ogLines {
					if line != oLines[i] {
						t.Errorf("mismatched line: \n%s\n%v\n%s\n%v", line, []byte(line), oLines[i], []byte(oLines[i]))
					}
				}

				return
			}
		}

		fmt.Println(og)
	}

	fmt.Println("-- Done! --")
	fmt.Println()
}

// not using benchmark because i can do what i want
func TestBenchmark(t *testing.T) {
	files, err := os.ReadDir("bench")
	if err != nil {
		t.Error("error reading bench directory:", err)
		return
	}

	f, err := os.Create("cpu.prof")
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// onlyBench := "10.luau"

	for _, f := range files {
		name := f.Name()
		// if name != onlyBench {
		// 	continue
		// }

		fmt.Println("\n-- Benchmarking", name, "--")
		filename := fmt.Sprintf("bench/%s", name)

		for o := range uint8(3) {
			output, time := litecode(t, filename, o)

			fmt.Println(" --", o, "Time:", time, "--")
			fmt.Print(output)
		}
	}

	fmt.Println()
}
