package main

import (
	"github.com/Heliodex/coputer/ast/ast"
	"github.com/Heliodex/coputer/ast/lex"
)

type Function struct {
	vararg    bool
	loopDepth uint32
}

type AstNode struct {
	classIndex int
	location   lex.Location
}

type AstType AstNode

type Local struct {
	local  *ast.Local[ast.Node]
	offset uint32
}

type Name struct {
	name     lex.AstName
	location lex.Location
}

type Binding struct {
	name          Name
	annotation    *AstType
	colonPosition lex.Position
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
