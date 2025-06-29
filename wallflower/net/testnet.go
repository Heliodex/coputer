package net

import (
	"errors"

	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
)

type TestNet struct {
	ExistingPeers []Peer
}

func NewTestNet() Net {
	return &TestNet{}
}

func (n *TestNet) SendRaw(p *Peer, m []byte) (err error) {
	for _, ep := range n.ExistingPeers {
		if p.Equals(ep) {
			ep.Transfer <- m
			// if we know we can reach the peer some other way then we should do that
			return
		}
	}

	return errors.New("sendraw: unknown peer")
}

func (n *TestNet) NewNode(kp keys.Keypair, mainAddr keys.Address, altAddrs ...keys.Address) (node *Node) {
	peer := Peer{
		Pk:       kp.Pk,
		MainAddr: mainAddr,
		AltAddrs: altAddrs,
		Transfer: make(chan EncryptedMsg),
	}

	n.ExistingPeers = append(n.ExistingPeers, peer)

	node = &Node{
		ThisPeer: ThisPeer{
			Peer: peer,
			Kp:   kp,
		},
		Peers:              make(map[keys.PK]*Peer),
		SendRaw:            n.SendRaw,
		resultsWaitingHash: make(map[InputHash]chan ProgramRets),
		resultsWaitingName: make(map[InputName]chan ProgramRets),
	}

	go node.Start()
	return
}
