package main

import (
	"fmt"
	"strings"

	"github.com/Heliodex/coputer/ast/lex"
)

// bindings

type Binding struct {
	*NodeLoc
	Name          lex.AstName
	Annotation    AstType
	ColonPosition *lex.Position
}

type BindingList []Binding

// --------------------------------------------------------------------------------
// -- PARSER RESULT TYPES
// --------------------------------------------------------------------------------

type ParseError struct {
	Location lex.Location
	Message  string
}

type Options struct {
	CaptureComments bool
	StoreCstData    bool
}

type Attrs []AstAttr

type HotComment struct {
	Header   bool
	Content  string
	Location lex.Location
}

func (c HotComment) String() string {
	var b strings.Builder

	b.WriteString("HotComment\n")
	b.WriteString(fmt.Sprintf("Header: %t\n", c.Header))
	b.WriteString(fmt.Sprintf("Content: %q\n", c.Content))
	b.WriteString(fmt.Sprintf("Location: %s\n", c.Location))

	return b.String()
}

type FunctionState struct {
	Vararg    bool
	LoopDepth int
}

type Result struct {
	Root             AstStatBlock
	CommentLocations []Comment
	HotComments      []HotComment
	CstNodeMap       map[AstNode]CstNode
	Errors           []ParseError
}

func (r Result) String() string {
	var b strings.Builder

	b.WriteString("Root:\n")
	b.WriteString(indentStart(r.Root.String(), 2))

	if len(r.CommentLocations) > 0 {
		b.WriteString("CommentLocations:\n")
		for _, c := range r.CommentLocations {
			b.WriteString(indentStart(c.String(), 2))
		}
	}

	if len(r.HotComments) > 0 {
		b.WriteString("HotComments:\n")
		for _, c := range r.HotComments {
			b.WriteString(indentStart(c.String(), 2))
		}
	}

	if len(r.CstNodeMap) > 0 {
		b.WriteString("CstNodeMap:\n")
		for k, v := range r.CstNodeMap {
			b.WriteString(indentStart(fmt.Sprintf("%v: %v\n", k, v), 2))
		}
	}

	if len(r.Errors) > 0 {
		b.WriteString("Errors:\n")
		for _, e := range r.Errors {
			b.WriteString(indentStart(fmt.Sprintf("%s at %s\n", e.Message, e.Location), 2))
		}
	}

	return b.String()
}
