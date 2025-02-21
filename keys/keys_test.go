package keys

import (
	"fmt"
	"testing"
)

func TestKeys(t *testing.T) {
	fmt.Println("starting")

	kp := GenerateKeys(1, nil)

	fmt.Println()
	fmt.Println(kp.Pk)
	pkf := kp.Pk.Encode()
	skf := kp.Sk.Encode()
	fmt.Println(pkf)
	fmt.Println(skf)
	fmt.Println()
}

// const pkf1 = "copub:1mdy2o0f9-s1a9rdjkt-vwut3s6fv-gd1nv0ezr-it04zc2le"
const skf1 = "cosec:1omi5wd5ry-acq82a36oo-d73ls1y7h8-tna64ml180-gb4cxjpgk4"

// const pkf2 = "copub:0jai56z6p-lzkysnq8n-930230ws9-gbm9d55jy-sqhjy8w20"
const skf2 = "cosec:1nikowcxso-yaxz7ewktj-n4cj0bklsd-xbdsl2ipaw-91vww4cex4"

func Assert(err error) {
	if err != nil {
		panic(err)
	}
}

func TestEncrypt(t *testing.T) {
	sk1, err := DecodeSK(skf1)
	Assert(err)
	sk2, err := DecodeSK(skf2)
	Assert(err)

	kp1, err := KeypairSK(sk1)
	Assert(err)
	kp2, err := KeypairSK(sk2)
	Assert(err)

	//

	message := []byte("what's up world!")

	enc, err := kp1.Encrypt(message, kp2.Pk)
	Assert(err)

	fmt.Println(len(enc), len(message))
	fmt.Println(len(enc) == len(message)+93)
	fmt.Println()

	dec, rpk, ok := kp2.Decrypt(enc)
	fmt.Println(string(dec))
	fmt.Println(ok)
	fmt.Println("from", rpk.Encode())
	fmt.Println()
}

func TestSign(t *testing.T) {
	sk1, err := DecodeSK(skf1)
	Assert(err)

	kp1, err := KeypairSK(sk1)
	Assert(err)

	//

	message := []byte("what's up world!")

	sig := kp1.Sk.Sign(message)

	fmt.Println(len(sig), len(message))
	fmt.Println(len(sig) == len(message)+16) // hey wouldya look at that, 16 bytes of overhead, i'm a genuis

	ver, ok := kp1.Pk.Verify(sig)
	fmt.Println(string(ver))
	fmt.Println(ok)
}
