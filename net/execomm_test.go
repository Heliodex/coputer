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

	res, err := RunProgram(hash, "cruel")
	if err != nil {
		panic(err)
	}
	fmt.Println(res, res[len(res)-1] == '\n')
	fmt.Println("ran1")
	
	res, err = RunProgram(hash, "cool")
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
	fmt.Println("ran2")
}
