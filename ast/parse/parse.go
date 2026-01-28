package main

import (
	"fmt"

	"github.com/Heliodex/coputer/ast/lex"
)

// These globals aren't that great to have around for now, though they'll stick here until compliance with the reference implementation is ensured.

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

// Attributes

func deprecatedArgsValidator(attrLoc lex.Location, args []AstExpr) (errors []ParseError) {
	if len(args) == 0 {
		return errors
	}
	if len(args) > 1 {
		errors = append(errors, ParseError{
			Location: attrLoc,
			Message:  "@deprecated can be parametrized only by 1 argument",
		})
		return errors
	}

	aarg := args[0]
	arg, ok := aarg.(AstExprTable)
	if !ok {
		errors = append(errors, ParseError{
			Location: attrLoc,
			Message:  "Unknown argument type for @deprecated",
		})
		return errors
	}

	for _, item := range arg.Items {
		if item.Key != nil && item.Kind == "Record" {
			if itemKey, ok := (*item.Key).(AstExprConstantString); ok {
				keyString := itemKey.Value
				if _, ok := item.Value.(AstExprConstantString); ok {
					if keyString != "use" && keyString != "reason" {
						errors = append(errors,
							ParseError{
								Location: item.Value.GetLocation(),
								Message:  fmt.Sprintf("Unknown key '%s' for @deprecated. Only string constants for 'use' and 'reason' are allowed", keyString),
							},
						)
					}
				} else {
					errors = append(errors,
						ParseError{
							Location: item.Value.GetLocation(),
							Message:  fmt.Sprintf("Only constant string allowed as value for '%s'", keyString),
						},
					)
				}
				continue
			}
		}
		errors = append(errors,
			ParseError{
				Location: item.Value.GetLocation(),
				Message:  "Only constants keys 'use' and 'reason' are allowed for @deprecated attribute",
			},
		)
	}
	return
}

type AttributeEntry struct {
	Type          string
	ArgsValidator func(lex.Location, []AstExpr) []ParseError
}

var kAttributeEntries = map[string]AttributeEntry{
	"checked": {
		Type: "Checked",
	},
	"native": {
		Type: "Native",
	},
	"deprecated": {
		Type:          "Deprecated",
		ArgsValidator: deprecatedArgsValidator,
	},
}

// Main

var (
	options Options
	source  string
)

// Settings init

var (
	captureComments = options.CaptureComments
	storeCstData    = options.StoreCstData
)

// Lexer State & Buffer

// var (
// 	buff = []byte(source)
// 	size = len(buff)
// )

// Current State

// var (
// 	offset = 0
// 	line = 0
// 	lineOffset = 0
// )

// Current Token State

var token_type = lex.Eof

// Locations
var (
	token_start_line = 0
	token_start_col  = 0
	token_end_line   = 0
	token_end_col    = 0
)

// Previous Token Location (for errors/end mismatch)
var (
	prev_start_line = 0
	prev_start_col  = 0
	prev_end_line   = 0
	prev_end_col    = 0
)

// Payload
var (
	token_string    *string = nil
	token_aux       *int    = nil
	token_codepoint *int    = nil
)

// Parser init

var recursionCounter = 0

var (
	comments    = []Comment{}
	hotcomments = []HotComment{}
	parseErrors = []ParseError{}
	cstNodes    = map[AstNode]CstNode{} // todo: change to pointer if needed
)

// All unlocalized Parser functions

// there would be 52 declarations here if Go supported forward declarations
// honestly they're still better than function hoisting

// Suspect State

var (
	next_type       = lex.Eof
	next_start_line = 0
	next_end_line   = 0
)

var (
	next_start_col = 0
	next_end_col   = 0
)

var (
	next_codepoint *int    = nil
	next_string    *string = nil
	next_aux       *int    = nil
)

var suspect_type = lex.Eof

var suspect_line = 0

var matchRecovery = [lex.Reserved_END]int{}

func init() {
	matchRecovery[lex.Eof] = 1
}

// Stacks

var functionStack = []FunctionState{
	{Vararg: true, LoopDepth: 0},
}

var (
	localStack = []*AstLocal{}
	localMap   = map[string]*AstLocal{}
)
