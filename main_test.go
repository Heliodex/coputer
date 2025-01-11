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
	luau_print := Fn(func(co *Coroutine, args ...any) (r Rets, err error) {
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
		"print": luau_print,
	}, map[string]Rets{})

	startTime := time.Now()
	_, err = co.Resume()
	if err != nil {
		panic(err)
	}
	endTime := time.Now()

	return b.String(), endTime.Sub(startTime)
}

func litecodeE(t *testing.T, f string, o uint8) (string, error) {
	cmd := exec.Command("luau-compile", "--binary", fmt.Sprintf("-O%d", o), f)
	bytecode, err := cmd.Output()
	if err != nil {
		t.Error("error running luau-compile:", err, cmd.Args)
		return "", err
	}

	deserialised := Deserialise(bytecode)

	b := strings.Builder{}
	luau_print := Fn(func(co *Coroutine, args ...any) (r Rets, err error) {
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
		"print": luau_print,
	}, map[string]Rets{})

	_, err = co.Resume()

	return b.String(), err
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

	onlyTest := "libtable.luau"

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		name := f.Name()
		if name != onlyTest {
			continue
		}

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
				// oLines := strings.Split(o, "\n")
				// ogLines := strings.Split(og, "\n")
				// for i, line := range ogLines {
				// 	if line != oLines[i] {
				// 		t.Errorf("mismatched line: \n%s\n%v\n%s\n%v", line, []byte(line), oLines[i], []byte(oLines[i]))
				// 	}
				// }

				return
			}
		}

		fmt.Println(og)
	}

	fmt.Println("-- Done! --")
	fmt.Println()
}

func errorsEqual(e0, e1, e2 error) bool {
	return e0.Error() == e1.Error() && e1.Error() == e2.Error()
}

func TestErrors(t *testing.T) {
	files, err := os.ReadDir("error")
	if err != nil {
		t.Error("error reading error directory:", err)
		return
	}

	has := []string{} // actually warranted to use one of these here

	for _, f := range files {
		name := f.Name()

		if strings.HasSuffix(name, ".luau") {
			has = append(has, strings.TrimSuffix(name, ".luau"))
		}
	}

	for _, name := range has {
		fmt.Println(" -- Testing", name, "--")
		filename := fmt.Sprintf("error/%s.luau", name)

		_, err0 := litecodeE(t, filename, 0)
		_, err1 := litecodeE(t, filename, 1)
		_, err2 := litecodeE(t, filename, 2)

		if err0 == nil || err1 == nil || err2 == nil {
			t.Error("expected error, got nil")
			return
		} else if !errorsEqual(err0, err1, err2) {
			t.Error("errors not equal for o1, o2, o3")
			return
		}
		
		errorname := fmt.Sprintf("error/%s.txt", name)
		og, err := os.ReadFile(errorname)
		if err != nil {
			t.Error("error reading error file (meta lol):", err)
			return
		}

		fmt.Println(err0)
		if err0.Error() != string(og) {
			t.Errorf("error mismatch:\n-- Expected\n%s\n-- Got\n%s", og, err0)
			return
		}
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

	// onlyBench := "largealloc.luau"

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

	fmt.Println("-- Done! --")
	fmt.Println()
}
