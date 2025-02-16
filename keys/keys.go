package keys

import (
	"errors"

	"github.com/ChainSafe/go-schnorrkel"
)

type Keypair struct {
	Pk  PK
	Sk  SK
	spk *schnorrkel.PublicKey
	ssk *schnorrkel.SecretKey
}

func GenerateKeys(threads uint8, stop <-chan struct{}) Keypair {
	found := make(chan Keypair)

	for range threads {
		go func() {
			for {
				// 🔥🔥 HOT PATH 🔥🔥
				// ~45000 hashes/core/s on my machine, about 6 minutes (1 rap god) to find a keypair with 1 thread
				skm, _ := schnorrkel.GenerateMiniSecretKey()

				kp, err := keypair(skm.ExpandEd25519().Encode())
				if err != nil {
					continue
				}

				select {
				case found <- kp:
				case <-stop:
					return
				}
			}
		}()
	}

	return <-found
}

func keypair(s SK) (kp Keypair, err error) {
	sk := schnorrkel.NewSecretKey(s, [32]byte{}) // nonce unused

	pk, err := sk.Public()
	if err != nil {
		return
	}

	pkb := pk.Encode()
	if pkb[0] != 0 || pkb[1] != 0 || pkb[2] != 0 {
		// not enough pow
		return Keypair{}, errors.New("private key maps to an invalid public key")
	}

	return Keypair{PK(pkb[3:]), s, pk, sk}, nil
}

var domain = []byte("coputer")

func (kp Keypair) Sign(msg []byte) (s [64]byte, err error) {
	ctx := schnorrkel.NewSigningContext(domain, msg)

	sig, err := kp.ssk.Sign(ctx)
	if err != nil {
		return
	}

	return sig.Encode(), nil
}

func (pk PK) Verify(msg []byte, sig [64]byte) (verified bool, err error) {
	ctx := schnorrkel.NewSigningContext(domain, msg)

	pkb := [32]byte{}
	copy(pkb[3:], pk[:])

	spk := &schnorrkel.PublicKey{}
	spk.Decode(pkb)

	s := &schnorrkel.Signature{}
	s.Decode(sig)

	return spk.Verify(s, ctx)
}

func (kp Keypair) Decrypt(ct []byte) (msg []byte, err error) {
	return
}

func (pk PK) Encrypt(msg []byte)(ct []byte, err error) {
	return
}
