package net

import (
	"fmt"
	"strings"

	"github.com/Heliodex/litecode/keys"
)

type Address [16]byte // can be whatever

type Peer struct {
	Pk      keys.PK
	Address Address
}

type Message []byte

type Node struct {
	Kp      keys.Keypair
	Address Address

	Peers   []Peer // known peers
	Send    func(peer Peer, msg Message)
	Receive <-chan Message
}

// unoptimised; debug
func (n Node) log(msg ...any) {
	pke := n.Kp.Pk.Encode()
	logId := pke[6:8]

	m := strings.ReplaceAll(fmt.Sprint(msg...), "\n", "\n     ")
	fmt.Printf("[%s] %s\n", logId, m)
}

func (n Node) Start() {
	pke := n.Kp.Pk.Encode()

	n.log(
		"Starting\n",
		"I'm ", pke, "\n",
		"My address is ", n.Address)

	for msg := range n.Receive {
		n.log("Received ", msg)
	}
}
