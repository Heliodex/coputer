package vm

import (
	"fmt"
	"os"
	"os/exec"
	"runtime/pprof"

	// "runtime/pprof"
	"strings"
	"testing"
	"time"

	"github.com/Heliodex/coputer/litecode/types"
)

const (
	conformanceDir = "../../test/conformance"
	errorsDir      = "../../test/error"
	benchDir       = "../../test/benchmark"
)

func trimext(s string) string {
	return strings.TrimSuffix(s, Ext)
}

func litecode(t *testing.T, f string, c types.Compiler) (string, time.Duration) {
	p, err := Compile(c, f)
	if err != nil {
		t.Fatal(err)
	}

	b := strings.Builder{}
	luau_print := MakeFn("print", func(args Args) (r []types.Val, err error) {
		// b.WriteString(fmt.Sprint(args...))
		for i, arg := range args.List {
			b.WriteString(ToString(arg))

			if i < len(args.List)-1 {
				b.WriteString("\t")
			}
		}
		b.WriteString("\n") // yeah2
		return
	})

	var env types.Env
	env.AddFn(luau_print)

	co, _ := p.Load(env, types.TestArgs{})

	startTime := time.Now()
	_, err = co.Resume()
	if err != nil {
		t.Fatal(err)
	}
	endTime := time.Now()

	return strings.ReplaceAll(b.String(), "\r\n", "\n"), endTime.Sub(startTime)
}

func litecodeE(t *testing.T, f string, c types.Compiler) (string, error) {
	p, err := Compile(c, f)
	if err != nil {
		t.Fatal(err)
	}

	b := strings.Builder{}
	luau_print := MakeFn("print", func(args Args) (r []types.Val, err error) {
		// b.WriteString(fmt.Sprint(args...))
		for i, arg := range args.List {
			b.WriteString(ToString(arg))

			if i < len(args.List)-1 {
				b.WriteString("\t")
			}
		}
		b.WriteString("\n")
		return
	})

	var env types.Env
	env.AddFn(luau_print)

	co, _ := p.Load(env, types.TestArgs{})

	_, err = co.Resume()

	return strings.ReplaceAll(b.String(), "\r\n", "\n"), err
}

func luau(f string) (string, error) {
	cmd := exec.Command("luau", f+Ext)
	o, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(o), nil
}

func TestConformance(t *testing.T) {
	files, err := os.ReadDir(conformanceDir)
	if err != nil {
		t.Fatal("error reading conformance tests directory:", err)
	}

	// const onlyTest = "luauception"

	c0, c1, c2 := NewCompiler(0), NewCompiler(1), NewCompiler(2)

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		name := trimext(f.Name())
		// if name != onlyTest {
		// 	continue
		// }

		t.Log(" -- Testing", name, "--")
		filename := fmt.Sprintf("%s/%s", conformanceDir, name)

		og, err := luau(filename)
		if err != nil {
			fmt.Println("failed on filename", filename)
			t.Fatal("error running luau:", err)
		}

		// fix all newlines to be \n
		og = strings.ReplaceAll(og, "\r\n", "\n")

		o0, _ := litecode(t, filename, c0)
		o1, _ := litecode(t, filename, c1)
		o2, _ := litecode(t, filename, c2)
		fmt.Println()

		for i, o := range []string{o0, o1, o2} {
			if o != og {
				t.Errorf("%d output mismatch:\n-- Expected\n%s\n-- Got\n%s\n", i, og, o)
				fmt.Println()

				// print mismatch
				oLines := strings.Split(o, "\n")
				ogLines := strings.Split(og, "\n")
				for i, line := range ogLines {
					if line != oLines[i] {
						t.Errorf("mismatched line: \n%s\n%v\n%s\n%v\n", line, []byte(line), oLines[i], []byte(oLines[i]))
					}
				}

				os.Exit(1)
			}
		}

		fmt.Println(og)
	}
}

func TestErrors(t *testing.T) {
	files, err := os.ReadDir(errorsDir)
	if err != nil {
		t.Fatal("error reading error tests directory:", err)
	}

	c1 := NewCompiler(1) // just test O1 for the time being

	// const onlyTest = "loc"

	for _, f := range files {
		fn := f.Name()
		if !strings.HasSuffix(fn, Ext) {
			continue
		}
		name := trimext(fn)

		// if name != onlyTest {
		// 	continue
		// }

		t.Log(" -- Testing", name, "--\n")
		filename := fmt.Sprintf("%s/%s", errorsDir, name)

		_, lerr := litecodeE(t, filename, c1)

		if lerr == nil {
			t.Fatal("expected error, got nil")
		}

		errorname := fmt.Sprintf("%s/%s.txt", errorsDir, name)
		og, err := os.ReadFile(errorname)
		if err != nil {
			t.Fatal("error reading error file (meta lol):", err)
		}

		strog := strings.ReplaceAll(string(og), "{PATH}", errorsDir)
		strog = strings.ReplaceAll(strog, "\r\n", "\n")

		fmt.Println(lerr)
		fmt.Println()
		if lerr.Error() != strog {
			t.Fatalf("error mismatch:\n-- Expected\n%s\n-- Got\n%s", strog, lerr)
		}
	}
}

// not using benchmark because i can do what i want
func TestBenchmark(t *testing.T) {
	files, err := os.ReadDir(benchDir)
	if err != nil {
		t.Fatal("error reading benchmark tests directory:", err)
	}

	f, err := os.Create("cpu.prof")
	if err != nil {
		t.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	// const onlyBench = "luauception"

	compilers := []types.Compiler{NewCompiler(0), NewCompiler(1), NewCompiler(2)}

	for _, f := range files {
		name := trimext(f.Name())
		// if name != onlyBench {
		// 	continue
		// }

		fmt.Println()
		t.Log("-- Benchmarking", name, "--")
		filename := fmt.Sprintf("%s/%s", benchDir, name)

		for o, compiler := range compilers {
			output, time := litecode(t, filename, compiler)

			t.Log("  --", o, "Time:", time, "--\n")
			fmt.Println(output)
		}
	}
}
