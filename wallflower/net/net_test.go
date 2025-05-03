package net

import (
	"crypto/sha3"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Heliodex/coputer/bundle"
	"github.com/Heliodex/coputer/litecode/types"
)

type ProgramTest[A types.ProgramArgs, R types.ProgramRets] struct {
	Name string
	Args A
	Rets R
}

func queryToMap(q url.Values) (m map[string]string) {
	m = make(map[string]string, len(q))
	for k, v := range q {
		m[k] = strings.Join(v, "")
	}

	return
}

func wurl(s string) (w types.WebArgsUrl) {
	url, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	return types.WebArgsUrl{
		Rawpath:  s,
		Path:     url.Path,
		Rawquery: url.RawQuery,
		Query:    queryToMap(url.Query()),
	}
}

const testProgramPath = "../../test/programs"

var webTests = [...]ProgramTest[types.WebArgs, types.WebRets]{
	{
		"web1",
		types.WebArgs{
			Url:    wurl("/"),
			Method: "GET",
		},
		types.WebRets{
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
		types.WebArgs{
			Url:    wurl("/submit?"),
			Method: "POST",
		},
		types.WebRets{
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
		types.WebArgs{
			Url:    wurl("/"),
			Method: "POST",
		},
		types.WebRets{
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
		types.WebArgs{
			Url:    wurl("/"),
			Method: "GET",
		},
		types.WebRets{
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
		types.WebArgs{
			Url:    wurl("/hello"),
			Method: "GET",
		},
		types.WebRets{
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
		types.WebArgs{
			Url:    wurl("/error"),
			Method: "GET",
		},
		types.WebRets{
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

		var lnet LocalNet

		n1 := lnet.NewNode()
		fs1 := n1.FindString()

		p1, err := PeerFromFindString(fs1)
		if err != nil {
			t.Fatal(err)
		} else if err = n1.StoreProgram(test.Name, b); err != nil {
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
