package net

import (
	"encoding/base64"
	"fmt"
	"iter"
	"strings"

	"github.com/Heliodex/litecode/keys"
)

const FindStart = "cofind:"

func PeerFromFindString(find string) (p keys.Peer, err error) {
	if !strings.HasPrefix(find, FindStart) || find[56] != '.' {
		return keys.Peer{}, fmt.Errorf("not a valid find string")
	}

	pk, err := keys.DecodePK(keys.PubStart + find[7:56]) // up until 1st dot
	if err != nil {
		return
	}

	rest := find[57:] // after the dot

	decodedAddrs, err := base64.RawURLEncoding.DecodeString(rest)
	if err != nil {
		return
	}

	addrs, ok := pk.Verify(decodedAddrs)
	if !ok {
		return keys.Peer{}, fmt.Errorf("invalid addresses signature")
	} else if len(addrs)%keys.AddressLen != 0 {
		return keys.Peer{}, fmt.Errorf("invalid addresses part length")
	}

	addresses := make([]keys.Address, len(addrs)/keys.AddressLen)
	for i := range addresses {
		copy(addresses[i][:], addrs[i*keys.AddressLen:][:keys.AddressLen])
	}

	return keys.Peer{Pk: pk, Addresses: addresses}, nil
}

type EncryptedMessage []byte

func (m EncryptedMessage) Decode(kp keys.Keypair) (msg Message, ok bool) {
	from, body, ok := kp.Decrypt(m)
	if !ok {
		return
	}

	return Message{from, body}, true
}

type Message struct {
	From keys.Peer
	Body []byte
}

type Node struct {
	keys.ThisPeer

	Peers      []keys.Peer // known peers
	SendRaw    func(peer keys.Peer, msg []byte) (err error)
	ReceiveRaw <-chan EncryptedMessage
}

// A find string encodes the pk and addresses
func (n Node) FindString() string {
	pk := n.Kp.Pk.Encode()[6:]

	addrs := make([]byte, len(n.Addresses)*keys.AddressLen)
	for i, addr := range n.Addresses {
		copy(addrs[i*keys.AddressLen:], addr[:])
	}

	signedAddrs := n.Kp.Sk.Sign(addrs)                                // yes, actually works now
	encodedAddrs := base64.RawURLEncoding.EncodeToString(signedAddrs) // we might do ipv6/libp2p/port enocding or smth later
	return fmt.Sprintf("cofind:%s.%s", pk, encodedAddrs)
}

// unoptimised; debug
func (n Node) log(msg ...any) {
	pke := n.Kp.Pk.Encode()
	logId := pke[6:8]

	m := strings.ReplaceAll(fmt.Sprint(msg...), "\n", "\n     ")
	fmt.Printf("[%s] %s\n", logId, m)
}

func (n Node) Send(p keys.Peer, str string) (err error) {
	ct, err := n.ThisPeer.Encrypt([]byte(str), p.Pk)
	if err != nil {
		return
	}

	return n.SendRaw(p, ct)
}

func (n Node) Receive() iter.Seq[Message] {
	return func(y func(Message) bool) {
		ct, ok := <-n.ReceiveRaw
		if !ok {
			return
		}

		msg, ok := ct.Decode(n.Kp)
		if !ok {
			return
		}

		y(msg)
	}
}

func (n Node) Start() {
	pke := n.Kp.Pk.Encode()

	n.log(
		"Starting\n",
		"I'm ", pke, "\n",
		"My primary address is ", n.Addresses[0])

	n.log("I know ", len(n.Peers), " peers")

	// Receiver
	go func() {
		for msg := range n.Receive() {
			n.log(
				"Received ", string(msg.Body), "\n",
				"From ", msg.From.Pk.Encode(), "\n",
				"@ ", msg.From.Addresses[0])
		}
	}()

	// Sender
	for _, peer := range n.Peers {
		n.log(
			"Sending to ", peer.Pk.Encode(), "\n",
			"@ ", peer.Addresses[0])
		n.Send(peer, "Hello, peer!")
	}
}
