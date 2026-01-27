package main

import (
	"github.com/Heliodex/coputer/ast/ast"
	"github.com/Heliodex/coputer/ast/lex"
)

type Function struct {
	vararg    bool
	loopDepth uint32
}

type Name struct {
	name     lex.AstName
	location lex.Location
}

type Parser struct {
	lexer lex.Lexer
}

func shouldParseTypePack(lexer *lex.Lexer) bool {
	if lexer.Current().Type == lex.Dot3 {
		return true
	} else if lexer.Current().Type == lex.Name && lexer.Lookahead().Type == lex.Dot3 {
		return true
	}

	return false
}

func (p *Parser) blockFollow(l lex.Lexeme) bool {
	return l.Type == lex.Eof || l.Type == lex.ReservedElse || l.Type == lex.ReservedElseif || l.Type == lex.ReservedEnd || l.Type == lex.ReservedUntil
}

func (p *Parser) parseChunk() ast.AstStatBlock[ast.ASTNode] {
	result := p.parseBlock()

	if p.lexer.Current().Type != lex.Eof {
		p.expectAndConsumeFail(lex.Eof, nil)
	}

	return result
}

// chunk ::= {stat [`;']} [laststat [`;']]
// block ::= chunk
func (p *Parser) parseBlock() ast.AstStatBlock[ast.ASTNode] {
	localsBegin := p.saveLocals()

	result := p.parseBlockNoScope()

	p.restoreLocals(localsBegin)

	return result
}

// func isStatLast() bool {}
