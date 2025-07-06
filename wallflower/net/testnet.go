package net

import "github.com/Heliodex/coputer/wallflower/keys"

type TestNet struct {
	peers map[keys.Address]*Node // known peers
}

func NewTestNet() Net {
	return &TestNet{
		peers: make(map[keys.Address]*Node),
	}
}

func (n *TestNet) sendToReceiver(addr keys.Address, msg EncryptedMsg) {
	if node, ok := n.peers[addr]; ok {
		node.ReceiveRaw <- msg
	}
}

func (n *TestNet) receiveFromSender(s Sender) {
	for msg := range s {
		addr := msg.Peer.MainAddr // TestNet only uses main addresses
		n.sendToReceiver(addr, msg.EncryptedMsg)
	}
}

func (n *TestNet) AddNode(node *Node) {
	n.peers[node.MainAddr] = node
	for _, addr := range node.AltAddrs {
		n.peers[addr] = node
	}

	go n.receiveFromSender(node.SendRaw)
}
