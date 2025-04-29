package net

import (
	"crypto/sha3"
	"net/http"
	"testing"

	"github.com/Heliodex/coputer/bundle"
	"github.com/Heliodex/coputer/litecode/vm"
)

type ProgramTest[A vm.ProgramArgs, R vm.ProgramRets] struct {
	Name string
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

const testProgramPath = "../../test/programs"

var webTests = [...]ProgramTest[vm.WebArgs, vm.WebRets]{
	{
		"web1",
		vm.WebArgs{
			Url:    url("/"),
			Method: "GET",
		},
		vm.WebRets{
			StatusCode:    200,
			StatusMessage: http.StatusText(200),
			Headers: map[string]string{
				"content-type": "text/plain; charset=utf-8",
			},
			Body: []byte("hello GET / world! /"),
		},
	},
	{
		"web1",
		vm.WebArgs{
			Url:    url("/submit?"),
			Method: "POST",
		},
		vm.WebRets{
			StatusCode:    200,
			StatusMessage: http.StatusText(200),
			Headers: map[string]string{
				"content-type": "text/plain; charset=utf-8",
			},
			Body: []byte("hello POST /submit world! /submit?"),
		},
	},
	{
		"web2",
		vm.WebArgs{
			Url:    url("/"),
			Method: "POST",
		},
		vm.WebRets{
			StatusCode:    405,
			StatusMessage: http.StatusText(405),
			Headers: map[string]string{
				"content-type": "text/plain; charset=utf-8",
			},
			Body: []byte(http.StatusText(405)),
		},
	},
	{
		"web2",
		vm.WebArgs{
			Url:    url("/"),
			Method: "GET",
		},
		vm.WebRets{
			StatusCode:    200,
			StatusMessage: http.StatusText(200),
			Headers: map[string]string{
				"content-type": "text/html; charset=utf-8",
			},
			Body: []byte("<h1>WELCOME TO MY WEBSITE</h1>"),
		},
	},
	{
		"web2",
		vm.WebArgs{
			Url:    url("/hello"),
			Method: "GET",
		},
		vm.WebRets{
			StatusCode:    200,
			StatusMessage: http.StatusText(200),
			Headers: map[string]string{
				"content-type": "text/html; charset=utf-8",
			},
			Body: []byte("<p>hello page</p>"),
		},
	},
	{
		"web2",
		vm.WebArgs{
			Url:    url("/error"),
			Method: "GET",
		},
		vm.WebRets{
			StatusCode:    454,
			StatusMessage: "Error 454",
			Headers: map[string]string{
				"content-type": "text/plain; charset=utf-8",
			},
			Body: []byte("Error 454"),
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
		b := getBundled(testProgramPath+"/"+test.Name, t)
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

		n1.Stop()
		n2.Stop()

		t.Log(string(res.Body))
		t.Log("Passed!\n")
	}
}
