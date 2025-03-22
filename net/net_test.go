package net

import (
	"crypto/sha3"
	"fmt"
	"testing"

	"github.com/Heliodex/coputer/exec"
)

const path = "../testb"

func getBundled() (b []byte) {
	b, err := exec.Bundle(path)
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

	fmt.Println()
	fmt.Println(fs1)
	fmt.Println()

	p1, err := PeerFromFindString(fs1)
	if err != nil {
		panic(err)
	}

	n2 := lnet.NewNode(p1) // tell it about n1

	err = n2.StoreProgram(b)
	if err != nil {
		panic(err)
	}

	fmt.Println("stored")

	res, err := n1.RunProgram(hash, "cruel")
	if err != nil {
		panic(err)
	}

	fmt.Println("ran")
	fmt.Println(res)

	select {}
}
