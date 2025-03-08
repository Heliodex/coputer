package exec

import (
	"fmt"
	"testing"
)

const path = "../testb"

func TestBundle(t *testing.T) {
	b, err := Bundle(path)
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
	b2, err := Bundle(path)
	if err != nil {
		panic(err)
	}

	if len(b) != len(b2) {
		panic("rebundled bundle is different")
	}
}
