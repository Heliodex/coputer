package net

import (
	"crypto/sha3"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
)

const FindStart = "cofind:"

func PeerFromFindString(find string) (p *keys.Peer, err error) {
	if !strings.HasPrefix(find, FindStart) || find[56] != '.' {
		return nil, errors.New("not a valid find string")
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
		return nil, errors.New("invalid addresses signature")
	}

	addrsSize := len(addrs)
	if addrsSize%keys.AddressLen != 0 {
		return nil, errors.New("invalid addresses part length")
	}
	addrsCount := addrsSize / keys.AddressLen

	mainAddr := keys.Address{}
	copy(mainAddr[:], addrs[:keys.AddressLen]) // first address is always the main

	altAddrs := make([]keys.Address, addrsCount-1)
	for i := range altAddrs {
		copy(altAddrs[i][:], addrs[(i+1)*keys.AddressLen:][:keys.AddressLen])
	}

	return &keys.Peer{Pk: pk, MainAddr: mainAddr, AltAddrs: altAddrs}, nil
}

func (e EncryptedMsg) Decode(kp keys.Keypair) (am AnyMsg, err error) {
	from, body, err := keys.Decrypt(kp, e)
	if err != nil {
		return
	}

	return AnyMsg{
		From: &from,
		Type: body[0],
		Body: body[1:],
	}, nil
}

// struct keys > nested maps
type InputHash struct {
	Hash, InputHash [32]byte
}

type InputName struct {
	Pk        keys.PK
	Name      string
	InputHash [32]byte
}

type Node struct {
	keys.ThisPeer

	Peers              map[keys.PK]*keys.Peer // known peers
	SendRaw            Sender
	ReceiveRaw         Receiver
	resultsWaitingHash map[InputHash]chan ProgramRets
	resultsWaitingName map[InputName]chan ProgramRets
	running            bool
}

func NewNode(kp keys.Keypair, mainAddr keys.Address, altAddrs ...keys.Address) (node *Node) {
	peer := keys.Peer{
		Pk:       kp.Pk,
		MainAddr: mainAddr,
		AltAddrs: altAddrs,
	}

	return &Node{
		ThisPeer: keys.ThisPeer{
			Peer: peer,
			Kp:   kp,
		},
		Peers:              make(map[keys.PK]*keys.Peer),
		SendRaw:            make(Sender),
		ReceiveRaw:         make(Receiver),
		resultsWaitingHash: make(map[InputHash]chan ProgramRets),
		resultsWaitingName: make(map[InputName]chan ProgramRets),
	}
}

func (n *Node) AddPeer(p *keys.Peer) {
	if p == nil {
		return
	}

	if _, ok := n.Peers[p.Pk]; !ok {
		n.Peers[p.Pk] = p
	}
}

func (n *Node) send(p *keys.Peer, sm SentMsg) (err error) {
	ct, err := n.Encrypt(sm.Serialise(), p.Pk)
	if err != nil {
		return
	}

	n.SendRaw <- AddressedMsg{
		EncryptedMsg: ct,
		Peer:         p,
	}
	return
}

// A find string encodes the pk and addresses
func (n Node) FindString() string {
	pk := n.Kp.Pk.Encode()[6:]

	addrs := make([]byte, (len(n.AltAddrs)+1)*keys.AddressLen)
	copy(addrs[:keys.AddressLen], n.MainAddr[:]) // main address is always first
	for i, addr := range n.AltAddrs {
		copy(addrs[(i+1)*keys.AddressLen:], addr[:])
	}

	signedAddrs := n.Kp.Sk.Sign(addrs)                                // yes, actually works now
	encodedAddrs := base64.RawURLEncoding.EncodeToString(signedAddrs) // we might do ipv6/libp2p/port enocding or smth later
	return fmt.Sprintf("cofind:%s.%s", pk, encodedAddrs)
}

// unoptimised; debug
func (n Node) log(msg ...any) {
	pke := n.Kp.Pk.Encode()
	logId := pke[6:][:2]

	m := strings.ReplaceAll(fmt.Sprint(msg...), "\n", "\n     ")
	fmt.Printf("[%s]\n     %s\n", logId, m)
}

func (n *Node) handleMessage(am AnyMsg) {
	dm, err := am.Deserialise()
	if err != nil {
		n.log("Failed to deserialise message\n", err)
		return
	}

	switch m := dm.(type) {
	case mHi:
		n.log("Received hi message from peer\n", am.From.Pk.Encode(), "\n", am.From.MainAddr)

	case mStore:
		hash, err := StoreProgram(am.From.Pk, m.Name, m.Bundled)
		if err != nil {
			n.log("Failed to store program\n", err)
			break
		}

		// show result was successful
		res := mStoreResult{hash}
		n.send(am.From, res)

	case mStoreResult:
		n.log("Program storage successful\n", "Hash: ", hex.EncodeToString(m.Hash[:]))

	case mRun:
		n.log("Running program\n", "PK: ", m.Pk.Encode(), "\n", "Name: ", m.Name)

		switch tin := m.Input.(type) {
		case WebArgs:
			ret, err := StartWebProgram(m.Pk, m.Name, tin)
			if err != nil {
				n.log("Failed to run program\n", err)
				break
			}

			// serialise as json
			// TODO: i think we're serialising this twice??? figure out how to get it from somewhere else
			inputBytes, err := json.Marshal(tin)
			if err != nil {
				n.log("Failed to serialise input for hashing\n", err)
				break
			}

			// return result
			res := mRunResult{WebProgramType, m.Pk, m.Name, sha3.Sum256(inputBytes), ret}
			n.send(am.From, res)

		default:
			n.log("Unknown program type\n", m.Input.Type())
		}

	case mRunResult:
		h := InputName{m.Pk, m.Name, m.InputHash}
		if ch, ok := n.resultsWaitingName[h]; ok {
			ch <- m.Result
			delete(n.resultsWaitingName, h)
		} else {
			n.log("Received name result for unexpected program\n", m.Result)
		}

	default:
		// any unknown is dropped
		n.log("Unknown message type\n", am.Type, m)
	}
}

func (n *Node) seenPeer(p *keys.Peer) {
	if _, ok := n.Peers[p.Pk]; !ok {
		n.Peers[p.Pk] = p
	}
	n.Peers[p.Pk].LastSeen = time.Now()
}

func (n *Node) StoreProgram(pk keys.PK, name string, b []byte) (err error) {
	if _, err = StoreProgram(pk, name, b); err != nil {
		return
	}

	for _, peer := range n.Peers {
		m := mStore{name, b}

		if err = n.send(peer, m); err != nil {
			return
		}
	}

	return
}

// we don't have the program; ask peers for it
func (n *Node) peerRunName(pk keys.PK, name string, inputhash [32]byte, ptype ProgramType, input ProgramArgs) (res ProgramArgs, err error) {
	if len(n.Peers) == 0 {
		return nil, errors.New("no peers to run program")
	}

	h := InputName{pk, name, inputhash}
	ch := make(chan ProgramRets)
	n.resultsWaitingName[h] = ch

	for _, peer := range n.Peers {
		m := mRun{ptype, pk, name, input}

		if err = n.send(peer, m); err != nil {
			return
		}
	}

	res = <-ch
	delete(n.resultsWaitingName, h)
	close(ch)

	return
}

func (n *Node) receive() {
	for {
		rec := <-n.ReceiveRaw
		if !n.running {
			break
		}

		msg, err := rec.Decode(n.Kp)
		if err != nil {
			n.log("Failed to decode message\n", err)
			continue
		}

		n.seenPeer(msg.From)

		n.log(
			"Received ", len(msg.Body), "\n",
			"From ", msg.From.Pk.Encode(), "\n",
			"@ ", msg.From.MainAddr, "\n")

		n.handleMessage(msg)
	}
}

func (n *Node) RunWebProgram(pk keys.PK, name string, input WebArgs, useLocal bool) (res WebRets, err error) {
	if useLocal { // testing; to prevent 2 communication servers (from realising they're) using the same execution server
		if res, err = StartWebProgram(pk, name, input); err == nil {
			return // we have the program!
		}
	}

	// serialise as json
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return WebRets{}, err
	}

	r, err := n.peerRunName(pk, name, sha3.Sum256(inputBytes), WebProgramType, input)
	if err != nil {
		return
	} else if r.Type() != WebProgramType {
		return WebRets{}, errors.New("invalid program type")
	}

	return r.(WebRets), nil
}

func (n *Node) Start() {
	pke := n.Kp.Pk.Encode()
	n.running = true

	n.log(
		"Starting\n",
		"I'm ", pke, "\n",
		"My primary address is ", n.MainAddr, "\n",
		"I know ", len(n.Peers), " peers")

	// Receiver
	go n.receive()

	for _, peer := range n.Peers {
		n.log("Sending hi message to peer\n", peer.Pk.Encode(), "\n", peer.MainAddr)
		if err := n.send(peer, mHi{}); err != nil {
			n.log("Failed to send hi message to peer\n", err)
		}
	}
}

func (n *Node) Stop() {
	n.log("Stopping")
	n.running = false
	close(n.SendRaw)
	close(n.ReceiveRaw)
}
