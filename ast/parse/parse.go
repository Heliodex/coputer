package main

import (
	"fmt"

	"github.com/Heliodex/coputer/ast/ast"
	"github.com/Heliodex/coputer/ast/lex"
)

type Mode uint8

const (
	NoCheck    Mode = iota // Do not perform any inference
	Nonstrict              // Unannotated symbols are any
	Strict                 // Unannotated symbols are inferred
	Definition             // Type definition module, has special parsing rules
)

type FragmentParseResumeSettings struct{}

type ParseOptions struct {
	allowDeclarationSyntax bool
	captureComments        bool
	parseFragment          *FragmentParseResumeSettings
	storeCstData           bool
	noErrorLimit           bool
}

type HotComment struct {
	header   bool
	location lex.Location
	content  string
}

type ParseResult struct {
	root  ast.AstStatBlock[ast.ASTNode]
	lines uint

	hotcomments []HotComment
	errors      []ParseError

	commentLocations []Comment
}
type ParseExprResult struct {
	root  AstExpr
	lines uint

	hotcomments []HotComment
	errors      []ParseError

	commentLocations []Comment
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Parse error at %v: %s", e.Location, e.Message)
}
