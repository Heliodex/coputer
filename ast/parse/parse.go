package main

import (
	"fmt"

	"github.com/Heliodex/coputer/ast/ast"
	"github.com/Heliodex/coputer/ast/lex"
)

type ParseError struct {
	Location lex.Location
	Message  string
}

type HotComment struct {
	header   bool
	location lex.Location
	content  string
}

type Comment struct {
	Type     lex.LexemeType
	location lex.Location
}

type ParseResult struct {
	root  ast.StatBlock[ast.Node]
	lines uint

	hotcomments []HotComment
	errors      []ParseError

	commentLocations []Comment
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Parse error at %v: %s", e.Location, e.Message)
}
