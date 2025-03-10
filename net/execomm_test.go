package net

import (
	"fmt"
	"testing"

	"github.com/Heliodex/coputer/exec"
)

const testpath = "../testb"

func TestExecComm(t *testing.T) {
	b, err := exec.Bundle(testpath)
	if err != nil {
		panic(err)
	}

	hash, err := StoreProgram(b)
	if err != nil {
		panic(err)
	}

	fmt.Println("stored", hash)

	res, err := RunProgram(hash)
	if err != nil {
		panic(err)
	}

	fmt.Println(res)
}
