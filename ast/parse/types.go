package main

import "github.com/Heliodex/coputer/ast/lex"

// bindings

type Binding struct {
	NodeLoc
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
