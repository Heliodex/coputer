package keys

import (
	"errors"
	"fmt"

	"github.com/ChainSafe/go-schnorrkel"
)

type Keypair struct {
	pk PK
	sk SK
}

func GenerateKeys(threads uint8, stop <-chan struct{}) Keypair {
	found := make(chan Keypair)

	for range threads {
		go func() {
			for {
				// 🔥🔥 HOT PATH 🔥🔥
				skm, _ := schnorrkel.GenerateMiniSecretKey()
				sk := skm.ExpandEd25519().Encode()

				pk, err := SKtoPK(sk)
				if err != nil {
					continue
				}

				fmt.Println("found keypair")

				select {
				case found <- Keypair{pk, sk}:
					fmt.Println("sent keypair")
				case <-stop:
					return
				}
			}
		}()
	}

	return <-found
}

func SKtoPK(skBytes SK) (PK, error) {
	sk := schnorrkel.NewSecretKey(skBytes, [32]byte{})

	pk, err := sk.Public()
	if err != nil {
		return PK{}, err
	}

	pkb := pk.Encode()
	if pkb[0] != 0 || pkb[1] != 0 || pkb[2] != 0 {
		// not enough pow
		return PK{}, errors.New("private key maps to an invalid public key")
	}

	return PK(pkb[3:]), nil
}
