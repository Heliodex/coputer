package keys

import (
	"bytes"
	"crypto/rand"
	"errors"
	"slices"

	"golang.org/x/crypto/nacl/box"
)

type Keypair struct {
	Pk PK
	Sk SK
}

const AddressLen = 16

type Address [AddressLen]byte // can be whatever (probably an ipv6 lel)

type Peer struct {
	Pk        PK
	Addresses []Address
}

type ThisPeer struct {
	Peer
	Kp Keypair
}

func (p Peer) Equals(p2 Peer) bool {
	return p.Pk == p2.Pk && slices.Equal(p.Addresses, p2.Addresses)
}

func KeyWorker(found chan<- Keypair, stop <-chan struct{}) {
	for {
		// 🔥🔥 HOT PATH 🔥🔥
		// ~21000 hashes/core/s on my machine, about 12 minutes (2 rap gods) to find a keypair with 1 thread
		pk, sk, err := box.GenerateKey(rand.Reader)
		if err != nil {
			panic(err)
		}

		kp, err := keypair(*pk, *sk)
		if err != nil {
			continue
		}

		select {
		case found <- kp:
		case <-stop:
			return
		}
	}
}

func GenerateKeys(threads uint8, stop <-chan struct{}) Keypair {
	found := make(chan Keypair)
	for range threads {
		go KeyWorker(found, stop)
	}

	return <-found
}

func keypair(pk, sk [32]byte) (kp Keypair, err error) {
	if pk[0] != 0 || pk[1] != 0 || pk[2] != 0 {
		// not enough pow
		return Keypair{}, errors.New("private key maps to an invalid public key")
	}

	return Keypair{PK(pk[3:]), sk}, nil
}

func KeypairSK(sk [32]byte) (kp Keypair, err error) {
	pk, _, err := box.GenerateKey(bytes.NewReader(sk[:]))
	if err != nil { // doesn't ever error as sk is always of sufficient size
		return
	}

	return keypair(*pk, sk)
}

// Who needs nonces anyway? Every request is idempotent becaue every computation is deterministic.
// Exactly-once delivery 4 tha win
var ZeroNonce = new([24]byte)

const EncPKLen = PKSize + box.AnonymousOverhead

// todo: version cyphertexts
func (kp Keypair) Decrypt(ct []byte) (from Peer, msg []byte, ok bool) {
	pk := new([32]byte)
	copy(pk[3:], kp.Pk[:])
	sk := [32]byte(kp.Sk)

	// first decrypt sender public key
	encpk, ct := ct[:EncPKLen], ct[EncPKLen:]

	interpk, ok := box.OpenAnonymous(nil, encpk, pk, &sk)
	if !ok {
		return
	}

	// we now know whom the message is from
	peerpk := new([32]byte)
	copy(peerpk[3:], interpk)

	// next decrypt ct with known peer pk
	inter, ok := box.Open(nil, ct, ZeroNonce, peerpk, &sk)
	if !ok {
		return
	}

	// inter is 1 byte of number of addresses and rest is addresses and message
	numAddrs, inter := inter[0], inter[1:]
	allAddrsLen := int(numAddrs) * AddressLen
	allAddrs, msg := inter[:allAddrsLen], inter[allAddrsLen:]

	addresses := make([]Address, numAddrs)
	for i := range addresses {
		copy(addresses[i][:], allAddrs[i*AddressLen:][:AddressLen])
	}

	return Peer{PK(peerpk[3:]), addresses}, msg, true
}

// We could do real anonymous messages someday but having keys for every message at the moment is valuable

// --- Sender pk [29]
// anonymously encrypted [48] with recipient pk [total 77]
// --- Number of addresses [1]
// --- Addresses [16]...
// --- Actual message [...]
// encrypted [16] with recipient pk
func (p ThisPeer) Encrypt(msg []byte, to PK) (out []byte, err error) {
	interLen := 1 + AddressLen*len(p.Addresses) + len(msg)
	outLen := EncPKLen + interLen + box.Overhead

	out = make([]byte, 0, outLen)

	// Sender pk [29]
	pk := new([32]byte)
	copy(pk[3:], to[:])

	// anonymously encrypted [48] with recipient pk [total 77]
	encpk, err := box.SealAnonymous(nil, p.Kp.Pk[:], pk, nil)
	if err != nil {
		return
	}

	// add it to the message
	out = append(out, encpk...)

	inter := make([]byte, 1, interLen)

	// Number of addresses [1]
	inter[0] = byte(len(p.Addresses)) // if it's over 255 gfy

	// Addresses [16]...
	for _, addr := range p.Addresses {
		inter = append(inter, addr[:]...)
	}

	// Actual message [...]
	inter = append(inter, msg...)

	// encrypted [16] with recipient pk
	sk := [32]byte(p.Kp.Sk)
	encinter := box.Seal(nil, inter, ZeroNonce, pk, &sk)

	return append(out, encinter...), nil
}

// fake signatures with encryption
// because at the moment, signing with ed25519 with key conversion from curve25519 is a pain, converting from existing curve25519 keys to ed25519 is even worse, and while using ristretto255 for everything would be nice, that would require me rolling my own everything which would actually be not nice, and also I'm pissy that I'd have to use 64-byte private keys with edwards curves
// this would kill any competent cryptographer from a mile away lmao

func (sk SK) Sign(msg []byte) []byte {
	skb := new([32]byte)
	copy(skb[:], sk[:])

	// encrypt to a pk everyone knows with our sk
	return box.Seal(nil, msg, ZeroNonce, ZeroPK, skb)
}

func (pk PK) Verify(sig []byte) (msg []byte, ok bool) {
	pkb := new([32]byte)
	copy(pkb[3:], pk[:])

	// decrypt with a sk everyone knows and their pk
	return box.Open(nil, sig, ZeroNonce, pkb, ZeroSK)
}
