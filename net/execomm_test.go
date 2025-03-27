package net

import (
	"fmt"
	"testing"

	"github.com/Heliodex/coputer/exec"
	"github.com/Heliodex/coputer/litecode/vm"
)

func TestWeb(t *testing.T) {
	const testpath = "../testweb"

	b, err := exec.Bundle(testpath)
	if err != nil {
		panic(err)
	}

	hash, err := StoreProgram(b)
	if err != nil {
		panic(err)
	}

	fmt.Println("stored", hash)

	args := vm.WebArgs{
		Url: vm.WebUrl{
			Rawpath:  "/?test=true",
			Path:     "/",
			Rawquery: "test=true",
			Query:    map[string]string{"test": "true"},
		},
		Method: "GET",
		Headers: map[string]string{
			"User-Agent": "Roblox/WinInet",
		},
	}

	res, err := RunWebProgram(hash, args)
	if err != nil {
		panic(err)
	}
	fmt.Println("ran1")

	res, err = RunWebProgram(hash, args)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
	fmt.Println("ran2")
}
