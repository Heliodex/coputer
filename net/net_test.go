package net

import (
	"fmt"
	"testing"
)

// signet lel
func TestNet(t *testing.T) {
	lnet := LocalNet{}

	n1 := lnet.NewNode()

	fmt.Println(n1.FindString())

	select {}
}
