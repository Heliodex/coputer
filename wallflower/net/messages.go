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
	MessageType  uint8
)

// more than 1 type per message someday?
const (
	_            MessageType = iota
	tStore                   // A program to store
	tStoreResult             // Program was stored successfully
	tRun                     // Hash of a program to execute
	tRunResult               // A result of a program execution
)

// sent messages
type SentMsg interface {
	Serialise() []byte
}

type AnyMsg struct {
	From *keys.Peer
	Type MessageType
	Body []byte
}

func (m AnyMsg) Deserialise() (SentMsg, error) {
	switch m.Type {
	case tStore:
		return mStore{m.Body}, nil
	case tStoreResult:
		var hash [32]byte
		copy(hash[:], m.Body)
		return mStoreResult{hash}, nil
	case tRun:
		ptype := types.ProgramType(m.Body[0])

		var hash [32]byte
		copy(hash[:], m.Body[1:][:32])
		rest := m.Body[33:]

		var in types.ProgramArgs
		switch ptype {
		case types.WebProgramType:
			var tin types.WebArgs
			if err := json.Unmarshal(rest, &tin); err != nil {
				return nil, fmt.Errorf("failed to unmarshal web args: %w", err)
			}
			in = tin
		default:
			return nil, errors.New("unknown program args type")
		}

		return mRun{ptype, hash, in}, nil
	case tRunResult:
		ptype := types.ProgramType(m.Body[0])

		var hash, inputhash [32]byte
		copy(hash[:], m.Body[1:][:32])
		copy(inputhash[:], m.Body[33:][:32])
		rest := m.Body[65:]

		var res types.ProgramRets

		switch ptype {
		case types.WebProgramType:
			var tres types.WebRets
			if err := json.Unmarshal(rest, &tres); err != nil {
				return nil, fmt.Errorf("failed to unmarshal web result: %w", err)
			}
			res = tres
		default:
			return nil, errors.New("unknown message type")
		}

		return mRunResult{ptype, hash, inputhash, res}, nil
	}

	return nil, errors.New("unknown message type")
}

func addType(t MessageType, m []byte) []byte {
	return append([]byte{byte(t)}, m...)
}

type mStore struct {
	Bundled []byte
}

func (m mStore) Serialise() []byte {
	return addType(tStore, m.Bundled)
}

type mStoreResult struct {
	Hash [32]byte
}

func (m mStoreResult) Serialise() []byte {
	return addType(tStoreResult, m.Hash[:])
}

type mRun struct {
	Type  types.ProgramType // 1
	Hash  [32]byte
	Input types.ProgramArgs
}

func (m mRun) Serialise() []byte {
	in, err := json.Marshal(m.Input)
	// TODO
	if err != nil {
		panic(err)
	}

	b := make([]byte, 1, 1+32+len(in))
	b[0] = byte(m.Type)
	b = append(b, m.Hash[:]...)
	b = append(b, in...)

	return addType(tRun, b)
}

type mRunResult struct {
	Type      types.ProgramType // 1
	Hash      [32]byte
	InputHash [32]byte
	Result    types.ProgramRets
}

func (m mRunResult) Serialise() []byte {
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

	return addType(tRunResult, b)
}
