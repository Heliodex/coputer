package keys

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"slices"
	"time"

	"golang.org/x/crypto/nacl/box"
)

type Keypair struct {
	Pk PK
	Sk SK
}

const AddressLen = 16

type Address [AddressLen]byte // can be whatever (probably an ipv6 lel)

type Peer[T any] struct {
	Pk       PK
	MainAddr Address
	AltAddrs []Address
	LastSeen time.Time
	Transfer T
}

type ThisPeer[T any] struct {
	Peer[T]
	Kp Keypair
}

func (p Peer[T]) Equals(p2 Peer[T]) bool {
	return p.Pk == p2.Pk && p.MainAddr == p2.MainAddr && slices.Equal(p.AltAddrs, p2.AltAddrs)
}

func KeyWorker(found chan<- Keypair) {
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

		found <- kp
	}
}

func GenerateKeys(threads int) (found chan Keypair) {
	found = make(chan Keypair)

	for range threads {
		go KeyWorker(found)
	}

	return
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

const (
	keySize   = 29 + 1
	keyEnc    = keySize + box.AnonymousOverhead
	chunkEnc  = 1<<16 - 1
	chunkSize = chunkEnc - box.Overhead
)

// We could do real anonymous messages someday but having keys for every message at the moment is valuable

// --- Sender pk [29]
// --- Address count [1]
// anonymously encrypted [48] with recipient pk [total 78]

// --- Addresses [16]...
// encrypted [16] with recipient pk

// then the chunks (split message into 60-or-so kb parts)
// --- Actual message [up to 65519 (chunkSize)]
// encrypted [16] with recipient pk [total up to 65535 (chunkEnc)]

func encryptKey[T any](p ThisPeer[T], pk *[32]byte) ([]byte, error) {
	c := make([]byte, keySize)

	// Sender pk [29] + Address count [1]
	copy(c, p.Kp.Pk[:])
	c[len(c)-1] = byte(len(p.AltAddrs))

	// anonymously encrypted [48] with recipient pk
	return box.SealAnonymous(nil, c, pk, nil)
}

func encryptAddresses[T any](p ThisPeer[T], pk *[32]byte, sk [32]byte) []byte {
	c := p.MainAddr[:] // main address is always first

	// Addresses [16]...
	for _, addr := range p.AltAddrs {
		c = append(c, addr[:]...)
	}

	// encrypted [16] with recipient pk
	return box.Seal(nil, c, ZeroNonce, pk, &sk)
}

// chunks are chunked because not having to hold the whole thing in memory at the same time (?)
// we cold do larger chunks but messages larger than 64k will probably be rare (apart from file transfers? large responses? I'll get to tis when I do gateway implementation soon)
func (p ThisPeer[T]) Encrypt(msg []byte, to PK) (out []byte, err error) {
	pk := new([32]byte)
	copy(pk[3:], to[:])
	sk := [32]byte(p.Kp.Sk)

	// Address count [1] (alternates only)
	addrCount := len(p.AltAddrs)
	if addrCount > 255 {
		return nil, errors.New("too many addresses")
	}

	addrsSize := AddressLen * addrCount
	addrsEnc := addrsSize + box.Overhead

	l := len(msg)

	chunkCount := l / chunkSize
	if l%chunkSize != 0 {
		chunkCount++
	}
	chunksEnc := l + chunkCount*box.Overhead

	out = make([]byte, 0, keyEnc+addrsEnc+chunksEnc)

	// encryption
	enc, err := encryptKey(p, pk)
	if err != nil {
		return nil, err
	}
	out = append(out, enc...)
	out = append(out, encryptAddresses(p, pk, sk)...)

	// chunking time
	for i := range chunkCount {
		c := msg[i*chunkSize : min((i+1)*chunkSize, l)]

		enc := box.Seal(nil, c, ZeroNonce, pk, &sk)

		// add it to the message
		out = append(out, enc...)
	}

	return out, nil
}

func decryptKey(encryptedKey []byte, pk *[32]byte, sk [32]byte) (peerpk PK, addrcount int, ok bool) {
	// anonymously encrypted [48] with recipient pk
	dec, ok := box.OpenAnonymous(nil, encryptedKey, pk, &sk)
	if !ok {
		return
	}

	return PK(dec[:29]), int(dec[29]), true
}

func decryptAddrs(encryptedAddrs []byte, peerpk *[32]byte, sk [32]byte) (addrs []Address, ok bool) {
	// encrypted [16] with recipient pk
	dec, ok := box.Open(nil, encryptedAddrs, ZeroNonce, peerpk, &sk)
	if !ok || len(dec)%AddressLen != 0 {
		return
	}

	addrs = make([]Address, len(dec)/AddressLen)
	for i := range addrs {
		copy(addrs[i][:], dec[i*AddressLen:][:AddressLen])
	}

	return addrs, true
}

// todo: version cyphertexts
func Decrypt[T any](kp Keypair, emsg []byte) (from Peer[T], msg []byte, err error) {
	pk := new([32]byte)
	copy(pk[3:], kp.Pk[:])
	sk := [32]byte(kp.Sk)

	// Sender pk [29] + Address count [1]
	if len(emsg) < keyEnc+1 {
		return Peer[T]{}, nil, fmt.Errorf("message too short (%d)", len(emsg))
	}

	encryptedKey, ct := emsg[:keyEnc], emsg[keyEnc:]
	ppk, addrCount, ok := decryptKey(encryptedKey, pk, sk)
	if !ok {
		return Peer[T]{}, nil, errors.New("key decryption failed")
	}

	from.Pk = ppk

	peerpk := new([32]byte)
	copy(peerpk[3:], ppk[:])

	addrsSize := 16 * (addrCount + 1)
	addrsEnc := addrsSize + box.Overhead

	// Addresses [16]...
	encryptedAddrs, ct := ct[:addrsEnc], ct[addrsEnc:]
	addrs, ok := decryptAddrs(encryptedAddrs, peerpk, sk)
	if !ok {
		return Peer[T]{}, nil, errors.New("addresses decryption failed")
	}

	if len(addrs) == 0 {
		return Peer[T]{}, nil, errors.New("no addresses found")
	}

	from.MainAddr, from.AltAddrs = addrs[0], addrs[1:]

	// chunking time
	chunksCount := len(ct) / chunkEnc
	if len(ct)%chunkEnc != 0 {
		chunksCount++
	}

	msg = make([]byte, 0, len(ct)-chunksCount*box.Overhead)

	var chunk []byte
	for len(ct) > 0 {
		clen := min(chunkEnc, len(ct)) // we don't actually need to know the chunk size
		chunk, ct = ct[:clen], ct[clen:]
		dec, ok := box.Open(nil, chunk, ZeroNonce, peerpk, &sk)

		if !ok {
			return Peer[T]{}, nil, errors.New("chunk decryption failed")
		}

		msg = append(msg, dec...)
	}

	return
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
