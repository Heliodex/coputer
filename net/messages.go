package net

import (
	"encoding/json"

	"github.com/Heliodex/coputer/keys"
	"github.com/Heliodex/coputer/litecode/vm"
)

type (
	EncryptedMsg []byte
	MessageType  uint8
)

// more than 1 type per message someday?
const (
	tMsg1        MessageType = iota
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

func (m AnyMsg) Deserialise() (msg SentMsg) {
	switch m.Type {
	case tMsg1:
		return mMsg1{string(m.Body)}
	case tStore:
		return mStore{m.Body}
	case tStoreResult:
		var hash [32]byte
		copy(hash[:], m.Body)
		return mStoreResult{hash}
	case tRun:
		var hash [32]byte
		copy(hash[:], m.Body[:32])
		rest := m.Body[32:]

		var input vm.ProgramArgs
		// decode json
		if err := json.Unmarshal(rest, &input); err != nil {
			return
		}

		return mRun{hash, input}
	case tRunResult:
		var hash, inputhash [32]byte
		copy(hash[:], m.Body[:32])
		copy(inputhash[:], m.Body[32:64])

		var res vm.ProgramRets
		if err := json.Unmarshal(m.Body[64:], &res); err != nil {
			return
		}

		return mRunResult{hash, inputhash, res}
	}

	return
}

func addType(t MessageType, m []byte) []byte {
	return append([]byte{byte(t)}, m...)
}

type mMsg1 struct {
	Body string
}

func (m mMsg1) Serialise() []byte {
	return addType(tMsg1, []byte(m.Body))
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
	Hash  [32]byte
	Input vm.ProgramArgs
}

func (m mRun) Serialise() []byte {
	input, err := json.Marshal(m.Input)
	// TODO
	if err != nil {
		panic(err)
	}

	return addType(tRun, append(m.Hash[:], input...))
}

type mRunResult struct {
	Hash      [32]byte
	InputHash [32]byte
	Result    vm.ProgramRets
}

func (m mRunResult) Serialise() []byte {
	s := make([]byte, 0, 64)
	copy(s, m.Hash[:])
	copy(s[32:], m.InputHash[:])

	res, err := json.Marshal(m.Result)
	// TODO
	if err != nil {
		panic(err)
	}

	return addType(tRunResult, append(s, res...))
}
