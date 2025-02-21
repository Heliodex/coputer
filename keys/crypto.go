package keys

import "golang.org/x/crypto/curve25519"

// ephemeral 32 keypair that everyone knows
var (
	ZeroSK = new([32]byte)
	ZeroPK = func() (publicKey *[32]byte) {
		publicKey = new([32]byte)

		curve25519.ScalarBaseMult(publicKey, ZeroSK)
		return
	}()
)
