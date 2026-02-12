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
	// token_start_line = 0
	// token_start_col  = 0
	// token_end_line   = 0
	// token_end_col    = 0
	token_location lex.Location
)

// Previous Token Location (for errors/end mismatch)
var (
	// prev_start_line = 0
	// prev_start_col  = 0
	// prev_end_line   = 0
	// prev_end_col    = 0
	prev_location lex.Location
)

// Payload
var (
	token_string    *string = nil
	token_aux       *int    = nil
	token_codepoint *uint32 = nil
)

// Parser init

var recursionCounter = 0

var (
	commentLocations = []Comment{}
	hotcomments      = []HotComment{}
	parseErrors      = []ParseError{}
	cstNodes         = map[AstNode]CstNode{} // todo: change to pointer if needed
)

// All unlocalized Parser functions

// there would be 52 declarations here if Go supported forward declarations
// honestly they're still better than function hoisting

// Suspect State

var next_type = lex.Eof

// next_start_line = 0
// next_end_line   = 0

// next_start_col = 0
// next_end_col   = 0
var next_location lex.Location

var (
	next_codepoint *uint32 = nil
	next_string    *string = nil
	next_aux       *int    = nil
)

var suspect_type = lex.Eof

var suspect_line uint32

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

// Lexer

// Uh...
// That's in the other package

var lexer lex.Lexer

// Parser Interface

func fillNext() {
	for {
		next := lexer.Lookahead()

		if next.Type == lex.Comment || next.Type == lex.BlockComment || next.Type == lex.BrokenComment {
			if captureComments {
				commentLocations = append(commentLocations, Comment{
					Type:    next.Type,
					NodeLoc: NodeLoc{next.Location},
				})
			}

			if next.Type == lex.Comment && next_string != nil && (*next_string)[0] == '!' {
				hotcomments = append(hotcomments, HotComment{
					Header:   hotcommentHeader,
					Location: next.Location,
					Content:  *next_string,
				})
			}

			if next.Type == lex.BrokenComment {
				return
			}

			continue
		}

		break
	}
}

func nextLexeme() {
	// Save previous current to prev
	prev_location = token_location

	// Move NEXT to CURRENT
	token_type = next_type
	token_location = next_location
	token_string = next_string
	token_aux = next_aux
	token_codepoint = next_codepoint

	// Refill NEXT
	fillNext()
}

// Parser Commons

func snapshot() lex.Location {
	return token_location
}

func get_lexeme() lex.Lexeme {
	return lex.Lexeme{
		Type:     token_type,
		Location: token_location,
	}
}

func getprev() lex.Location {
	return prev_location
}

// Error reports

func report(loc lex.Location, msg string) {
	if len(parseErrors) > 0 && parseErrors[len(parseErrors)-1].Location == loc {
		return
	}

	parseErrors = append(parseErrors, ParseError{Location: loc, Message: msg})

	if ErrorLimit == 1 {
		panic(msg)
	}

	if len(parseErrors) >= ErrorLimit {
		panic(fmt.Sprintf("Reached error limit (%d)", ErrorLimit))
	}
}

func expectAndConsumeFail(type_ lex.LexemeType, context *string) {
	typeString := lex.Lexeme{Type: type_}.String()

	lexLex := lex.Lexeme{Codepoint: token_codepoint}
	if token_string != nil {
		lexLex.Data = []byte(*token_string)
	}
	lexString := lexLex.String()

	if context != nil {
		report(snapshot(), fmt.Sprintf("Expected %s when parsing %s, got %s", typeString, *context, lexString))
	} else {
		report(snapshot(), fmt.Sprintf("Expected %s, got %s", typeString, lexString))
	}
}

func expectMatchAndConsumeFail(type_, begin_type lex.LexemeType, position lex.Position, extra ...string) {
	typeString := lex.Lexeme{Type: type_}.String()
	matchString := lex.Lexeme{Type: begin_type}.String()
	currLex := lex.Lexeme{Type: token_type, Codepoint: token_codepoint}
	if token_string != nil {
		currLex.Data = []byte(*token_string)
	}
	currString := currLex.String()

	var xtra string
	if len(extra) > 0 {
		xtra = extra[0]
	}

	if token_location.Begin.Line == position.Line {
		report(snapshot(), fmt.Sprintf("Expected %s (to close %s at column) %d, got %s%s", typeString, matchString, position.Column+1, currString, xtra))
	} else {
		report(snapshot(), fmt.Sprintf("Expected %s (to close %s at line) %d, got %s%s", typeString, matchString, position.Line, currString, xtra))
	}
}

func expectAndConsume(type_ lex.LexemeType, context *string) bool {
	if token_type != type_ {
		expectAndConsumeFail(type_, context)

		if next_type == type_ {
			nextLexeme()
			nextLexeme()
		}

		return false
	}

	nextLexeme()
	return true
}

func expectMatchAndConsume(value, begin_type lex.LexemeType, position lex.Position, seachForMissing *bool) bool {
	if token_type != value {
		expectMatchAndConsumeFail(value, begin_type, position)

		if seachForMissing != nil && (*seachForMissing) {
			currentLine := prev_location.End.Line
			type_ := token_type

			for currentLine == token_location.Begin.Line && type_ != value && matchRecovery[type_] == 0 {
				nextLexeme()
				type_ = token_type
			}

			if type_ == value {
				nextLexeme()
				return true
			}
		} else {
			if next_type == value {
				nextLexeme()
				nextLexeme()
				return true
			}
		}

		return false
	}

	nextLexeme()
	return true
}

func expectMatchEndAndConsume(type_, begin_type lex.LexemeType, position lex.Position) bool {
	if token_type != type_ {
		if suspect_type != lex.Eof && suspect_line > position.Line {
			suggestionLex := lex.Lexeme{Type: suspect_type, Codepoint: next_codepoint}
			if token_string != nil {
				suggestionLex.Data = []byte(*token_string)
			}
			suggestionString := suggestionLex.String()

			suggestion := fmt.Sprintf("; did you forget to close %s at line %d?", suggestionString, position.Line+1)

			expectMatchAndConsumeFail(type_, begin_type, position, suggestion)
		} else {
			expectMatchAndConsumeFail(type_, begin_type, position)
		}

		if next_type == type_ {
			nextLexeme()
			nextLexeme()
			return true
		}

		return false
	}

	if token_location.Begin.Line != position.Line && token_location.Begin.Column != position.Column && suspect_line < position.Line {
		suspect_line = position.Line
		suspect_type = begin_type
	}

	nextLexeme()
	return true
}

// Ast reports

func reportStatError(location lex.Location, exprs []AstExpr, stats []AstStat, msg string) AstStatError {
	report(location, msg)

	return AstStatError{
		NodeLoc:      NodeLoc{location},
		Expressions:  exprs,
		Statements:   stats,
		MessageIndex: len(parseErrors) - 1,
	}
}

func reportExprError(location lex.Location, exprs []AstExpr, msg string) AstExprError {
	report(location, msg)

	return AstExprError{
		NodeLoc:      NodeLoc{location},
		Expressions:  exprs,
		MessageIndex: len(parseErrors) - 1,
	}
}

func reportTypeError(location lex.Location, types []AstType, msg string) AstTypeError {
	report(location, msg)

	return AstTypeError{
		NodeLoc:      NodeLoc{location},
		Types:        types,
		MessageIndex: len(parseErrors) - 1,
	}
}

func reportNameError(context *string) {
	currLex := lex.Lexeme{Type: token_type, Codepoint: token_codepoint}
	if token_string != nil {
		currLex.Data = []byte(*token_string)
	}
	currString := currLex.String()

	if context != nil {
		report(snapshot(), fmt.Sprintf("Expected identifier when parsing %s, got %s", *context, currString))
	} else {
		report(snapshot(), fmt.Sprintf("Expected identifier, got %s", currString))
	}
}

// Locals helpers

func restoreLocals(offset int) {
	for i := len(localStack) - 1; i >= offset; i-- {
		l := localStack[i]
		// l better not be nil bruh
		localMap[l.Name] = l.Shadow
	}

	localStack = localStack[:offset]
}

func pushLocal(binding Binding) *AstLocal {
	name := binding.Name.Value
	shadow := localMap[name]

	local := &AstLocal{
		Name:          name,
		NodeLoc:       binding.NodeLoc,
		Shadow:        shadow,
		FunctionDepth: len(functionStack) - 1,
		LoopDepth:     functionStack[len(functionStack)-1].LoopDepth,
		Annotation:    binding.Annotation,
	}

	localMap[name] = local
	localStack = append(localStack, local)

	return local
}

func incrementRecursionCounter(context string) {
	recursionCounter++

	if recursionCounter > RecursionLimit {
		msg := fmt.Sprintf("Exceeded allowed recursion depth; simplify your %s to make the code compile", context)
		report(snapshot(), msg) // lol y
		panic(msg)
	}
}

// The core of the code

func parseBinding() Binding {
	nameOpt := parseNameOpt("variable name")

	var bindingName Binding
	if nameOpt != nil {
		bindingName = *nameOpt
	} else {
		bindingName = Binding{
			Name:    lex.AstName{Value: nameError},
			NodeLoc: NodeLoc{snapshot()},
		}
	}

	colonPos := token_location.Begin
	annotation := parseOptionalType()

	return Binding{
		Name:          bindingName.Name,
		NodeLoc:       bindingName.NodeLoc,
		Annotation:    annotation,
		ColonPosition: &colonPos,
	}
}

// bindinglist ::= (binding | `...') [`,' bindinglist]
func parseBindingList(result *[]Binding, allowDot3 bool, commaPositions *[]lex.Position, initialComma *lex.Position, varargAnnotColonPos *[]*lex.Position) (bool, *lex.Location, AstTypePack) {
	localCommaPositions := []lex.Position{}

	if commaPositions != nil && initialComma != nil {
		localCommaPositions = append(localCommaPositions, *initialComma)
	}

	for {
		if token_type == lex.Dot3 && allowDot3 {
			varargLocation := snapshot()
			nextLexeme()

			var tailAnnotation AstTypePack

			if token_type == ':' {
				if varargAnnotColonPos != nil {
					(*varargAnnotColonPos)[0] = &token_location.Begin
				}

				nextLexeme()
				tailAnnotation = parseVariadicArgumentTypePack()
			}

			if commaPositions != nil {
				for _, v := range localCommaPositions {
					*commaPositions = append(*commaPositions, v)
				}
			}

			return true, &varargLocation, tailAnnotation
		}

		*result = append(*result, parseBinding())

		if token_type != ',' {
			break
		}

		if commaPositions != nil {
			localCommaPositions = append(localCommaPositions, token_location.Begin)
		}

		nextLexeme()
	}

	if commaPositions != nil {
		for _, v := range localCommaPositions {
			*commaPositions = append(*commaPositions, v)
		}
	}

	return false, nil, nil
}

// stat ::=
// varlist `=' explist |
// functioncall |
// do block end |
// while exp do block end |
// repeat block until exp |
// if exp then block {elseif exp then block} [else block] end |
// for binding `=' exp `,' exp [`,' exp] do block end |
// for namelist in explist do block end |
// function funcname funcbody |
// attributes function funcname funcbody |
// local function Name funcbody |
// local attributes function Name funcbody |
// local namelist [`=' explist]
// laststat ::= return [explist] | break
func parseStat() AstStat {
	type_ := token_type

	switch type_ {
	case lex.ReservedIf:
		return parseIf()
	case lex.ReservedWhile:
		return parseWhile()
	case lex.ReservedDo:
		return parseDo()
	case lex.ReservedFor:
		return parseFor()
	case lex.ReservedRepeat:
		return parseRepeat()
	case lex.ReservedFunction:
		return parseFunctionStat(nil)
	case lex.ReservedLocal:
		return parseLocal(nil)
	case lex.ReservedReturn:
		return parseReturn()
	case lex.ReservedBreak:
		return parseBreak()
	case lex.Attribute, lex.AttributeOpen:
		return parseAttributeStat()
	}
}

func parseBlock() AstStatBlock
func parseIf() AstStatIf
func parseWhile() AstStatWhile
func parseRepeat() AstStatRepeat
func parseDo() AstStatBlock
func parseBreak() AstStatBreakOrError
func parseContinue() AstStatContinueOrError
func parseFor() AstStatForOrForIn
func parseFunctionStat(attributes Attrs) AstStatFunction
func parseAttribute(attributes Attrs)
func parseAttributeStat() AstStat
func parseLocal(attributes Attrs) AstStatLocal
func parseReturn() AstStatReturn
func parseTypeAlias(start lex.Location, exported bool, typeKeywordPosition lex.Position) AstStatTypeAliasOrTypeFunction

var typeFunctionDepth = 0

func parseTypeFunction(start lex.Location, exported bool, typeKeywordPosition lex.Position) AstStatTypeFunction
func parseNameOpt(context *string) *Binding
func parseName(context *string) Binding
func tableSeparator() *int
func parseListExpr(result []AstExpr, commaPositions *[]lex.Position)

// varlist `=' explist
func parseAssignment(initial AstExpr) AstStatAssign

// var [`+=' | `-=' | `*=' | `/=' | `%=' | `^=' | `..='] exp
func parseCompoundAssignment(initial AstExpr, op int) AstStatCompoundAssign
func prepareFunctionArguments(start lex.Location, hasself bool, args []Binding)

func shouldParseTypePack() bool {
	t := token_type

	if t == lex.Dot3 {
		return true
	}

	if t == lex.Name && next_type == lex.Dot3 {
		return true
	}

	return false
}

// funcbody ::= `(' [parlist] `)' [`:' ReturnType] block end
// parlist ::= bindinglist [`,' `...'] | `...'
func parseFunctionBody(hasself bool, matchFunction lex.Lexeme, debugname, localName *string, attributes Attrs) (AstExprFunction, *AstLocal)
func parseGenericTypeList(withDefaultValues bool, openPosRef, commaPosRef, closePosRef *[]lex.Position) ([]AstGenericType, []AstGenericTypePack)
func parseOptionalType() AstType

// TypeList ::= Type [`,' TypeList] | ...Type
func parseTypeList(result []AstType, resultNames []*AstArgumentName, commaPositions *[]lex.Position, nameColonPositions *[]*lex.Position) AstTypePack

func parseVariadicArgumentTypePack() AstTypePackVariadicOrGeneric
