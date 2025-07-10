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

// Encoding can't really error, decoding can
// encoding could probably be cached if we care

// ProgramArgs represents the arguments passed to a program.
type ProgramArgs interface {
	Type() ProgramType
	Encode() []byte
}

// type EncodedArgs[T ProgramArgs] struct {
// 	Args    T
// 	Encoded []byte
// }

// func DecodeArgs[T ProgramArgs](encoded []byte) (args EncodedArgs[T], err error) {
// 	var rargs T
// 	if err = json.Unmarshal(encoded, &rargs); err != nil {
// 		return
// 	}

// 	return EncodedArgs[T]{rargs, encoded}, nil
// }

func DecodeArgs[T ProgramArgs](encoded []byte) (args T, err error) {
	return args, json.Unmarshal(encoded, &args)
}

// ProgramRets represents the response returned from a program.
type ProgramRets interface {
	// Equal(ProgramRets) error
	Type() ProgramType
	Encode() []byte
}

// type EncodedRets[T ProgramRets] struct {
// 	Rets    T
// 	Encoded []byte
// }

// func DecodeRets[T ProgramRets](encoded []byte) (rets EncodedRets[T], err error) {
// 	var rrets T
// 	if err = json.Unmarshal(encoded, &rrets); err != nil {
// 		return
// 	}

// 	return EncodedRets[T]{rrets, encoded}, nil
// }

func DecodeRets[T ProgramRets](encoded []byte) (args T, err error) {
	return args, json.Unmarshal(encoded, &args)
}
