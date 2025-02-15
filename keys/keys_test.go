package keys

import (
	"fmt"
	"testing"
)

func TestKeys(t *testing.T) {
	fmt.Println("starting")

	kp := GenerateKeys(6, nil)

	fmt.Println()
	fmt.Println(kp.pk)
	pkf := kp.pk.Encode()
	skf := kp.sk.Encode()
	fmt.Println(pkf)
	fmt.Println(skf)
	fmt.Println()
}

func TestEncode(t *testing.T) {
	const pkf = "copub:4wjzd3p8o-fso3bbdm9-aecqtu6zz-vesozdnna-zbcdeuo9r"
	const skf = "cosec:4eh0gemdyr-rqvuo9ijxc-ahzegh6881-taulhavmh4-lo1ziwy3v2"

	pk, err := DecodePK(pkf)
	if err != nil {
		panic(err)
	}

	sk, err := DecodeSK(skf)
	if err != nil {
		panic(err)
	}

	pk2, err := SKtoPK(sk)
	if err != nil {
		panic(err)
	}

	fmt.Println(pk)
	fmt.Println(pk2)

	fmt.Println(pk.Encode())
	fmt.Println(pkf)
	fmt.Println(sk.Encode())
	fmt.Println(skf)
}
