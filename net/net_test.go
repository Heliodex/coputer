package net

import (
	"crypto/sha3"
	"testing"

	"github.com/Heliodex/coputer/bundle"
	"github.com/Heliodex/coputer/litecode/vm"
)

type ProgramTest[A vm.ProgramArgs, R vm.ProgramRets] struct {
	Path string
	Args A
	Rets R
}

func url(p string) vm.WebUrl {
	wurl, err := vm.WebUrlFromString(p)
	if err != nil {
		panic(err)
	}
	return wurl
}

var webTests = [...]ProgramTest[vm.WebArgs, vm.WebRets]{
	{
		"../test/web1",
		vm.WebArgs{
			Url:    url("/"),
			Method: "GET",
		},
		vm.WebRets{
			StatusCode:    200,
			StatusMessage: "OK",
			Headers: map[string]string{
				"content-type": "text/html",
			},
			Body: []byte("hello GET / world! /"),
		},
	},
	{
		"../test/web1",
		vm.WebArgs{
			Url:    url("/submit?"),
			Method: "POST",
		},
		vm.WebRets{
			StatusCode:    200,
			StatusMessage: "OK",
			Headers: map[string]string{
				"content-type": "text/html",
			},
			Body: []byte("hello POST /submit world! /submit?"),
		},
	},
}

func getBundled(p string, t *testing.T) (b []byte) {
	b, err := bundle.Bundle(p)
	if err != nil {
		t.Fatal(err)
	}

	return
}

// signet lel
func TestNet(t *testing.T) {
	for _, test := range webTests {
		b := getBundled(test.Path, t)
		hash := sha3.Sum256(b)

		lnet := LocalNet{}

		n1 := lnet.NewNode()
		fs1 := n1.FindString()

		p1, err := PeerFromFindString(fs1)
		if err != nil {
			t.Fatal(err)
		} else if err = n1.StoreProgram(b); err != nil {
			t.Fatal(err)
		}

		n2 := lnet.NewNode()
		n2.AddPeer(p1) // tell it about n1

		res, err := n2.RunWebProgram(hash, test.Args, false)
		if err != nil {
			t.Fatal(err)
		} else if err := test.Rets.Equal(res); err != nil {
			t.Fatal("return value not equal:", err)
		}

		t.Log(string(res.Body))
		t.Log("Passed!\n")
		break
	}
}
