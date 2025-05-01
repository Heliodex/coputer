package types

// TestArgs stores the arguments passed to a test program.
type TestArgs struct{}

// Type returns TestProgramType.
func (TestArgs) Type() ProgramType {
	return TestProgramType
}

// TestRets stores the response returned from a test program.
type TestRets struct{}

func (r1 TestRets) Equal(r2 TestRets) error {
	return nil
}

// Type returns TestProgramType.
func (TestRets) Type() ProgramType {
	return TestProgramType
}
