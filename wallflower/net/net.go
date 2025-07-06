package net

import "github.com/Heliodex/coputer/wallflower/keys"

type (
	AddressedMsg struct {
		EncryptedMsg
		*keys.Peer
	}
	Sender   chan AddressedMsg
	Receiver chan EncryptedMsg
)

type Net interface {
	AddNode(node *Node)
}
