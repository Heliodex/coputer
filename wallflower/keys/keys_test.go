package keys

import (
	"fmt"
	"testing"
)

// not actually a good test
// func TestKeys(t *testing.T) {
// 	fmt.Println("starting")

// 	kp := GenerateKeys(1, nil)

// 	fmt.Println()
// 	fmt.Println(kp.Pk)
// 	pkf := kp.Pk.Encode()
// 	skf := kp.Sk.Encode()
// 	fmt.Println(pkf)
// 	fmt.Println(skf)
// 	fmt.Println()
// }

const (
	pkf1 = "copub:1mdy2o0f9-s1a9rdjkt-vwut3s6fv-gd1nv0ezr-it04zc2le"
	skf1 = "cosec:1omi5wd5ry-acq82a36oo-d73ls1y7h8-tna64ml180-gb4cxjpgk4"
)

func Assert(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestEncode(t *testing.T) {
	pk, err := DecodePK(pkf1)
	Assert(t, err)

	if pk.Encode() != pkf1 {
		t.Fatal("public key encoding mismatch")
	}

	sk, err := DecodeSK(skf1)
	Assert(t, err)

	if sk.Encode() != skf1 {
		t.Fatal("secret key encoding mismatch")
	}

	pkn, err := DecodePKNoPrefix(pkf1[6:])
	Assert(t, err)

	if pkn.Encode() != pkf1 {
		t.Fatal("public key no prefix encoding mismatch")
	}

	skn, err := DecodeSKNoPrefix(skf1[6:])
	Assert(t, err)

	if skn.Encode() != skf1 {
		t.Fatal("secret key no prefix encoding mismatch")
	}

	kp, err := KeypairSK(sk)
	Assert(t, err)

	if kp.Pk.Encode() != pkf1 {
		t.Fatal("public key generated from secret key encoding mismatch")
	}
}

func TestSign(t *testing.T) {
	sk1, err := DecodeSK(skf1)
	Assert(t, err)

	kp1, err := KeypairSK(sk1)
	Assert(t, err)

	message := []byte("what's up world!")

	sig := kp1.Sk.Sign(message)

	fmt.Println(len(sig), len(message))
	if len(sig) != len(message)+16 { // hey wouldya look at that, 16 bytes of overhead, i'm a genuis
		t.Fatal("signature length mismatch")
	}

	ver, ok := kp1.Pk.Verify(sig)
	fmt.Println(string(ver))

	if !ok {
		t.Fatal("signature verification failed")
	}
}
