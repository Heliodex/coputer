package main

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/Heliodex/coputer/exec"
	"github.com/Heliodex/coputer/litecode/vm"
)

func Start(c vm.Compiler, hash string) (runfn func(input string) (output string, err error), err error) {
	p, err := c.Compile(filepath.Join(exec.ProgramsDir, hash, exec.Entrypoint))
	if err != nil {
		return
	}

	luau_print := vm.MakeFn("print", func(args vm.Args) (r vm.Rets, err error) {
		for _, arg := range args.List {
			fmt.Print("\t")
			fmt.Print(vm.ToString(arg))
		}
		fmt.Println() // yeah
		return
	})

	var env vm.Env
	env.AddFn(luau_print)

	co, cancel := p.Load(env)

	go func() {
		time.Sleep(5 * time.Second)
		cancel()
	}()

	r, err := co.Resume()
	if err != nil {
		return
	} else if len(r) != 1 {
		return nil, errors.New("program did not return a single value")
	}

	fn, ok := r[0].(vm.Function)
	if !ok {
		return nil, errors.New("program did not return a function")
	}

	return func(input string) (output string, err error) {
		rets, err := (*fn.Run)(&co, input)
		if err != nil {
			return
		} else if len(rets) != 1 {
			return "", errors.New("program did not return a single value")
		} else if output, ok = rets[0].(string); !ok {
			return "", errors.New("program did not return a string")
		}

		return
	}, nil
}
