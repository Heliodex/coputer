package net

import (
	"crypto/rand"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Heliodex/coputer/bundle"
	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
)

type ProgramTest[A ProgramArgs, R ProgramRets] struct {
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

func wurl(s string) (w WebArgsUrl) {
	url, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	return WebArgsUrl{
		Rawpath:  s,
		Path:     url.Path,
		Rawquery: url.RawQuery,
		Query:    queryToMap(url.Query()),
	}
}

const testProgramPath = "../../test/programs"

var webTests = [...]ProgramTest[WebArgs, WebRets]{
	{
		"web1",
		WebArgs{
			Url:    wurl("/"),
			Method: "GET",
		},
		WebRets{
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
		WebArgs{
			Url:    wurl("/submit?"),
			Method: "POST",
		},
		WebRets{
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
		WebArgs{
			Url:    wurl("/"),
			Method: "POST",
		},
		WebRets{
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
		WebArgs{
			Url:    wurl("/"),
			Method: "GET",
		},
		WebRets{
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
		WebArgs{
			Url:    wurl("/hello"),
			Method: "GET",
		},
		WebRets{
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
		WebArgs{
			Url:    wurl("/error"),
			Method: "GET",
		},
		WebRets{
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

// real secret keys for the purposes of testing
var sampleKeys = [...]string{
	"cosec:0aqouiilz3-ynmmxunwx1-7u6e5xppqa-hmz7q8yd3f-5l92e17yos",
	"cosec:0ot4jpb8z4-iq7yu96m3f-9bh2ze9s7w-m7r7vowu2k-tl8pmbetoz",
	"cosec:50u4onk3m0-owyszhfou0-5uvrymlofu-brye4mkomo-3vr2cta2sa",
	"cosec:1omi5wd5ry-acq82a36oo-d73ls1y7h8-tna64ml180-gb4cxjpgk4",
	"cosec:1nikowcxso-yaxz7ewktj-n4cj0bklsd-xbdsl2ipaw-91vww4cex4",
	"cosec:3a1r7x85ki-duan0b0wlk-ate5tun2ag-mdmk5kghrc-3rcpir16w6",
	"cosec:08al1krxnf-u0kmgplotd-yr7fatryv8-9ktqeba3xz-xmzwviykjc",
}
var sampleKeysUsed uint8

func getSampleKeypair() (kp keys.Keypair) {
	if skBytes, err := keys.DecodeSK(sampleKeys[sampleKeysUsed]); err != nil {
		panic("invalid sample key")
	} else if kp, err = keys.KeypairSK(skBytes); err != nil {
		panic("invalid keypair")
	}

	sampleKeysUsed = (sampleKeysUsed + 1) % uint8(len(sampleKeys))
	return
}

func getSampleAddress() (addr keys.Address) {
	rand.Read(addr[:])
	return
}

// signet lel
func TestWeb(t *testing.T) {
	for _, test := range webTests {
		b := getBundled(testProgramPath+"/"+test.Name, t)

		net := NewTestNet()

		n1 := net.NewNode(getSampleKeypair(), getSampleAddress())
		fs1 := n1.FindString()

		p1, err := PeerFromFindString(fs1)
		if err != nil {
			t.Fatal(err)
		} else if err = n1.StoreProgram(n1.Pk, test.Name, b); err != nil {
			t.Fatal(err)
		}

		n2 := net.NewNode(getSampleKeypair(), getSampleAddress())
		n2.AddPeer(p1) // tell it about n1

		resn, err := n2.RunWebProgram(n1.Pk, test.Name, test.Args, false)
		if err != nil {
			t.Fatal(err)
		} else if err := test.Rets.Equal(resn); err != nil {
			t.Fatal("hash return value not equal:", err)
		}

		n1.Stop()
		n2.Stop()

		t.Log(string(resn.Body))
		t.Log("Passed!\n")
	}
}
