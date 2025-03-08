package net

import (
	"testing"

	"github.com/Heliodex/coputer/exec"
)

const path = "../testb"

func TestExecComm(t *testing.T) {
	b, err := exec.Bundle(path)
	if err != nil {
		panic(err)
	}

	err = StoreProgram(b)
	if err != nil {
		panic(err)
	}
}
