package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Heliodex/coputer/exec"
	"github.com/Heliodex/coputer/litecode/vm"
)

func Run(c vm.Compiler, hash string) (res vm.Rets, err error) {
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

	return co.Resume()
}
