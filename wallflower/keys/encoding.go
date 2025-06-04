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
	byteMap  = make(map[byte]int, 36)
)

func init() {
	for i, v := range alphabet {
		byteMap[v] = i
	}
}

const (
	base           = len(alphabet)
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

func (s PK) EncodeNoPrefix() string {
	var digits [pkStringLength]int
	di := 1

	for i := range s {
		carry := int(s[i])
		for j := range di {
			carry += digits[j] << 8
			digits[j] = carry % base
			carry /= base
		}
		for carry > 0 {
			digits[di] = carry % base
			di++
			carry /= base
		}
	}

	var res [pkStringLength]byte
	for i, v := range digits {
		res[len(res)-i-1] = alphabet[v]
	}

	// 5 sections of 9 characters each
	var b strings.Builder
	for i := range 5 {
		b.Write(res[i*9:][:9])
		if i < 4 {
			b.WriteByte('-')
		}
	}

	return b.String()
}

// Takes a public key and returns a string representation of it using a custom alphabet
func (s PK) Encode() string {
	return PubStart + s.EncodeNoPrefix()
}

// Takes a secret key and returns a string representation of it using a custom alphabet
func (s SK) Encode() string {
	var digits [skStringLength]int
	di := 1

	for i := range s {
		carry := int(s[i])
		for j := range di {
			carry += digits[j] << 8
			digits[j] = carry % base
			carry /= base
		}
		for carry > 0 {
			digits[di] = carry % base
			di++
			carry /= base
		}
	}

	var res [skStringLength]byte
	for i, v := range digits {
		res[len(res)-i-1] = alphabet[v]
	}

	// 5 sections of 10 characters each
	var b strings.Builder
	b.WriteString(SecStart)

	for i := range 5 {
		b.Write(res[i*10:][:10])
		if i < 4 {
			b.WriteByte('-')
		}
	}

	return b.String()
}

// Takes a string representation of a public key and returns the byte representation of it
func DecodePK(key string) (bs PK, err error) {
	if expectedLength := pkStringLength + 10; len(key) != expectedLength {
		return PK{}, fmt.Errorf("invalid length: expected %d, got %d", expectedLength, len(key))
	} else if key[:6] != PubStart {
		return PK{}, errors.New("invalid key type")
	}

	bp := 1

	for _, v := range []byte(key[6:15] + key[16:25] + key[26:35] + key[36:45] + key[46:]) {
		carry, ok := byteMap[v]
		if !ok {
			return PK{}, fmt.Errorf("non-base character '%c'", v)
		}

		for j := range bs {
			carry += int(bs[j]) * base
			bs[j] = byte(carry)
			carry >>= 8
		}
		for carry > 0 {
			bs[bp] = byte(carry)
			bp++
			carry >>= 8
		}
	}

	slices.Reverse(bs[:])

	return
}

// Takes a string representation of a secret key and returns the byte representation of it
func DecodeSK(key string) (bs SK, err error) {
	if expectedLength := skStringLength + 10; len(key) != expectedLength {
		return SK{}, fmt.Errorf("invalid length: expected %d, got %d", expectedLength, len(key))
	} else if key[:6] != SecStart {
		return SK{}, errors.New("invalid key type")
	}

	bp := 1

	for _, v := range []byte(key[6:16] + key[17:27] + key[28:38] + key[39:49] + key[50:]) {
		carry, ok := byteMap[v]
		if !ok {
			return SK{}, errors.New("non-base character")
		}

		for j := range bp {
			carry += int(bs[j]) * base
			bs[j] = byte(carry)
			carry >>= 8
		}
		for carry > 0 {
			bs[bp] = byte(carry)
			bp++
			carry >>= 8
		}
	}

	slices.Reverse(bs[:])

	return
}
