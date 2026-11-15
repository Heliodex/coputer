package types

type (
	// TestArgs stores the arguments passed to a test program.
	TestArgs struct{}

	// TestRets stores the response returned from a test program.
	TestRets struct{}
)

func (TestArgs) Encode() []byte {
	return nil
}

// Type returns TestProgramType.
func (TestArgs) Type() ProgramType {
	return TestProgramType
}

func (TestRets) Encode() []byte {
	return nil
}

func (TestRets) Equal(TestRets) error {
	return nil
}

// Type returns TestProgramType.
func (TestRets) Type() ProgramType {
	return TestProgramType
}
