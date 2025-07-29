package main

import (
	"fmt"

	. "github.com/Heliodex/coputer/ast/lex"
)

type ParseError struct {
	Location Location
	Message  string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Parse error at %v: %s", e.Location, e.Message)
}
