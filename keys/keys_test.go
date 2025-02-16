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

func TestEncode(t *testing.T) {
	// const pkf = "copub:4wjzd3p8o-fso3bbdm9-aecqtu6zz-vesozdnna-zbcdeuo9r"
	const skf = "cosec:4eh0gemdyr-rqvuo9ijxc-ahzegh6881-taulhavmh4-lo1ziwy3v2"

	sk, err := DecodeSK(skf)
	if err != nil {
		panic(err)
	}

	kp, err := keypair(sk)
	if err != nil {
		panic(err)
	}

	pk := kp.Pk
	fmt.Println(pk)

	pkf := pk.Encode()

	fmt.Println(pkf)
	fmt.Println(skf)
	fmt.Println()

	const message = "This is a message!"

	sig, err := kp.Sign([]byte(message))
	if err != nil {
		panic(err)
	}

	fmt.Println(sig)


	ver, err := pk.Verify([]byte(message), sig)
	if err != nil {
		panic(err)
	}

	fmt.Println(ver)

	fmt.Println()
}
