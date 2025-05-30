package net

import (
	"github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
)

// real secret keys for the purposes of testing
var sampleKeys = [...]string{
	"cosec:0aqouiilz3-ynmmxunwx1-7u6e5xppqa-hmz7q8yd3f-5l92e17yos",
	"cosec:0ot4jpb8z4-iq7yu96m3f-9bh2ze9s7w-m7r7vowu2k-tl8pmbetoz",
	"cosec:50u4onk3m0-owyszhfou0-5uvrymlofu-brye4mkomo-3vr2cta2sa",
	"cosec:1omi5wd5ry-acq82a36oo-d73ls1y7h8-tna64ml180-gb4cxjpgk4",
	"cosec:1nikowcxso-yaxz7ewktj-n4cj0bklsd-xbdsl2ipaw-91vww4cex4",
	"cosec:3a1r7x85ki-duan0b0wlk-ate5tun2ag-mdmk5kghrc-3rcpir16w6",
	"cosec:08al1krxnf-u0kmgplotd-yr7fatryv8-9ktqeba3xz-xmzwviykjc",
}
var sampleKeysUsed uint8

func getSampleKeypair() (kp keys.Keypair) {
	if skBytes, err := keys.DecodeSK(sampleKeys[sampleKeysUsed]); err != nil {
		panic("invalid sample key")
	} else if kp, err = keys.KeypairSK(skBytes); err != nil {
		panic("invalid keypair")
	}

	sampleKeysUsed = (sampleKeysUsed + 1) % uint8(len(sampleKeys))
	return
}

type LocalPeer struct {
	keys.Peer
	Receive chan<- EncryptedMsg
}

type LocalNet struct {
	ExistingPeers []LocalPeer
}

func (n *LocalNet) AddPeer(p keys.Peer, recv chan<- EncryptedMsg) {
	n.ExistingPeers = append(n.ExistingPeers, LocalPeer{p, recv})
}

func (n *LocalNet) SendRaw(p *keys.Peer, m []byte) (err error) {
	for _, ep := range n.ExistingPeers {
		if p.Equals(ep.Peer) {
			ep.Receive <- m
			// if we know we can reach the peer some other way then we should do that
			return
		}
	}

	return
}

func (n *LocalNet) NewNode() (node *Node) {
	kp := getSampleKeypair()
	peer := keys.Peer{
		Pk:        kp.Pk,
		Addresses: []keys.Address{{sampleKeysUsed}}, // sequential placeholder
	}

	recv := make(chan EncryptedMsg)
	n.AddPeer(peer, recv)
	node = &Node{
		ThisPeer: keys.ThisPeer{
			Peer: peer,
			Kp:   kp,
		},
		Peers:              make(map[keys.PK]*keys.Peer),
		SendRaw:            n.SendRaw,
		ReceiveRaw:         recv,
		resultsWaitingHash: make(map[InputHash]chan types.ProgramRets),
		resultsWaitingName: make(map[InputName]chan types.ProgramRets),
	}

	go node.Start()
	return
}
