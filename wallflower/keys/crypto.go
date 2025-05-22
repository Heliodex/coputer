package keys

import "golang.org/x/crypto/curve25519"

// ephemeral 32 keypair that everyone knows
var ZeroPK, ZeroSK = new([32]byte), new([32]byte)

func init() {
	curve25519.ScalarBaseMult(ZeroPK, ZeroSK)
}
