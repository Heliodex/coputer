package net

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
)

type (
	EncryptedMsg []byte
	MessageType  = byte
)

// more than 1 type per message someday?
const (
	// A program to store
	tStore MessageType = iota
	// Program was stored successfully
	tStoreResult
	// Input to a program, indexed by hash
	tRunHash
	// A result of a program execution, indexed by hash
	tRunHashResult
	// Input to a program, indexed by name and pubkey
	tRunName
	// A result of a program execution, indexed by name and pubkey
	tRunNameResult
)

// sent messages
type SentMsg interface {
	Serialise() []byte
}

func addType(t MessageType, m []byte) []byte {
	return append([]byte{t}, m...)
}

type mStore struct {
	Name    string
	Bundled []byte
}

func (m mStore) Serialise() []byte {
	nl := len(m.Name)
	// TODO
	if nl > 255 {
		panic("name too long")
	}

	b := make([]byte, 1, 1+nl+len(m.Bundled))
	b[0] = byte(nl)
	b = append(b, m.Name...)
	b = append(b, m.Bundled...)

	return addType(tStore, b)
}

type mStoreResult struct {
	Hash [32]byte
}

func (m mStoreResult) Serialise() []byte {
	return addType(tStoreResult, m.Hash[:])
}

type mRunHash struct {
	Type  types.ProgramType // 1
	Hash  [32]byte
	Input types.ProgramArgs
}

func (m mRunHash) Serialise() []byte {
	in, err := json.Marshal(m.Input)
	// TODO
	if err != nil {
		panic(err)
	}

	b := make([]byte, 1, 1+32+len(in))
	b[0] = byte(m.Type)
	b = append(b, m.Hash[:]...)
	b = append(b, in...)

	return addType(tRunHash, b)
}

type mRunHashResult struct {
	Type      types.ProgramType // 1
	Hash      [32]byte
	InputHash [32]byte
	Result    types.ProgramRets
}

func (m mRunHashResult) Serialise() []byte {
	res, err := json.Marshal(m.Result)
	// TODO
	if err != nil {
		panic(err)
	}

	b := make([]byte, 1, 1+64+len(res))
	b[0] = byte(m.Type)
	b = append(b, m.Hash[:]...)
	b = append(b, m.InputHash[:]...)
	b = append(b, res...)

	return addType(tRunHashResult, b)
}

type mRunName struct {
	Type  types.ProgramType // 1
	Pk    keys.PK           // 29
	Name  string            // 1 + length
	Input types.ProgramArgs
}

func (m mRunName) Serialise() []byte {
	in, err := json.Marshal(m.Input)
	// TODO
	if err != nil {
		panic(err)
	}

	b := make([]byte, 1, 1+keys.PKSize+1+len(m.Name)+len(in))
	b[0] = byte(m.Type)
	b = append(b, m.Pk[:]...)
	b = append(b, byte(len(m.Name)))
	b = append(b, m.Name...)
	b = append(b, in...)

	return addType(tRunName, b)
}

type mRunNameResult struct {
	Type      types.ProgramType // 1
	Pk        keys.PK           // 29
	Name      string            // 1 + length
	InputHash [32]byte
	Result    types.ProgramRets
}

func (m mRunNameResult) Serialise() []byte {
	res, err := json.Marshal(m.Result)
	// TODO
	if err != nil {
		panic(err)
	}

	b := make([]byte, 1, 1+keys.PKSize+1+len(m.Name)+len(res))
	b[0] = byte(m.Type)
	b = append(b, m.Pk[:]...)
	b = append(b, byte(len(m.Name)))
	b = append(b, m.Name...)
	b = append(b, m.InputHash[:]...)
	b = append(b, res...)

	return addType(tRunNameResult, b)
}

type AnyMsg struct {
	From *keys.Peer
	Type MessageType
	Body []byte
}

func unmarshalInput(ptype types.ProgramType, rest []byte) (types.ProgramArgs, error) {
	switch ptype {
	case types.WebProgramType:
		var tin types.WebArgs
		if err := json.Unmarshal(rest, &tin); err != nil {
			return nil, fmt.Errorf("failed to unmarshal web args: %w", err)
		}
		return tin, nil
	}
	return nil, errors.New("unknown program args type")
}

func unmarshalResult(ptype types.ProgramType, rest []byte) (types.ProgramRets, error) {
	switch ptype {
	case types.WebProgramType:
		var tres types.WebRets
		if err := json.Unmarshal(rest, &tres); err != nil {
			return nil, fmt.Errorf("failed to unmarshal web result: %w", err)
		}
		return tres, nil
	}
	return nil, errors.New("unknown program results type")
}

func (m AnyMsg) Deserialise() (SentMsg, error) {
	switch m.Type {
	case tStore:
		nl := m.Body[0]
		if int(nl) > len(m.Body) || nl == 0 {
			return nil, errors.New("invalid message length")
		}

		name, bundled := m.Body[1:nl+1], m.Body[nl+1:]

		return mStore{string(name), bundled}, nil
	case tStoreResult:
		var hash [32]byte
		copy(hash[:], m.Body)
		return mStoreResult{hash}, nil
	case tRunHash:
		ptype := types.ProgramType(m.Body[0])

		var hash [32]byte
		copy(hash[:], m.Body[1:][:32])
		rest := m.Body[33:]

		in, err := unmarshalInput(ptype, rest)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal program args: %w", err)
		}

		return mRunHash{ptype, hash, in}, nil
	case tRunHashResult:
		ptype := types.ProgramType(m.Body[0])

		var hash, inputhash [32]byte
		copy(hash[:], m.Body[1:][:32])
		copy(inputhash[:], m.Body[33:][:32])
		rest := m.Body[65:]

		res, err := unmarshalResult(ptype, rest)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal program result: %w", err)
		}

		return mRunHashResult{ptype, hash, inputhash, res}, nil
	case tRunName:
		ptype := types.ProgramType(m.Body[0])

		var pk keys.PK
		copy(pk[:], m.Body[1:][:29])
		nl, rest := m.Body[30], m.Body[31:]
		if int(nl) > len(m.Body) || nl == 0 {
			return nil, errors.New("invalid message length")
		}
		name, rest := string(rest[:nl]), rest[nl:]

		in, err := unmarshalInput(ptype, rest)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal program args: %w", err)
		}

		return mRunName{ptype, pk, name, in}, nil
	case tRunNameResult:
		ptype := types.ProgramType(m.Body[0])

		var pk keys.PK
		copy(pk[:], m.Body[1:][:29])
		nl, rest := m.Body[30], m.Body[31:]
		if int(nl) > len(m.Body) || nl == 0 {
			return nil, errors.New("invalid message length")
		}
		name, rest := string(rest[:nl]), rest[nl:]

		var inputhash [32]byte
		copy(inputhash[:], rest[:32])
		rest = rest[32:]

		res, err := unmarshalResult(ptype, rest)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal program result: %w", err)
		}

		return mRunNameResult{ptype, pk, name, inputhash, res}, nil
	}

	return nil, errors.New("unknown message type")
}
