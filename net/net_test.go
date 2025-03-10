package net

import (
	"fmt"
	"testing"
	"time"

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
	
	lnet := LocalNet{}

	n1 := lnet.NewNode()
	fs1 := n1.FindString()

	fmt.Println()
	fmt.Println(fs1)
	fmt.Println()

	time.Sleep(time.Second)

	p1, err := PeerFromFindString(fs1)
	if err != nil {
		panic(err)
	}

	n2 := lnet.NewNode(p1) // tell it about n1

	n2.StoreProgram(b)

	select {}
}
