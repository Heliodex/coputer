package keys

import (
	"errors"
	"fmt"
	"time"

	"github.com/ChainSafe/go-schnorrkel"
)

type Keypair struct {
	pk PK
	sk SK
}

func GenerateKeys(threads uint8, found chan<- Keypair, done <-chan struct{} /* signal to stop*/) {
	var hashes, seconds uint64

	for range threads {
		go func() {
			for {
				// 🔥🔥 HOT PATH 🔥🔥
				skm, _ := schnorrkel.GenerateMiniSecretKey()
				sk := skm.ExpandEd25519().Encode()

				hashes++

				pk, err := SKtoPK(sk)
				if err != nil {
					continue
				}

				fmt.Println("found keypair")

				select {
				case found <- Keypair{pk, sk}:
					fmt.Println("sent keypair")
				case <-done:
					return
				}
			}
		}()
	}

	go func() {
		for {
			time.Sleep(time.Second * 3)
			seconds += 3
			fmt.Println("hashes per second:", hashes/seconds)

			select {
			case <-done:
				return
			default:
			}
		}
	}()
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
