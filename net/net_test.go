package net

import (
	"fmt"
	"testing"
	"time"
)

// signet lel
func TestNet(t *testing.T) {
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

	lnet.NewNode(p1) // tell it about n1

	select {}
}
