package main

import . "github.com/Heliodex/coputer/ast/lex"

type Parser struct{}

func shouldParseTypePack(lexer *Lexer) bool {
	if lexer.Current().Type == Dot3 {
		return true
	} else if lexer.Current().Type == Name && lexer.Lookahead().Type == Dot3 {
		return true
	}

	return false
}
