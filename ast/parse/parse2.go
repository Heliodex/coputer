package main

import (
	"github.com/Heliodex/coputer/ast/lex"
)

// parse 2 go!

// parse resets the parser state and parses the whole source, storing the root
// block in p.parseRoot.
func (p *Parser) parse() {
	p.token_type = lex.Eof
	p.token_location = lex.Location{}
	p.prev_location = lex.Location{}
	p.token_string = nil
	p.token_aux = nil
	p.token_codepoint = nil

	p.recursionCounter = 0

	p.commentLocations = nil
	p.pendingComments = nil
	p.hotcomments = nil
	p.parseErrors = nil
	p.cstNodes = map[AstNode]CstNode{}

	p.declaredExportBindings = map[string]lex.Location{}
	p.hasModuleReturn = false

	p.hotcommentHeader = true

	p.suspect_type = lex.Eof
	p.suspect_line = 0

	p.matchRecovery = [lex.Reserved_END]int{}
	p.matchRecovery[lex.Eof] = 1

	p.functionStack = []FunctionState{
		{Vararg: true, LoopDepth: 0},
	}

	p.localStack = nil
	p.localMap = map[string]*AstLocal{}
	p.braceStack = nil

	p.fillNext()
	p.nextLexeme()
	p.hotcommentHeader = false

	localsBegin := len(p.localStack)
	result := p.parseBlockNoScope()
	p.restoreLocals(localsBegin)

	if p.token_type != lex.Eof {
		p.expectAndConsumeFail(lex.Eof, nil)
	}

	p.parseRoot = result
}

// Parse is the exported entry point
func Parse(src string, opts Options) (bool, Result) {
	p := newParser(src, opts)

	func() {
		defer func() {
			if r := recover(); r != nil {
				// on panic, return what we have
			}
		}()
		p.parse()
	}()

	var rootBlock AstStatBlock
	if root := p.parseRoot; root != nil {
		rootBlock = *root
	}

	// comments are always collected so that Source() can reproduce them, but
	// they're only reported in the result when requested
	var commentLocations []Comment
	if p.captureComments {
		commentLocations = p.commentLocations
	}

	return len(p.parseErrors) == 0,
		Result{
			Root:             rootBlock,
			CommentLocations: commentLocations,
			HotComments:      p.hotcomments,
			CstNodeMap:       p.cstNodes,
			Errors:           p.parseErrors,
		}
}
