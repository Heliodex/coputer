package keys

// base(d) on eknkc/basex

import (
	"errors"
	"fmt"
	"slices"
)

type (
	PK [29]byte
	SK [32]byte
)

func alphabetToMap(a [36]byte) (m map[byte]byte) {
	m = make(map[byte]byte, 36)
	for i, r := range a {
		m[r] = byte(i)
	}
	return
}

var (
	alphabet = [36]byte{
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'a', 'b',
		'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
		'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	}
	byteMap = alphabetToMap(alphabet)
)

const (
	base           = len(alphabet)
	pkStringLength = 45
	skStringLength = 50
	PubStart       = "copub:"
	SecStart       = "cosec:" // lel
)

type KeyType uint8

const (
	Public KeyType = iota
	Secret
)

// Takes a public key and returns a string representation of it using a custom alphabet
func (s PK) Encode() (formatted string) {
	digits := make([]int, 1, pkStringLength)

	for i := range s {
		carry := int(s[i])
		for j := range digits {
			carry += digits[j] << 8
			digits[j] = carry % base
			carry /= base
		}
		for carry > 0 {
			digits = append(digits, carry%base)
			carry /= base
		}
	}

	var res []byte
	for q := len(digits) - 1; q >= 0; q-- {
		res = append(res, alphabet[digits[q]])
	}

	// Pad with first characters to reach the ideal length
	for range pkStringLength - len(res) {
		res = append([]byte{alphabet[0]}, res...)
	}

	// 5 sections of 9 characters each
	formatted = PubStart

	for i := range 5 {
		formatted += string(res[i*9 : (i+1)*9])
		if i < 4 {
			formatted += "-"
		}
	}

	return formatted
}

// Takes a secret key and returns a string representation of it using a custom alphabet
func (s SK) Encode() (formatted string) {
	digits := make([]int, 1, skStringLength)

	for i := range s {
		carry := int(s[i])
		for j := range digits {
			carry += digits[j] << 8
			digits[j] = carry % base
			carry /= base
		}
		for carry > 0 {
			digits = append(digits, carry%base)
			carry /= base
		}
	}

	var res []byte
	for q := len(digits) - 1; q >= 0; q-- {
		res = append(res, alphabet[digits[q]])
	}

	// Pad with first characters to reach the ideal length
	for range skStringLength - len(res) {
		res = append([]byte{alphabet[0]}, res...)
	}

	// 5 sections of 10 characters each
	formatted = SecStart

	for i := range 5 {
		formatted += string(res[i*10 : (i+1)*10])
		if i < 4 {
			formatted += "-"
		}
	}

	return
}

// Takes a string representation of a public key and returns the byte representation of it
func DecodePK(key string) (PK, error) {
	expectedLength := pkStringLength + 10
	if len(key) != expectedLength {
		return PK{}, fmt.Errorf("invalid length: expected %d, got %d", expectedLength, len(key))
	} else if key[:6] != PubStart {
		return PK{}, errors.New("invalid key type")
	}

	bytes := []byte{0}
	raw := key[6:16] + key[17:26] + key[27:36] + key[37:46] + key[47:]
	for _, v := range []byte(raw) {
		value, ok := byteMap[v]
		if !ok {
			return PK{}, errors.New("non-base character")
		}

		carry := int(value)

		for j := range bytes {
			carry += int(bytes[j]) * base
			bytes[j] = byte(carry)
			carry >>= 8
		}

		for carry > 0 {
			bytes = append(bytes, byte(carry))
			carry >>= 8
		}
	}

	pkb := PK(bytes)
	slices.Reverse(pkb[:])

	return pkb, nil
}

// Takes a string representation of a secret key and returns the byte representation of it
func DecodeSK(key string) (SK, error) {
	expectedLength := skStringLength + 10
	if len(key) != expectedLength {
		return SK{}, fmt.Errorf("invalid length: expected %d, got %d", expectedLength, len(key))
	} else if key[:6] != SecStart {
		return SK{}, errors.New("invalid key type")
	}

	bytes := []byte{0}
	raw := key[6:17] + key[18:28] + key[29:39] + key[40:50] + key[51:]
	for _, v := range []byte(raw) {
		value, ok := byteMap[v]
		if !ok {
			return SK{}, errors.New("non-base character")
		}

		carry := int(value)

		for j := range bytes {
			carry += int(bytes[j]) * base
			bytes[j] = byte(carry)
			carry >>= 8
		}

		for carry > 0 {
			bytes = append(bytes, byte(carry))
			carry >>= 8
		}
	}


	skb := SK(bytes)
	slices.Reverse(skb[:])

	return skb, nil
}
