package keys

// base(d) on eknkc/basex

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

const (
	PKSize = 29
	SKSize = 32
)

type (
	PK [PKSize]byte
	SK [SKSize]byte
)

var (
	alphabet = [36]byte([]byte("0123456789abcdefghijklmnopqrstuvwxyz"))
	byteMap  = make(map[byte]uint, 36)
)

func init() {
	for i, v := range alphabet {
		byteMap[v] = uint(i)
	}
}

const (
	base           = uint(len(alphabet))
	PubStart       = "copub:"
	SecStart       = "cosec:" // lel
	pkStringLength = 5 * 9
	skStringLength = 5 * 10
)

type KeyType uint8

const (
	Public KeyType = iota
	Secret
)

func digitEncoding[T [pkStringLength]uint | [skStringLength]uint, K PK | SK](k K) (digits T) {
	var di int

	for i := 0; i < len(k); i++ {
		carry := uint(k[i])
		for j := range di {
			carry += digits[j] << 8
			digits[j] = carry % base
			carry /= base
		}
		for ; carry > 0; di++ {
			digits[di] = carry % base
			carry /= base
		}
	}
	return
}

func section(res []byte, size int) string {
	var b strings.Builder
	for i := range 5 {
		b.Write(res[i*size:][:size])
		if i < 4 {
			b.WriteByte('-')
		}
	}
	return b.String()
}

func (pk PK) EncodeNoPrefix() string {
	var res [pkStringLength]byte
	for i, v := range digitEncoding[[pkStringLength]uint](pk) {
		res[len(res)-i-1] = alphabet[v]
	}

	// 5 sections of 9 characters each
	return section(res[:], 9)
}

// Takes a public key and returns a string representation of it using a custom alphabet
func (pk PK) Encode() string {
	return PubStart + pk.EncodeNoPrefix()
}

// Takes a secret key and returns a string representation of it using a custom alphabet
func (sk SK) Encode() string {
	var res [skStringLength]byte
	for i, v := range digitEncoding[[skStringLength]uint](sk) {
		res[len(res)-i-1] = alphabet[v]
	}

	// 5 sections of 10 characters each
	return SecStart + section(res[:], 10)
}

var errInvalidKeyType = errors.New("invalid key type")

func parseKey[T PK | SK](bs []byte) (bf T, err error) {
	var bp int

	for _, v := range bs {
		carry, ok := byteMap[v]
		if !ok {
			return bf, fmt.Errorf("non-base character '%c'", v)
		}

		for j := range bp {
			carry += uint(bf[j]) * base
			bf[j] = byte(carry)
			carry >>= 8
		}
		for ; carry > 0; bp++ {
			bf[bp] = byte(carry)
			carry >>= 8
		}
	}
	return
}

// Takes a string representation of a public key and returns the byte representation of it
func DecodePKNoPrefix(key string) (bf PK, err error) {
	if expectedLength := pkStringLength + 4; len(key) != expectedLength {
		return PK{}, fmt.Errorf("invalid length: expected %d, got %d", expectedLength, len(key))
	}

	if bf, err = parseKey[PK]([]byte(key[:9] + key[10:19] + key[20:29] + key[30:39] + key[40:])); err != nil {
		return
	}
	slices.Reverse(bf[:])
	return
}

func DecodePK(key string) (PK, error) {
	if key[:len(PubStart)] != PubStart {
		return PK{}, errInvalidKeyType
	}

	return DecodePKNoPrefix(key[len(PubStart):])
}

// Takes a string representation of a secret key and returns the byte representation of it
func DecodeSKNoPrefix(key string) (bf SK, err error) {
	if expectedLength := skStringLength + 4; len(key) != expectedLength {
		return SK{}, fmt.Errorf("invalid length: expected %d, got %d", expectedLength, len(key))
	}

	if bf, err = parseKey[SK]([]byte(key[:10] + key[11:21] + key[22:32] + key[33:43] + key[44:])); err != nil {
		return
	}
	slices.Reverse(bf[:])
	return
}

func DecodeSK(key string) (SK, error) {
	if key[:len(SecStart)] != SecStart {
		return SK{}, errInvalidKeyType
	}

	return DecodeSKNoPrefix(key[len(SecStart):])
}
