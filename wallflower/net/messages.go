package net

import (
	"encoding/json"
	"errors"
	"fmt"

	. "github.com/Heliodex/coputer/litecode/types"
	"github.com/Heliodex/coputer/wallflower/keys"
)

type (
	EncryptedMsg []byte
	MessageType  = byte
)

const (
	// A message sent on startup, without any other information
	tHi MessageType = iota
	// A program to store
	tStore
	// Program was stored successfully
	tStoreResult
	// Input to a program, indexed by name and pubkey
	tRun
	// A result of a program execution, indexed by name and pubkey
	tRunResult
)

// sent messages
type SentMsg interface {
	Serialise() ([]byte, error)
}

func addType(t MessageType, m []byte) []byte {
	return append([]byte{t}, m...)
}

type mHi struct{}

func (m mHi) Serialise() ([]byte, error) {
	return []byte{tHi}, nil
}

type mStore struct {
	Name    string
	Bundled []byte
}

func (m mStore) Serialise() ([]byte, error) {
	nl := len(m.Name)
	if nl > 255 {
		return nil, errors.New("name too long")
	}

	b := make([]byte, 1, 1+nl+len(m.Bundled))
	b[0] = byte(nl)
	b = append(b, m.Name...)
	b = append(b, m.Bundled...)

	return addType(tStore, b), nil
}

type mStoreResult struct {
	Hash [32]byte
}

func (m mStoreResult) Serialise() ([]byte, error) {
	return addType(tStoreResult, m.Hash[:]), nil
}

type mRun struct {
	Type  ProgramType // 1
	Pk    keys.PK     // 29
	Name  string      // 1 + length
	Input ProgramArgs
}

func (m mRun) Serialise() (s []byte, err error) {
	in, err := json.Marshal(m.Input)
	if err != nil {
		return
	}

	b := make([]byte, 1, 1+keys.PKSize+1+len(m.Name)+len(in))
	b[0] = byte(m.Type)
	b = append(b, m.Pk[:]...)
	b = append(b, byte(len(m.Name)))
	b = append(b, m.Name...)
	b = append(b, in...)

	return addType(tRun, b), nil
}

type mRunResult struct {
	Type      ProgramType // 1
	Pk        keys.PK     // 29
	Name      string      // 1 + length
	InputHash [32]byte
	Result    ProgramRets
}

func (m mRunResult) Serialise() (s []byte, err error) {
	res, err := json.Marshal(m.Result)
	if err != nil {
		return
	}

	b := make([]byte, 1, 1+keys.PKSize+1+len(m.Name)+len(res))
	b[0] = byte(m.Type)
	b = append(b, m.Pk[:]...)
	b = append(b, byte(len(m.Name)))
	b = append(b, m.Name...)
	b = append(b, m.InputHash[:]...)
	b = append(b, res...)

	return addType(tRunResult, b), nil
}

type AnyMsg struct {
	From *keys.Peer
	Type MessageType
	Body []byte
}

func unmarshalInput(ptype ProgramType, rest []byte) (ProgramArgs, error) {
	switch ptype {
	case WebProgramType:
		var tin WebArgs
		if err := json.Unmarshal(rest, &tin); err != nil {
			return nil, fmt.Errorf("failed to unmarshal web args: %w", err)
		}
		return tin, nil
	}
	return nil, errors.New("unknown program args type")
}

func unmarshalResult(ptype ProgramType, rest []byte) (ProgramRets, error) {
	switch ptype {
	case WebProgramType:
		var tres WebRets
		if err := json.Unmarshal(rest, &tres); err != nil {
			return nil, fmt.Errorf("failed to unmarshal web result: %w", err)
		}
		return tres, nil
	}
	return nil, errors.New("unknown program results type")
}

func (m AnyMsg) Deserialise() (SentMsg, error) {
	switch m.Type {
	case tHi:
		return mHi{}, nil
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
	case tRun:
		ptype := ProgramType(m.Body[0])

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

		return mRun{ptype, pk, name, in}, nil
	case tRunResult:
		ptype := ProgramType(m.Body[0])

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

		return mRunResult{ptype, pk, name, inputhash, res}, nil
	}

	return nil, errors.New("unknown message type")
}
