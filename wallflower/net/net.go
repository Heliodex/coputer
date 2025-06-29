package net

import "github.com/Heliodex/coputer/wallflower/keys"

type Transfer chan EncryptedMsg

type (
	Peer     = keys.Peer[Transfer]
	ThisPeer = keys.ThisPeer[Transfer]
)

type Net interface {
	SendRaw(p *keys.Peer[Transfer], m []byte) (err error)
	NewNode(kp keys.Keypair, mainAddr keys.Address, altAddrs ...keys.Address) (node *Node)
}
