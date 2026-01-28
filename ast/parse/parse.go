package main

import "github.com/Heliodex/coputer/ast/lex"

// Parser Settings

var LuauExplicitTypeInstantiationSyntax = false

var (
	TypeLengthLimit = 1000
	RecursionLimit  = 1000
	ErrorLimit      = 100
)

var hotcommentHeader = true

type QuoteStyle uint8

const (
	QuoteStyle_QuotedSimple QuoteStyle = iota
	QuoteStyle_QuotedSingle
	QuoteStyle_QuotedRaw
	QuoteStyle_Unquoted
)

type CstQuotes uint8

const (
	CstQuotes_Single CstQuotes = iota
	CstQuotes_Double
	CstQuotes_Raw
	CstQuotes_Interp
)

type UnaryOp uint8

const (
	UnaryOp_Not UnaryOp = iota
	UnaryOp_Minus
	UnaryOp_Len
)

type BinaryOp uint8

const (
	BinaryOp_Add BinaryOp = iota
	BinaryOp_Sub
	BinaryOp_Mul
	BinaryOp_Div
	BinaryOp_FloorDiv
	BinaryOp_Mod
	BinaryOp_Pow
	BinaryOp_Concat
	BinaryOp_CompareNe
	BinaryOp_CompareEq
	BinaryOp_CompareLt
	BinaryOp_CompareLe
	BinaryOp_CompareGt
	BinaryOp_CompareGe
	BinaryOp_And
	BinaryOp_Or
)

var BinaryPriority = map[BinaryOp][2]int{
	BinaryOp_Add:       {6, 6},
	BinaryOp_Sub:       {6, 6},
	BinaryOp_Mul:       {7, 7},
	BinaryOp_Div:       {7, 7},
	BinaryOp_FloorDiv:  {7, 7},
	BinaryOp_Mod:       {7, 7},
	BinaryOp_Pow:       {10, 9},
	BinaryOp_Concat:    {5, 4},
	BinaryOp_CompareNe: {3, 3},
	BinaryOp_CompareEq: {3, 3},
	BinaryOp_CompareLt: {3, 3},
	BinaryOp_CompareLe: {3, 3},
	BinaryOp_CompareGt: {3, 3},
	BinaryOp_CompareGe: {3, 3},
	BinaryOp_And:       {2, 2},
	BinaryOp_Or:        {1, 1},
}

var CompoundLookup = map[lex.LexemeType]BinaryOp{
	lex.FloorDivAssign: BinaryOp_FloorDiv,
	lex.ConcatAssign:   BinaryOp_Concat,
	lex.ModAssign:      BinaryOp_Mod,
	lex.PowAssign:      BinaryOp_Pow,
	lex.AddAssign:      BinaryOp_Add,
	lex.SubAssign:      BinaryOp_Sub,
	lex.MulAssign:      BinaryOp_Mul,
	lex.DivAssign:      BinaryOp_Div,
}

var BinaryOpLookup = map[lex.LexemeType]BinaryOp{
	43: BinaryOp_Add,
	45: BinaryOp_Sub,
	42: BinaryOp_Mul,
	47: BinaryOp_Div,

	lex.FloorDiv: BinaryOp_FloorDiv,

	37: BinaryOp_Mod,
	94: BinaryOp_Pow,

	lex.Dot2:     BinaryOp_Concat,
	lex.NotEqual: BinaryOp_CompareNe,
	lex.Equal:    BinaryOp_CompareEq,

	60: BinaryOp_CompareLt,

	lex.LessEqual: BinaryOp_CompareLe,

	62: BinaryOp_CompareGt,

	lex.GreaterEqual: BinaryOp_CompareGe,
	lex.ReservedAnd:  BinaryOp_And,
	lex.ReservedOr:   BinaryOp_Or,
}

var UnaryOpLookup = map[lex.LexemeType]UnaryOp{
	lex.ReservedNot: UnaryOp_Not,
	45:              UnaryOp_Minus,
	35:              UnaryOp_Len,
}

var BlockFollow = map[lex.LexemeType]bool{
	lex.ReservedElseif: true,
	lex.ReservedUntil:  true,
	lex.ReservedElse:   true,
	lex.ReservedEnd:    true,
	lex.Eof:            true,
}

// var ConstantLiteral = map[string]bool{
// 	"ExprConstantNil":    true,
// 	"ExprConstantBool":   true,
// 	"ExprConstantNumber": true,
// 	"ExprConstantString": true,
// }

func ConstantLiteral(expr AstExpr) bool {
	switch expr.(type) {
	case AstExprConstantNil, AstExprConstantBool, AstExprConstantNumber, AstExprConstantString:
		return true
	}
	return false
}

var ExprLValues = map[string]bool{
	"Exprvar":       true,
	"ExprGlobal":    true,
	"ExprIndexExpr": true,
	"ExprIndexName": true,
}

// Lookups for Lexer

var (
	HexDigits = map[int]bool{}
	HexVal    = map[int]int{}
	Digits    = map[int]bool{}
	Alpha     = map[int]bool{}
)

// var Spaces = map[int]bool{
// 	9:  true, // \t
// 	10: true, // \n
// 	11: true, // \v
// 	12: true, // \f
// 	13: true, // \r
// 	32: true, // space
// }

// var Escapes = map[int]int{
// 	97:  7,
// 	98:  8,
// 	102: 12,
// 	110: 10,
// 	114: 13,
// 	116: 9,
// 	118: 11,
// }

func init() {
	for i := 48; i <= 57; i++ {
		HexDigits[i] = true
		Digits[i] = true
	}

	for i := 65; i <= 90; i++ {
		if i <= 70 {
			HexDigits[i] = true
		}
		Alpha[i] = true
	}

	for i := 97; i <= 122; i++ {
		if i <= 102 {
			HexDigits[i] = true
		}
		Alpha[i] = true
	}

	for i := 48; i <= 57; i++ {
		HexVal[i] = i - 48
	}
	for i := 65; i <= 70; i++ {
		HexVal[i] = i - 55
	}
	for i := 97; i <= 102; i++ {
		HexVal[i] = i - 87
	}
}

// Parser Constants

const (
	nameError  = "%error-id%"
	nameNumber = "number"
	nameSelf   = "self"
	nameNil    = "nil"
)

// Lexer helpers

// nothing here

// Parser helpers

func isLiteralTable(aexpr AstExpr) bool {
	// todo: check & change this to a pointer if it fux up
	expr, ok := aexpr.(AstExprTable)
	if !ok {
		return false
	}

	for _, item := range expr.Items {
		if item.Kind == "General" {
			return false
		}
		if item.Kind == "Record" || item.Kind == "List" {
			if !ConstantLiteral(item.Value) && !isLiteralTable(item.Value) {
				return false
			}
		}
	}

	return true
}
