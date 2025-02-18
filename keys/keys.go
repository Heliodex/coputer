package keys

import (
	"bytes"
	"crypto/rand"
	"errors"

	"golang.org/x/crypto/nacl/box"
)

type Keypair struct {
	Pk PK
	Sk SK
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

// todo: version cyphertexts
func (kp Keypair) Decrypt(ct []byte) (msg []byte, from PK, ok bool) {
	pk := new([32]byte)
	copy(pk[3:], kp.Pk[:])

	// first decrypt anonymous message
	sk := [32]byte(kp.Sk)
	inter, ok := box.OpenAnonymous(nil, ct, pk, &sk) // removes 48 bytes (includes ephemeral 32-byte pk)
	if !ok {
		return
	}

	// inter is 29 bytes of pk and rest is ciphertext
	interpk, interct := inter[:29], inter[29:] // removes 29 bytes

	// we now know whom the message is from
	peerpk := new([32]byte)
	copy(peerpk[3:], interpk)

	// next decrypt with known peer pk (could use a signature here)
	msg, ok = box.Open(nil, interct, ZeroNonce, peerpk, &sk) // removes 16 bytes

	return msg, PK(interpk), ok
}

// encryption adds total 93 bytes of overhead
// We could do real anonymous messages someday but having keys for every message at the moment is valuable
func (kp Keypair) Encrypt(msg []byte, to PK) ([]byte, error) {
	pk := new([32]byte)
	copy(pk[3:], to[:])

	// first encrypt message with our own pk (could use a signature here)
	sk := [32]byte(kp.Sk)
	inter := box.Seal(nil, msg, ZeroNonce, pk, &sk) // adds 16 bytes (doesn't include 32-byte pk)

	// inter is 29 bytes of pk and rest is ciphertext
	inter = append(kp.Pk[:], inter...) // adds 29 bytes

	// next encrypt inter anonymously
	return box.SealAnonymous(nil, inter, pk, nil) // adds 48 bytes (includes ephemeral 32-byte pk)
}
