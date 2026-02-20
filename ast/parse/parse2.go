package main

import (
	"github.com/Heliodex/coputer/ast/lex"
)

// parse 2 go!

// braceStack for interpolated string parsing
var braceStack []lex.BraceType

var parseRoot *AstStatBlock

func parseInternal(src string, opts Options) {
	captureComments = opts.CaptureComments
	storeCstData = opts.StoreCstData

	lexer = lex.NewLexer(src)

	token_type = lex.Eof
	token_location = lex.Location{}
	prev_location = lex.Location{}
	token_string = nil
	token_aux = nil
	token_codepoint = nil

	recursionCounter = 0

	commentLocations = nil
	hotcomments = nil
	parseErrors = nil
	cstNodes = map[AstNode]CstNode{}

	hotcommentHeader = true

	suspect_type = lex.Eof
	suspect_line = 0

	matchRecovery = [lex.Reserved_END]int{}
	matchRecovery[lex.Eof] = 1

	functionStack = []FunctionState{
		{Vararg: true, LoopDepth: 0},
	}

	localStack = nil
	localMap = map[string]*AstLocal{}
	braceStack = nil

	fillNext()
	nextLexeme()
	hotcommentHeader = false

	localsBegin := len(localStack)
	result := parseBlockNoScope()
	restoreLocals(localsBegin)

	if token_type != lex.Eof {
		expectAndConsumeFail(lex.Eof, nil)
	}

	parseRoot = result
}

// Parse is the exported entry point
func Parse(src string, opts Options) (bool, Result) {
	func() {
		defer func() {
			if r := recover(); r != nil {
				// on panic, return what we have
			}
		}()
		parseInternal(src, opts)
	}()

	var rootBlock AstStatBlock
	if root := parseRoot; root != nil {
		rootBlock = *root
	}

	return len(parseErrors) == 0,
		Result{
			Root:             rootBlock,
			CommentLocations: commentLocations,
			HotComments:      hotcomments,
			CstNodeMap:       cstNodes,
			Errors:           parseErrors,
		}
}
