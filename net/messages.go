package net

import "github.com/Heliodex/coputer/keys"

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
		return mRun{hash, string(m.Body[32:])}
	case tRunResult:
		var hash, inputhash [32]byte
		copy(hash[:], m.Body[:32])
		copy(inputhash[:], m.Body[32:64])

		return mRunResult{hash, inputhash, string(m.Body[64:])}
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
	Input string
}

func (m mRun) Serialise() []byte {
	return addType(tRun, append(m.Hash[:], []byte(m.Input)...))
}

type mRunResult struct {
	Hash      [32]byte
	InputHash [32]byte
	Result    string
}

func (m mRunResult) Serialise() []byte {
	s := make([]byte, 0, 64+len(m.Result))
	copy(s, m.Hash[:])
	copy(s[32:], m.InputHash[:])
	copy(s[64:], []byte(m.Result))

	return addType(tRunResult, s)
}
