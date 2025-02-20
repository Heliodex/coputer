package net

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Heliodex/litecode/keys"
)

const AddressLen = 16

type Address [AddressLen]byte // can be whatever (probably an ipv6 lel)

type Peer struct {
	Pk        keys.PK
	Addresses []Address
}

func (p Peer) Equals(p2 Peer) bool {
	return p.Pk == p2.Pk
}

type Message []byte

type Node struct {
	Peer
	Kp keys.Keypair

	Peers   []Peer // known peers
	Send    func(peer Peer, msg Message)
	Receive <-chan Message
}

// A find string encodes the pk and addresses
func (n Node) FindString() string {
	pk := n.Kp.Pk.Encode()[6:]

	addrs := make([]byte, len(n.Addresses)*AddressLen)
	for i, addr := range n.Addresses {
		copy(addrs[i*AddressLen:], addr[:])
	}

	signedAddrs := n.Kp.Sk.Sign(addrs)
	fmt.Println(string(signedAddrs))

	encodedAddrs := base64.RawURLEncoding.EncodeToString(addrs) // we might do ipv6/libp2p/port enocding or smth later
	return fmt.Sprintf("cofind:%s.%s", pk, encodedAddrs)
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
		"My primary address is ", n.Addresses[0])

	n.log("I know ", len(n.Peers), " peers")

	for msg := range n.Receive {
		n.log("Received ", msg)
	}
}
