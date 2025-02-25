package exec

import (
	"fmt"
	"testing"
)

func TestExec(t *testing.T) {
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
}
