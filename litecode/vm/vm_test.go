package vm

import (
	"fmt"
	"os"
	"os/exec"

	// "runtime/pprof"
	"strings"
	"sync"
	"testing"
	"time"
)

func trimext(s string) string {
	return strings.TrimSuffix(s, Ext)
}

func litecode(t *testing.T, f string, c Compiler) (string, time.Duration) {
	p, err := c.Compile(f)
	if err != nil {
		t.Error(err)
		return "", 0
	}

	b := strings.Builder{}
	luau_print := MakeFn("print", func(args Args) (r Rets, err error) {
		// b.WriteString(fmt.Sprint(args...))
		for i, arg := range args.List {
			b.WriteString(ToString(arg))

			if i < len(args.List)-1 {
				b.WriteString("\t")
			}
		}
		b.WriteString("\r\n") // yeah
		return
	})

	var env Env
	env.AddFn(luau_print)

	co, _ := p.Load(env)

	startTime := time.Now()
	_, err = co.Resume()
	if err != nil {
		panic(err)
	}
	endTime := time.Now()

	return b.String(), endTime.Sub(startTime)
}

func litecodeE(t *testing.T, f string, c Compiler) (string, error) {
	p, err := c.Compile(f)
	if err != nil {
		t.Error(err)
		return "", err
	}

	b := strings.Builder{}
	luau_print := MakeFn("print", func(args Args) (r Rets, err error) {
		// b.WriteString(fmt.Sprint(args...))
		for i, arg := range args.List {
			b.WriteString(ToString(arg))

			if i < len(args.List)-1 {
				b.WriteString("\t")
			}
		}
		b.WriteString("\r\n") // yeah
		return
	})

	var env Env
	env.AddFn(luau_print)

	co, _ := p.Load(env)

	_, err = co.Resume()

	return b.String(), err
}

func luau(f string) (string, error) {
	cmd := exec.Command("luau", f+Ext)
	o, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(o), nil
}

const parallel = false

func TestConformance(t *testing.T) {
	files, err := os.ReadDir("test")
	if err != nil {
		t.Error("error reading test directory:", err)
		return
	}

	// onlyTest := "calls"
	var wg sync.WaitGroup

	c0, c1, c2 := NewCompiler(0), NewCompiler(1), NewCompiler(2)

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		name := trimext(f.Name())
		// if name != onlyTest {
		// 	continue
		// }

		run := func() {
			if parallel {
				defer wg.Done()
			}

			fmt.Println(" -- Testing", name, "--")
			filename := fmt.Sprintf("./test/%s", name) // ./ required for requires

			og, err := luau(filename)
			if err != nil {
				fmt.Println("error running luau:", err)
				os.Exit(1)
			}

			o0, _ := litecode(t, filename, c0)
			o1, _ := litecode(t, filename, c1)
			o2, _ := litecode(t, filename, c2)
			fmt.Println()

			for i, o := range []string{o0, o1, o2} {
				if o != og {
					fmt.Printf("%d output mismatch:\n-- Expected\n%s\n-- Got\n%s\n", i, og, o)
					fmt.Println()

					// print mismatch
					oLines := strings.Split(o, "\n")
					ogLines := strings.Split(og, "\n")
					for i, line := range ogLines {
						if line != oLines[i] {
							fmt.Printf("mismatched line: \n%s\n%s\n", line, oLines[i])
						}
					}

					os.Exit(1)
				}
			}

			fmt.Println(og)
		}

		if parallel {
			wg.Add(1)

			go run()
		} else {
			run()
		}
	}

	wg.Wait()

	fmt.Println("-- Done! --")
	fmt.Println()
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

		if strings.HasSuffix(name, Ext) {
			has = append(has, trimext(name))
		}
	}

	c1 := NewCompiler(1) // just test O1 for the time being

	// onlyTest := "requireinit"

	for _, name := range has {
		// if name != onlyTest {
		// 	continue
		// }

		fmt.Println(" -- Testing", name, "--")
		filename := fmt.Sprintf("error/%s", name)

		_, lerr := litecodeE(t, filename, c1)

		if lerr == nil {
			t.Error("expected error, got nil")
			return
		}

		errorname := fmt.Sprintf("error/%s.txt", name)
		og, err := os.ReadFile(errorname)
		if err != nil {
			t.Error("error reading error file (meta lol):", err)
			return
		}

		fmt.Println(lerr)
		if lerr.Error() != strings.ReplaceAll(string(og), "\r\n", "\n") {
			t.Errorf("error mismatch:\n-- Expected\n%s\n-- Got\n%s", og, lerr)
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

	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	panic(err)
	// }
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()

	// onlyBench := "largealloc"

	compilers := []Compiler{NewCompiler(0), NewCompiler(1), NewCompiler(2)}

	for _, f := range files {
		name := trimext(f.Name())
		// if name != onlyBench {
		// 	continue
		// }

		fmt.Println("\n-- Benchmarking", name, "--")
		filename := fmt.Sprintf("bench/%s", name)

		for o := range uint8(3) {
			output, time := litecode(t, filename, compilers[o])

			fmt.Println(" --", o, "Time:", time, "--")
			fmt.Print(output)
		}
	}

	fmt.Println("-- Done! --")
	fmt.Println()
}
