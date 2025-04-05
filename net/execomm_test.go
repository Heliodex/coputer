package net

import (
	"fmt"
	"testing"

	"github.com/Heliodex/coputer/bundle"
	"github.com/Heliodex/coputer/litecode/vm"
)

var webArgs = vm.WebArgs{
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

func TestWeb(t *testing.T) {
	const testpath = "../test/web1"

	b, err := bundle.Bundle(testpath)
	if err != nil {
		panic(err)
	}

	hash, err := StoreProgram(b)
	if err != nil {
		panic(err)
	}

	fmt.Println("stored", hash)

	res, err := StartWebProgram(hash, webArgs)
	if err != nil {
		panic(err)
	}
	fmt.Println("ran1")

	res, err = StartWebProgram(hash, webArgs)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(res.Body))
	fmt.Println("ran2")
}
