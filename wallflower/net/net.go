package net

import "github.com/Heliodex/coputer/wallflower/keys"

type (
	AddressedMsg struct {
		EncryptedMsg
		*keys.Peer
	}

	Net interface {
		AddNode(node *Node)
	}

	Receiver chan EncryptedMsg
	Sender chan AddressedMsg
)
