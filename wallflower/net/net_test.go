package net

import (
	"crypto/rand"
	"net/http"
	"net/url"
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

func wurl(s string) (w WebArgsUrl) {
	url, err := url.Parse(s)
	if err != nil {
		panic(err)
	}

	return WebArgsUrl{
		Rawpath:  s,
		Path:     url.Path,
		Rawquery: url.RawQuery,
		Query:    url.Query(),
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
			Headers: map[string]string{
				"content-type": "text/plain; charset=utf-8",
			},
			Body: []byte("Error 454"),
		},
	},
	{
		"web3",
		WebArgs{
			Url:    wurl("/error"),
			Method: "GET",
		},
		WebRets{
			StatusCode:    404,
			Headers: map[string]string{
				"content-type": "text/plain; charset=utf-8",
			},
			Body: []byte(http.StatusText(404)),
		},
	},
	{
		"web3",
		WebArgs{
			Url:    wurl("/"),
			Method: "GET",
		},
		WebRets{
			StatusCode:    200,
			Body: []byte(`Raw query: `),
		},
	},
	{
		"web3",
		WebArgs{
			Url:    wurl("/?a="),
			Method: "GET",
		},
		WebRets{
			StatusCode:    200,
			Body: []byte(`a =
- 
Raw query: a=`),
		},
	},
	{
		"web3",
		WebArgs{
			Url:    wurl("/?a=b&c=b"),
			Method: "GET",
		},
		WebRets{
			StatusCode:    200,
			Body: []byte(`a =
- b
c =
- b
Raw query: a=b&c=b`),
		},
	},
	{
		"web3",
		WebArgs{
			Url:    wurl("/?c=b&a=b"),
			Method: "GET",
		},
		WebRets{
			StatusCode:    200,
			Body: []byte(`a =
- b
c =
- b
Raw query: c=b&a=b`),
		},
	},
	{
		"web3",
		WebArgs{
			Url:    wurl("/?a=b&a=b"),
			Method: "GET",
		},
		WebRets{
			StatusCode:    200,
			Body: []byte(`a =
- b
- b
Raw query: a=b&a=b`),
		},
	},
	{
		"web3",
		WebArgs{
			Url:    wurl("/?a=c&a=b"),
			Method: "GET",
		},
		WebRets{
			StatusCode:    200,
			Body: []byte(`a =
- c
- b
Raw query: a=c&a=b`),
		},
	},
	{
		"web3",
		WebArgs{
			Url:    wurl("/?a=b&a=c"),
			Method: "GET",
		},
		WebRets{
			StatusCode:    200,
			Body: []byte(`a =
- b
- c
Raw query: a=b&a=c`),
		},
	},
	{
		"web3",
		WebArgs{
			Url:    wurl("/?=b&a=c"),
			Method: "GET",
		},
		WebRets{
			StatusCode:    200,
			Body: []byte(` =
- b
a =
- c
Raw query: =b&a=c`),
		},
	},
	{
		"web3",
		WebArgs{
			Url:    wurl("/?=b&=c"),
			Method: "GET",
		},
		WebRets{
			StatusCode:    200,
			Body: []byte(` =
- b
- c
Raw query: =b&=c`),
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
	"cosec:6c6agugf3g-vs0sq9g3cz-8pgwka0ukw-la26ik86q4-meib6pl634",
	"cosec:0347fqcnc7-zw2o8zslpk-tgddlt5f0z-8yq2lm9rcc-vjczoocdwu",
	"cosec:5v1uy4hyyr-hu0p1lm64j-3hxcm54xtx-amlipibg5i-5ygtw9suve",
	"cosec:0109c65giz-5zj6im75aj-qp2kvcy6f6-dzg3akk10r-y4oruplziw",
	"cosec:10f83ngd0p-axezx8t56i-y8h0klu6ed-85qkof6z7a-66mnch39r7",
	"cosec:0dlwd7uctr-qli9cgp2wh-z4zdukjcso-zwtnieatlq-xkn5fdvxbp",
	"cosec:0bsep2586t-y4wsd552kz-svsgs5cvm0-qny598wmi0-hp9r1i658y",
	"cosec:6cg6g52fgh-qk14xzpzbd-6giuh9wlu3-cy3yeulrpv-dukyvye0y3",
	"cosec:5ki3i1p8ey-v3atgco6qj-eqttki3ad8-blj4arxb5b-wuaakmn0ib",
	"cosec:5vb3gg5slk-jklp9qufn7-gviwuysl26-ht0e8ik23g-xyno6ki2xj",
	"cosec:00bckkuo81-xu6fnu7nnp-sg2dyhnpel-tduilo1r46-ssoxisbyq4",
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

		n1 := NewNode(getSampleKeypair(), getSampleAddress())
		net.AddNode(n1)
		n1.Start()

		fs1 := n1.FindString()

		p1, err := PeerFromFindString(fs1)
		if err != nil {
			t.Fatal(err)
		}

		if err = n1.StoreProgram(n1.Pk, test.Name, b); err != nil {
			t.Fatal(err)
		}

		n2 := NewNode(getSampleKeypair(), getSampleAddress())
		n2.AddPeer(p1) // tell it about n1
		net.AddNode(n2)
		n2.Start()

		resn, err := n2.RunWebProgram(n1.Pk, test.Name, test.Args, false)
		if err != nil {
			t.Fatal(err)
		}

		if err := test.Rets.Equal(resn); err != nil {
			t.Fatal("hash return value not equal:", err)
		}

		n1.Stop()
		n2.Stop()

		t.Log(string(resn.Body))
		t.Log("Passed!\n")
	}
}
