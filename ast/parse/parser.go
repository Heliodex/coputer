package main

import (
	"github.com/Heliodex/coputer/ast/lex"
)

// Parser holds all mutable state for a single parse. Keeping the state on a struct (instead of package-level globals) means each Parse call gets an isolated parser, and parsing no longer relies on shared mutable globals.
type Parser struct {
	captureComments bool
	storeCstData    bool

	lexer lex.Lexer

	// Current token
	token_type      lex.LexemeType
	token_location  lex.Location
	token_string    *string
	token_aux       *int
	token_codepoint *uint32

	// Previous token location (for errors/end mismatch)
	prev_location lex.Location

	// Lookahead token
	next_type      lex.LexemeType
	next_location  lex.Location
	next_string    *string
	next_codepoint *uint32
	next_aux       *int

	// Confusable tracking
	suspect_type lex.LexemeType
	suspect_line uint32

	matchRecovery [lex.Reserved_END]int

	functionStack []FunctionState
	localStack    []*AstLocal
	localMap      map[string]*AstLocal

	commentLocations []Comment
	// pendingComments are comments not yet attached to the block they appear in
	pendingComments []Comment
	hotcomments     []HotComment
	parseErrors     []ParseError
	cstNodes        map[AstNode]CstNode

	recursionCounter int

	hotcommentHeader bool

	// Export value syntax state (top-level `export local/function/const`)
	declaredExportBindings map[string]lex.Location
	hasModuleReturn        bool

	typeFunctionDepth int

	// Interpolated string parsing state
	braceStack []lex.BraceType

	parseRoot *AstStatBlock
}

// newParser creates a Parser for src with the given options.
func newParser(src string, opts Options) *Parser {
	p := &Parser{
		captureComments: opts.CaptureComments,
		storeCstData:    opts.StoreCstData,
		lexer:           lex.NewLexer(src),
	}
	p.matchRecovery[lex.Eof] = 1
	return p
}
