package types

// TestArgs stores the arguments passed to a test program.
type TestArgs struct{}

// Type returns TestProgramType.
func (TestArgs) Type() ProgramType {
	return TestProgramType
}

func (TestArgs) Encode() []byte {
	return nil
}

// TestRets stores the response returned from a test program.
type TestRets struct{}

func (TestRets) Equal(TestRets) error {
	return nil
}

// Type returns TestProgramType.
func (TestRets) Type() ProgramType {
	return TestProgramType
}

func (TestRets) Encode() []byte {
	return nil
}
