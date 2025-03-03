package exec

import (
	"fmt"
	"os"
	"testing"

	lc "github.com/Heliodex/litecode"
)

func TestExec(t *testing.T) {
	// clear data directory
	if err := os.RemoveAll("data"); err != nil {
		panic(err)
	}

	path := "./test"
	entrypoint := "main.luau"

	b, err := Bundle(path, entrypoint)
	if err != nil {
		panic(err)
	}

	luau_print := lc.MakeFn("print", func(args lc.Args) (r lc.Rets, err error) {
		for i, arg := range args.List {
			fmt.Print(lc.ToString(arg))

			if i < len(args.List)-1 {
				fmt.Print("\t")
			}
		}

		fmt.Println() // yeah
		return
	})

	var env lc.Env
	env.AddFn(luau_print)

	u := NewUnbundler()

	co, err := u.Execute(b, env)
	if err != nil {
		panic(err)
	}

	res, err := co.Resume()
	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println("Result:", res)
}

func TestBundle(t *testing.T) {
	path := "./test"
	entrypoint := "main.luau"

	b, err := Bundle(path, entrypoint)
	if err != nil {
		panic(err)
	}

	fmt.Println("Bundle:", len(b))
	ub, err := Unbundle(b)
	if err != nil {
		panic(err)
	}

	fmt.Println("Unbundle:")
	for _, f := range ub {
		fmt.Println(f.path, len(f.data))
	}

	// rebundle
	b2, err := Bundle(path, entrypoint)
	if err != nil {
		panic(err)
	}

	if len(b) != len(b2) {
		panic("rebundled bundle is different")
	}
}
