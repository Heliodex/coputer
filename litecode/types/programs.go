package types

import "encoding/json"

// ProgramType represents the type of a program.
type ProgramType uint8

const (
	// TestProgramType represents the type of a test program.
	// Test programs are to be used for debugging and testing purposes only.
	TestProgramType ProgramType = iota
	// WebProgramType represents the type of a web program.
	WebProgramType
)

// TODO: either merge or distinguish these

// ProgramArgs represents the arguments passed to a program.
type ProgramArgs interface {
	Type() ProgramType
	Encode() ([]byte, error)
}

func DecodeArgs[T ProgramArgs](encoded []byte) (args T, err error) {
	return args, json.Unmarshal(encoded, &args)
}

// ProgramRets represents the response returned from a program.
type ProgramRets interface {
	// Equal(ProgramRets) error
	Type() ProgramType
	Encode() ([]byte, error)
}

func DecodeRets[T ProgramRets](encoded []byte) (args T, err error) {
	return args, json.Unmarshal(encoded, &args)
}
