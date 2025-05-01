package types

// ProgramType represents the type of a program.
type ProgramType uint8

const (
	// TestProgramType represents the type of a test program.
	// Test programs are to be used for debugging and testing purposes only.
	TestProgramType ProgramType = iota
	// WebProgramType represents the type of a web program.
	WebProgramType
)

// ProgramArgs represents the arguments passed to a program.
type ProgramArgs interface {
	Type() ProgramType
}

// ProgramRets represents the response returned from a program.
type ProgramRets interface {
	Type() ProgramType
}
