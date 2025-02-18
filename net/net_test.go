package net

import (
	"testing"
	"time"
)

// signet lel
func TestNet(t *testing.T) {
	lnet := LocalNet{}

	lnet.NewNode()

	time.Sleep(time.Second)

	lnet.NewNode()

	select {}
}
