package net

import (
	"crypto/sha3"
	"fmt"
	"testing"

	"github.com/Heliodex/coputer/bundle"
)

const path = "../test/web1"

func getBundled() (b []byte) {
	b, err := bundle.Bundle(path)
	if err != nil {
		panic(err)
	}

	return
}

// signet lel
func TestNet(t *testing.T) {
	b := getBundled()
	hash := sha3.Sum256(b)

	lnet := LocalNet{}

	n1 := lnet.NewNode()
	fs1 := n1.FindString()

	fmt.Println(fs1)
	fmt.Println()

	p1, err := PeerFromFindString(fs1)
	if err != nil {
		panic(err)
	}

	err = n1.StoreProgram(b)
	if err != nil {
		panic(err)
	}
	
	n2 := lnet.NewNode()
	n2.AddPeer(p1) // tell it about n1

	res, err := n2.RunWebProgram(hash, webArgs, false)
	if err != nil {
		panic(err)
	}

	fmt.Println("ran")
	fmt.Println(res)

	select {}
}
