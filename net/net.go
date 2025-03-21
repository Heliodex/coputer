package net

import (
	"crypto/sha3"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Heliodex/coputer/keys"
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
	} else if len(addrs)%keys.AddressLen != 0 {
		return nil, errors.New("invalid addresses part length")
	}

	addresses := make([]keys.Address, len(addrs)/keys.AddressLen)
	for i := range addresses {
		copy(addresses[i][:], addrs[i*keys.AddressLen:][:keys.AddressLen])
	}

	return &keys.Peer{Pk: pk, Addresses: addresses}, nil
}

func (e EncryptedMsg) Decode(kp keys.Keypair) (m AnyMsg, err error) {
	from, body, err := kp.Decrypt(e)
	if err != nil {
		return
	}

	return AnyMsg{
		From: &from,
		Type: MessageType(body[0]),
		Body: body[1:],
	}, nil
}

type Node struct {
	keys.ThisPeer

	Peers          map[keys.PK]*keys.Peer // known peers
	SendRaw        func(peer *keys.Peer, msg []byte) (err error)
	ReceiveRaw     <-chan EncryptedMsg
	resultsWaiting map[[32]byte]map[[32]byte]chan string // [program hash][input hash]
}

func (n *Node) send(pk keys.PK, sm SentMsg) (err error) {
	peer, ok := n.Peers[pk]
	if !ok {
		return errors.New("unknown peer")
	}

	ct, err := n.Encrypt(sm.Serialise(), pk)
	if err != nil {
		return
	}

	return n.SendRaw(peer, ct)
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
	fmt.Printf("[%s]\n     %s\n", logId, m)
}

func (n *Node) handleMessage(am AnyMsg) {
	switch m := am.Deserialise().(type) {
	case mMsg1:
		n.log("Received message: ", m.Body)

		res := mMsg1{"Hello, " + m.Body}
		n.send(am.From.Pk, res) // infinite loop can't be avoided if you take a route straight through what is known as

	case mStore:
		hash, err := StoreProgram(m.Bundled)
		if err != nil {
			n.log("Failed to store program\n", err)
			break
		}

		// show result was successful
		res := mStoreResult{hash}
		n.send(am.From.Pk, res)

	case mStoreResult:
		n.log("Program storage successful\n", "Hash: ", hex.EncodeToString(m.Hash[:]))

	case mRun:
		n.log("Running program\n", "Hash: ", hex.EncodeToString(m.Hash[:]))

		ret, err := RunProgram(m.Hash, m.Input)
		if err != nil {
			n.log("Failed to run program\n", err)
			break
		}

		// return result
		res := mRunResult{m.Hash, sha3.Sum256([]byte(m.Input)), ret}
		n.send(am.From.Pk, res)

	case mRunResult:
		if p, ok := n.resultsWaiting[m.Hash]; ok {
			if ch, ok := p[m.Hash]; ok {
				ch <- m.Result
				delete(p, m.Hash)
			} else {
				n.log("Received result for unknown input\n", m.Result)
			}
		} else {
			n.log("Received result for unexpected program\n", m.Result)
		}

	default:
		// any unknown is dropped
		n.log("Unknown message type\n", am.Type)
	}
}

func (n *Node) seenPeer(p *keys.Peer) {
	if _, ok := n.Peers[p.Pk]; !ok {
		n.Peers[p.Pk] = p
	}
	n.Peers[p.Pk].LastSeen = time.Now()
}

func (n *Node) StoreProgram(b []byte) (err error) {
	if _, err = StoreProgram(b); err != nil {
		return
	}

	for _, peer := range n.Peers {
		m := mStore{b}

		if err = n.send(peer.Pk, m); err != nil {
			return
		}
	}

	return
}

func (n *Node) RunProgram(hash [32]byte, input string) (res string, err error) {
	if res, err = RunProgram(hash, input); err == nil {
		return
	}

	inputhash := sha3.Sum256([]byte(input))

	ch := make(chan string)
	if _, ok := n.resultsWaiting[hash]; !ok {
		n.resultsWaiting[hash] = make(map[[32]byte]chan string)
	}
	n.resultsWaiting[hash][inputhash] = ch

	for _, peer := range n.Peers {
		m := mRun{hash, input}

		if err = n.send(peer.Pk, m); err != nil {
			return
		}
	}

	res = <-ch
	delete(n.resultsWaiting[hash], inputhash)
	close(ch)

	return
}

func (n *Node) Start() {
	pke := n.Kp.Pk.Encode()

	n.log(
		"Starting\n",
		"I'm ", pke, "\n",
		"My primary address is ", n.Addresses[0], "\n",
		"I know ", len(n.Peers), " peers")

	// Receiver
	for {
		rec := <-n.ReceiveRaw
		msg, err := rec.Decode(n.Kp)
		if err != nil {
			n.log("Failed to decode message\n", err)
			continue
		}

		n.seenPeer(msg.From)

		n.log(
			"Received ", len(msg.Body), "\n",
			"From ", msg.From.Pk.Encode(), "\n",
			"@ ", msg.From.Addresses[0], "\n")

		n.handleMessage(msg)
	}
}
