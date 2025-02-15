package keys

import (
	"fmt"
	"testing"
)

func TestKeys(t *testing.T) {
	fmt.Println("starting")

	fmt.Println()

	found := make(chan Keypair)
	done := make(chan struct{})

	GenerateKeys(6, found, done)

	kp := <-found

	fmt.Println()
	fmt.Println(kp.pk)
	pkf := kp.pk.Encode()
	skf := kp.sk.Encode()
	fmt.Println(pkf)
	fmt.Println(skf)
	fmt.Println()

	close(done)
}
