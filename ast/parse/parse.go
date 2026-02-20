package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Heliodex/coputer/ast/lex"
)

// These globals aren't that great to have around for now, though they'll stick here until compliance with the reference implementation is ensured.

// Parser Settings

const (
	LuauExplicitTypeInstantiationSyntax = false
	DesugaredArrayTypeReferenceIsEmpty  = false
	// LuauCstStatDoWithStatsStart         = true
)

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

func ExprLValues(expr AstExpr) bool {
	switch expr.(type) {
	case AstExprLocal, AstExprGlobal, AstExprIndexExpr, AstExprIndexName:
		return true
	}
	return false
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

type ArgsValidator func(lex.Location, []AstExpr) []ParseError

type AttributeEntry struct {
	Type          string
	ArgsValidator ArgsValidator
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
		next := lexer.Next0()

		fmt.Println("lexed next type", lex.Lexeme{Type: next.Type}.String())

		next_type = next.Type
		next_location = next.Location
		next_codepoint = next.Codepoint
		nstr := string(next.Data)
		next_string = &nstr
		next_aux = next.Aux

		if next.Type == lex.Comment || next.Type == lex.BlockComment || next.Type == lex.BrokenComment {
			if captureComments {
				commentLocations = append(commentLocations, Comment{
					Type:    next.Type,
					NodeLoc: &NodeLoc{next.Location},
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

	fmt.Println("filled next with type", lex.Lexeme{Type: next_type}.String())
}

func nextLexeme() {
	// Save previous current to prev
	prev_location = token_location

	// Move NEXT to CURRENT
	token_type = next_type
	fmt.Println("set token_type to", lex.Lexeme{Type: token_type}.String())
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

func reportStatError(location lex.Location, exprs []AstExpr, stats []AstStat, msg string) *AstStatError {
	report(location, msg)

	return &AstStatError{
		NodeLoc:      &NodeLoc{location},
		Expressions:  exprs,
		Statements:   stats,
		MessageIndex: len(parseErrors) - 1,
	}
}

func reportExprError(location lex.Location, exprs []AstExpr, msg string) *AstExprError {
	report(location, msg)

	return &AstExprError{
		NodeLoc:      &NodeLoc{location},
		Expressions:  exprs,
		MessageIndex: len(parseErrors) - 1,
	}
}

func reportTypeError(location lex.Location, types []AstType, msg string) *AstTypeError {
	report(location, msg)

	return &AstTypeError{
		NodeLoc:      &NodeLoc{location},
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

	// setting to nil wouldn't change the array length, which I assume we're relying on somewhere...
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
	context := "variable name"
	nameOpt := parseNameOpt(&context)

	var bindingName Binding
	if nameOpt != nil {
		bindingName = *nameOpt
	} else {
		bindingName = Binding{
			Name:    lex.AstName{Value: nameError},
			NodeLoc: &NodeLoc{snapshot()},
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

	start_line := token_location.Begin.Line
	start_column := token_location.Begin.Column
	expr := parsePrimaryExpr(true)

	if e, ok := expr.(AstExprCall); ok {
		return &AstStatExpr{
			NodeLoc: e.NodeLoc,
			Expr:    e,
		}
	}

	if token_type == ',' || token_type == '=' {
		return parseAssignment(expr)
	}

	operator, ok := CompoundLookup[token_type]
	if ok {
		return parseCompoundAssignment(expr, operator)
	}

	var ident *string
	if e, ok := expr.(AstExprGlobal); ok {
		ident = &e.Name
	} else if e, ok := expr.(AstExprLocal); ok {
		ident = &e.Local.Name
	}

	if ident != nil && *ident == "type" {
		loc := expr.GetLocation()
		return parseTypeAlias(loc, false, loc.Begin)
	}

	if ident != nil && *ident == "export" && token_type == lex.Name && token_string != nil && *token_string == "type" {
		typeKeywordPos := token_location.Begin
		nextLexeme()
		return parseTypeAlias(expr.GetLocation(), true, typeKeywordPos)
	}

	if ident != nil && *ident == "continue" {
		return parseContinue(expr.GetLocation())
	}

	if start_line == token_location.Begin.Line && start_column == token_location.Begin.Column {
		nextLexeme()
	}

	return reportStatError(expr.GetLocation(), []AstExpr{expr}, nil, "Incomplete statement: expected assignment or a function call")
}

func parseBlockNoScope() *AstStatBlock {
	var body []AstStat

	prevPos := prev_location.End

	fmt.Println("Current token type at start of block:", token_type)
	for !BlockFollow[token_type] {
		oldRecursion := recursionCounter
		recursionCounter++

		stat := parseStat()

		recursionCounter = oldRecursion

		if token_type == ';' {
			nextLexeme()
			stat.SetHasSemicolon()

			loc := stat.GetLocation()
			// the fact that a table assignment isn't used here in the Luau implementation makes me suspicious that it's intended to be modified later on after returning
			stat.SetLocation(lex.Location{
				Begin: loc.Begin,
				End:   prevPos,
			})
		}

		body = append(body, stat)

		switch stat.(type) {
		case *AstStatBreak, *AstStatContinue, *AstStatReturn:
		default:
			continue
		}

		break // cuz I don't wanna label the loop
	}

	fmt.Println("Parsed block with body:", body)

	return &AstStatBlock{
		NodeLoc: &NodeLoc{
			lex.Location{
				Begin: prevPos,
				End:   token_location.Begin,
			},
		},
		Body:   body,
		HasEnd: false,
	}
}

// chunk ::= {stat [`;']} [laststat [`;']]
// block ::= chunk
func parseBlock() *AstStatBlock {
	localsBegin := len(localStack)
	result := parseBlockNoScope()
	restoreLocals(localsBegin)
	return result
}

// if exp then block {elseif exp then block} [else block] end
func parseIf() *AstStatIf {
	start := snapshot()

	nextLexeme()

	cond := parseExpr(0)

	// Then_location := token_location

	// okay what the package main import ( "fmt" "net/http" "time" ) func greet(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "Hello World! %s", time.Now()) } func main() { http.HandleFunc("/", greet) http.ListenAndServe(":8080", nil) }
	Then_begin := token_location.Begin
	Then_end := token_location.End

	var thenLocation *lex.Location
	if expectAndConsume(lex.ReservedThen, nil) {
		// do we intend to copy it here or smth or what??
		thenLocation = &lex.Location{
			Begin: Then_begin,
			End:   Then_end,
		}
	}

	thenBody := parseBlock()

	var elsebody AstStat
	end := start
	var elseLocation *lex.Location

	if token_type == lex.ReservedElseif {
		thenBody.HasEnd = true
		oldRecursionCount := recursionCounter
		recursionCounter++

		el := snapshot()
		elseLocation = &el
		elsebody = parseIf()
		end = elsebody.GetLocation()

		recursionCounter = oldRecursionCount
	} else {
		ThenElse_type := token_type

		ThenElse_begin := token_location.Begin
		ThenElse_end := token_location.End

		if token_type == lex.ReservedElse {
			thenBody.HasEnd = true
			el := snapshot()
			elseLocation = &el

			ThenElse_type = token_type

			ThenElse_begin = token_location.Begin
			ThenElse_end = token_location.End

			nextLexeme()

			body := parseBlock()
			body.Location.Begin = ThenElse_end
			elsebody = body
		}

		end = snapshot()

		hasEnd := expectMatchEndAndConsume(lex.ReservedEnd, ThenElse_type, ThenElse_begin)

		if elsebody != nil {
			if eb, ok := elsebody.(*AstStatBlock); ok {
				eb.HasEnd = hasEnd
			}
		} else {
			thenBody.HasEnd = hasEnd
		}
	}

	return &AstStatIf{
		NodeLoc:      &NodeLoc{lex.Location{Begin: start.Begin, End: end.End}},
		Condition:    cond, // sorry, it's my cawndishawn
		ThenBody:     *thenBody,
		ElseBody:     elsebody,
		ThenLocation: thenLocation,
		ElseLocation: elseLocation,
	}
}

// while exp do block end
func parseWhile() *AstStatWhile {
	start := snapshot()
	nextLexeme()

	cond := parseExpr(0)

	Do_type := token_type
	Do_begin := token_location.Begin
	Do_end := token_location.End

	context := "while loop"
	hasDo := expectAndConsume(lex.ReservedDo, &context)

	functionStack[len(functionStack)-1].LoopDepth++
	body := parseBlock()
	functionStack[len(functionStack)-1].LoopDepth--

	end := snapshot()
	hasEnd := expectMatchEndAndConsume(lex.ReservedEnd, Do_type, Do_begin)

	body.HasEnd = hasEnd

	return &AstStatWhile{
		NodeLoc:   &NodeLoc{lex.Location{Begin: start.Begin, End: end.End}},
		Condition: cond,
		Body:      body,
		HasDo:     hasDo,
		DoLocation: lex.Location{
			Begin: Do_begin,
			End:   Do_end,
		},
	}
}

// repeat block until exp
func parseRepeat() *AstStatRepeat {
	start := snapshot()

	Repeat_type := token_type
	Repeat_begin := token_location.Begin

	nextLexeme() // repeat

	localsBegin := len(localStack)

	functionStack[len(functionStack)-1].LoopDepth++
	body := parseBlock()
	functionStack[len(functionStack)-1].LoopDepth--

	untilPosition := token_location.Begin
	hasUntil := expectMatchAndConsume(lex.ReservedUntil, Repeat_type, Repeat_begin, nil)

	cond := parseExpr(0)

	restoreLocals(localsBegin)

	node := &AstStatRepeat{
		NodeLoc:   &NodeLoc{lex.Location{Begin: start.Begin, End: cond.GetLocation().End}},
		Condition: cond,
		Body:      body,
		HasUntil:  hasUntil,
	}

	if storeCstData {
		cstNodes[node] = CstStatRepeat{
			UntilPosition: untilPosition,
		}
	}

	return node
}

// do block end
func parseDo() *AstStatBlock {
	start := snapshot()

	Do_type := token_type
	Do_begin := token_location.Begin

	nextLexeme() // do

	body := parseBlock()
	body.Location.Begin = start.Begin

	endLocation := snapshot()
	body.HasEnd = expectMatchEndAndConsume(lex.ReservedEnd, Do_type, Do_begin)

	if body.HasEnd {
		body.Location.End = endLocation.End
	}

	if storeCstData {
		cstNodes[body] = CstStatDo{
			EndPosition: endLocation.Begin,
		}
	}

	return body
}

// break
func parseBreak() AstStatBreakOrError {
	start := snapshot()
	nextLexeme()

	if functionStack[len(functionStack)-1].LoopDepth == 0 {
		return reportStatError(start, nil, []AstStat{
			&AstStatContinue{NodeLoc: &NodeLoc{start}},
		}, "break statement must be inside a loop")
	}

	return &AstStatBreak{NodeLoc: &NodeLoc{start}}
}

// continue
func parseContinue(start lex.Location) AstStatContinueOrError {
	if functionStack[len(functionStack)-1].LoopDepth == 0 {
		return reportStatError(start, nil, []AstStat{
			&AstStatBreak{NodeLoc: &NodeLoc{start}},
		}, "continue statement must be inside a loop")
	}

	// note: the token is already parsed for us!

	return &AstStatContinue{NodeLoc: &NodeLoc{start}}
}

func extractAnnotationColonPositions(bindings []Binding) []*lex.Position {
	positions := make([]*lex.Position, len(bindings))
	for i, binding := range bindings {
		positions[i] = binding.ColonPosition
	}
	return positions
}

// for binding `=' exp `,' exp [`,' exp] do block end |
// for bindinglist in explist do block end |
func parseFor() AstStatForOrForIn {
	start := snapshot()
	nextLexeme() // for

	varname := parseBinding()

	if token_type == '=' { // === lel
		equalsPosition := token_location.Begin
		nextLexeme()

		from := parseExpr(0)

		endCommaPosition := token_location.Begin
		context := "index range"
		expectAndConsume(',', &context)

		to := parseExpr(0)

		var stepCommaPosition *lex.Position
		var step AstExpr

		if token_type == ',' {
			stepCommaPosition = &token_location.Begin
			nextLexeme()
			step = parseExpr(0)
		}

		Do_type := token_type
		Do_begin := token_location.Begin
		Do_end := token_location.End

		context2 := "for loop"
		hasDo := expectAndConsume(lex.ReservedDo, &context2)

		localsBegin := len(localStack)
		functionStack[len(functionStack)-1].LoopDepth++

		var_ := pushLocal(varname) // and here I was laughing at the fact I could call variables 'end'...
		body := parseBlock()

		functionStack[len(functionStack)-1].LoopDepth--
		restoreLocals(localsBegin)

		end := token_location.End
		hasEnd := expectMatchEndAndConsume(lex.ReservedEnd, Do_type, Do_begin)
		body.HasEnd = hasEnd

		node := &AstStatFor{
			NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: end}},
			Var:     var_,
			From:    from,
			To:      to,
			Step:    step,
			Body:    body,
			HasDo:   hasDo,
			DoLocation: lex.Location{
				Begin: Do_begin,
				End:   Do_end,
			},
		}

		if storeCstData {
			cstNodes[node] = CstStatFor{
				AnnotationColonPosition: varname.ColonPosition,
				EqualsPosition:          equalsPosition,
				EndCommaPosition:        endCommaPosition,
				StepCommaPosition:       stepCommaPosition,
			}
		}

		return node
	} else {
		names := &[]Binding{varname}
		varsCommaPosition := &[]lex.Position{}

		if token_type == ',' {
			initialCommaPos := &token_location.Begin
			nextLexeme()
			parseBindingList(names, false, varsCommaPosition, initialCommaPos, nil)
		}

		inLocation := snapshot()
		context := "for loop"
		hasIn := expectAndConsume(lex.ReservedIn, &context)

		values := []AstExpr{}

		valuesCommaPositions := []lex.Position{}
		if storeCstData {
			parseExprList(&values, &valuesCommaPositions)
		} else {
			parseExprList(&values, nil)
		}

		Do_type := token_type
		Do_begin := token_location.Begin
		Do_end := token_location.End

		hasDo := expectAndConsume(lex.ReservedDo, &context)

		localsBegin := len(localStack)
		functionStack[len(functionStack)-1].LoopDepth++

		var vars []*AstLocal
		for _, binding := range *names {
			vars = append(vars, pushLocal(binding))
		}

		body := parseBlock()

		functionStack[len(functionStack)-1].LoopDepth--
		restoreLocals(localsBegin)

		end := token_location.End
		hasEnd := expectMatchEndAndConsume(lex.ReservedEnd, Do_type, Do_begin)
		body.HasEnd = hasEnd

		node := &AstStatForIn{
			NodeLoc:    &NodeLoc{lex.Location{Begin: start.Begin, End: end}},
			Vars:       vars,
			Values:     values,
			Body:       body,
			HasIn:      hasIn,
			InLocation: inLocation,
			HasDo:      hasDo,
			DoLocation: lex.Location{
				Begin: Do_begin,
				End:   Do_end,
			},
		}

		if storeCstData {
			cstNodes[node] = CstStatForIn{
				VarsAnnotationColonPositions: extractAnnotationColonPositions(*names),
				VarsCommaPositions:           *varsCommaPosition,
				ValuesCommaPositions:         *varsCommaPosition, // TODO: check lel
			}
		}

		return node
	}
}

// funcname ::= Name {`.' Name} [`:' Name]
func parseFunctionName(hasRef []bool, debugNameRef *[]*string) *AstExprIndexName {
	if token_type == lex.Name {
		(*debugNameRef)[0] = token_string // TODO: slice bounds
	}

	// parse funcname into a chain of indexing operators
	expr := AstExpr(parseNameExpr("function name"))
	var newExpr *AstExprIndexName // split it into this for better types

	oldRecursionCount := recursionCounter

	for token_type == '.' {
		opPosition := token_location.Begin
		nextLexeme()

		context := "field name"
		name := parseName(&context)

		// while we could concatenate the name chain, for now let's just write the short name
		(*debugNameRef)[0] = &name.Name.Value

		newExpr = &AstExprIndexName{
			NodeLoc:       &NodeLoc{lex.Location{Begin: expr.GetLocation().Begin, End: name.Location.End}},
			Expr:          expr,
			Index:         name.Name.Value,
			IndexLocation: name.Location,
			OpPosition:    opPosition,
			Op:            ',',
		}

		// note: while the parser isn't recursive here, we're generating recursive structures of unbounded depth
		incrementRecursionCounter("function name")
	}

	recursionCounter = oldRecursionCount

	// finish with :
	if token_type == ':' {
		opPosition := token_location.Begin
		nextLexeme()

		context := "method name"
		name := parseName(&context)

		// while we could concatenate the name chain, for now let's just write the short name
		(*debugNameRef)[0] = &name.Name.Value

		newExpr = &AstExprIndexName{
			NodeLoc:       &NodeLoc{lex.Location{Begin: expr.GetLocation().Begin, End: name.Location.End}},
			Expr:          expr,
			Index:         name.Name.Value,
			IndexLocation: name.Location,
			OpPosition:    opPosition,
			Op:            ':',
		}

		hasRef[0] = true // again, todo bounds check
	}

	return newExpr
}

// function funcname funcbody
func parseFunctionStat(attributes Attrs) *AstStatFunction {
	start := snapshot()
	if len(attributes) > 0 {
		start = attributes[0].Location
	}

	matchFunction := get_lexeme()
	nextLexeme()

	hasRef := []bool{false}
	debugnameRef := []*string{nil} // todo length check bruh
	expr := parseFunctionName(hasRef, &debugnameRef)

	matchRecovery[lex.ReservedEnd]++

	body, _ := parseFunctionBody(hasRef[0], matchFunction, debugnameRef[0], nil, attributes)

	matchRecovery[lex.ReservedEnd]--

	node := &AstStatFunction{
		NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: body.GetLocation().End}},
		Name:    expr,
		Func:    body,
	}

	if storeCstData {
		cstNodes[node] = CstStatFunction{
			FunctionKeywordPosition: matchFunction.Location.Begin,
		}
	}

	return node
}

func validateAttribute(loc lex.Location, attributeName string, attributes Attrs, args []AstExpr) *string {
	// checks if the attribute name is valid
	entry, ok := kAttributeEntries[attributeName]
	var type_ *string
	var argsValidator ArgsValidator

	if ok {
		type_ = &entry.Type
		argsValidator = entry.ArgsValidator
	} else {
		if len(attributeName) == 0 {
			report(loc, "Attribute name is missing")
		} else {
			report(loc, fmt.Sprintf("Invalid attribute '@%s'", attributeName))
		}
	}

	if type_ != nil {
		// check that attribute is not duplicated
		for _, attr := range attributes {
			if attr.Type == *type_ {
				report(loc, fmt.Sprintf("Duplicate attribute '@%s'", attributeName))
			}
		}

		if argsValidator != nil {
			errors := argsValidator(loc, args)
			for _, err := range errors {
				report(err.Location, err.Message) // dk about the formatting, guess i'll add a TODO
			}
		}
	}

	return type_
}

// attribute ::= '@' NAME
func parseAttribute(attributes *Attrs) {
	if token_type == lex.Attribute {
		loc := snapshot()
		name := ""
		if token_string != nil {
			name = *token_string
		}
		type_ := validateAttribute(loc, name, *attributes, nil)
		nextLexeme()
		var typ string
		if type_ != nil {
			typ = *type_
		}
		nameCopy := name
		*attributes = append(*attributes, AstAttr{
			NodeLoc: &NodeLoc{loc},
			Type:    typ,
			Args:    nil,
			Name:    &nameCopy,
		})
	} else {
		// AttributeOpen case
		open_type := token_type
		open_begin := token_location.Begin
		open_end := token_location.End
		nextLexeme()
		if token_type != ']' {
			for {
				ctx := "attribute name"
				name_ := parseName(&ctx)
				nameLoc := name_.NodeLoc.Location
				attrName := name_.Name.Value
				var args []AstExpr
				argsLocation := snapshot()

				if token_type == lex.RawString || token_type == lex.QuotedString || token_type == '{' || token_type == '(' {
					var argsOpenLoc lex.Location
					args, argsLocation, argsOpenLoc = parseCallList(nil)
					_ = argsOpenLoc
					for _, arg := range args {
						if !ConstantLiteral(arg) && !isLiteralTable(arg) {
							report(argsLocation, "Only literals can be passed as arguments for attributes")
						}
					}
				}

				validateAttribute(nameLoc, attrName, *attributes, args)

				attrNameCopy := attrName
				*attributes = append(*attributes, AstAttr{
					NodeLoc: &NodeLoc{nameLoc},
					Type:    "Unknown",
					Args:    args,
					Name:    &attrNameCopy,
				})

				if token_type == ',' {
					nextLexeme()
				} else {
					break
				}
			}
		} else {
			report(lex.Location{Begin: open_begin, End: open_end}, "Attribute list cannot be empty")
		}
		expectMatchAndConsume(']', open_type, open_begin, nil)
	}
}

// attributes ::= {attribute}
func parseAttributes() Attrs {
	var attributes Attrs

	for token_type == lex.Attribute || token_type == lex.AttributeOpen {
		parseAttribute(&attributes)
	}

	return attributes
}

// attributes local function Name funcbody
// attributes function funcname funcbody
// attributes `declare function' Name`(' [parlist] `)' [`:` Type]
// declare Name '{' Name ':' attributes `(' [parlist] `)' [`:` Type] '}'
func parseAttributeStat() AstStat {
	attributes := parseAttributes()
	type_ := token_type

	if type_ == lex.ReservedFunction {
		return parseFunctionStat(attributes)
	} else if type_ == lex.ReservedLocal {
		return parseLocal(attributes)
	}

	currLex := lex.Lexeme{Type: token_type, Codepoint: token_codepoint}
	if token_string != nil {
		currLex.Data = []byte(*token_string)
	}
	return reportStatError(
		snapshot(), nil, nil,
		fmt.Sprintf("Expected 'function', 'local function', 'declare function' or a function type declaration after attribute, but got %s instead", currLex.String()),
	)
}

// parseLocal handles `local function Name funcbody | local namelist [`=' explist]
func parseLocal(attributes Attrs) AstStat {
	start := snapshot()
	if len(attributes) > 0 {
		start = attributes[0].Location
	}

	localKeywordPosition := token_location.Begin
	nextLexeme() // consume 'local'

	if token_type == lex.ReservedFunction {
		matchFunction := get_lexeme()
		functionKeywordPosition := matchFunction.Location.Begin
		nextLexeme()

		// Adjust start position
		if len(attributes) > 0 {
			matchFunction.Location.Begin = start.Begin
		}

		ctx := "variable name"
		name := parseName(&ctx)

		matchRecovery[lex.ReservedEnd]++

		debugname := name.Name.Value
		body, funLocal := parseFunctionBody(false, matchFunction, &debugname, &debugname, attributes)

		matchRecovery[lex.ReservedEnd]--

		var varLocal AstLocal
		if funLocal != nil {
			varLocal = *funLocal
		}

		node := &AstStatLocalFunction{
			NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: body.GetLocation().End}},
			Name:    varLocal,
			Func:    body,
		}

		if storeCstData {
			cstNodes[node] = CstStatLocalFunction{
				LocalKeywordPosition:    localKeywordPosition,
				FunctionKeywordPosition: functionKeywordPosition,
			}
		}

		return node
	} else {

		if len(attributes) != 0 {
			currLex := lex.Lexeme{Type: token_type, Codepoint: token_codepoint}
			if token_string != nil {
				currLex.Data = []byte(*token_string)
			}
			return reportStatError(
				snapshot(), nil, nil,
				fmt.Sprintf("Expected 'function' after local declaration with attribute, but got %s instead", currLex.String()),
			)
		}

		matchRecovery['=']++

		var names []Binding
		var varsCommaPositions []lex.Position

		if storeCstData {
			parseBindingList(&names, false, &varsCommaPositions, nil, nil)
		} else {
			parseBindingList(&names, false, nil, nil, nil)
		}

		matchRecovery['=']--

		var values []AstExpr
		var valuesCommaPositions []lex.Position
		var equalsSignLocation *lex.Location

		if token_type == '=' {
			loc := snapshot()
			equalsSignLocation = &loc
			nextLexeme()
			if storeCstData {
				parseExprList(&values, &valuesCommaPositions)
			} else {
				parseExprList(&values, nil)
			}
		}

		// Push all locals after parsing values (correct scoping)
		var vars []AstLocal
		for _, binding := range names {
			vars = append(vars, *pushLocal(binding))
		}

		end := prev_location.End
		if len(values) > 0 {
			end = values[len(values)-1].GetLocation().End
		}

		node := &AstStatLocal{
			NodeLoc:            &NodeLoc{lex.Location{Begin: start.Begin, End: end}},
			Vars:               vars,
			Values:             values,
			EqualsSignLocation: equalsSignLocation,
		}

		if storeCstData {
			cstNodes[node] = CstStatLocal{
				VarsAnnotationColonPositions: extractAnnotationColonPositions(names),
				VarsCommaPositions:           varsCommaPositions,
				ValuesCommaPositions:         valuesCommaPositions,
			}
		}

		return node
	}
}

// parseReturn parses `return [explist]'
func parseReturn() *AstStatReturn {
	start := snapshot()
	nextLexeme()

	var list []AstExpr
	var commaPositions []lex.Position

	if !BlockFollow[token_type] && token_type != ';' {
		if storeCstData {
			parseExprList(&list, &commaPositions)
		} else {
			parseExprList(&list, nil)
		}
	}

	end := start.End
	if len(list) > 0 {
		end = list[len(list)-1].GetLocation().End
	}

	node := &AstStatReturn{
		NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: end}},
		List:    list,
	}

	if storeCstData {
		cstNodes[node] = CstStatReturn{CommaPositions: commaPositions}
	}

	return node
}

// parseTypeAlias parses `type Name [<...>] = Type' or `type function ...'
func parseTypeAlias(start lex.Location, exported bool, typeKeywordPosition lex.Position) AstStatTypeAliasOrTypeFunction {
	if token_type == lex.ReservedFunction {
		return parseTypeFunction(start, exported, typeKeywordPosition)
	}

	ctx := "type name"
	nameOpt := parseNameOpt(&ctx)
	var name Binding
	if nameOpt != nil {
		name = *nameOpt
	} else {
		name = Binding{
			Name:    lex.AstName{Value: nameError},
			NodeLoc: &NodeLoc{snapshot()},
		}
	}

	var genericsOpenPos lex.Position
	var genericsCommaPos []lex.Position
	var genericsClosePos lex.Position
	var genericsOpenPosRef *lex.Position
	var genericsClosePosRef *lex.Position

	if storeCstData {
		genericsOpenPosRef = &genericsOpenPos
		genericsClosePosRef = &genericsClosePos
	}

	generics, genericPacks := parseGenericTypeList(true, genericsOpenPosRef, &genericsCommaPos, genericsClosePosRef)

	equalsPosition := token_location.Begin
	ctx2 := "type alias"
	expectAndConsume('=', &ctx2)

	type_ := parseType(false)
	typeLoc := type_.GetLocation()

	node := &AstStatTypeAlias{
		NodeLoc:      &NodeLoc{lex.Location{Begin: start.Begin, End: typeLoc.End}},
		Name:         name.Name.Value,
		NameLocation: name.Location,
		Generics:     generics,
		GenericPacks: genericPacks,
		Type:         type_,
		Exported:     exported,
	}

	if storeCstData {
		cstNodes[node] = CstStatTypeAlias{
			TypeKeywordPosition:    typeKeywordPosition,
			GenericsOpenPosition:   genericsOpenPosRef,
			GenericsCommaPositions: genericsCommaPos,
			GenericsClosePosition:  genericsClosePosRef,
			EqualsPosition:         equalsPosition,
		}
	}

	return node
}

// parseTypeFunction parses `type function Name funcbody end'
func parseTypeFunction(start lex.Location, exported bool, typeKeywordPosition lex.Position) *AstStatTypeFunction {
	matchFn := get_lexeme()
	nextLexeme()

	errorsAtStart := len(parseErrors)

	ctx := "type function name"
	fnNameOpt := parseNameOpt(&ctx)
	var fnName Binding
	if fnNameOpt != nil {
		fnName = *fnNameOpt
	} else {
		fnName = Binding{
			Name:    lex.AstName{Value: nameError},
			NodeLoc: &NodeLoc{snapshot()},
		}
	}

	matchRecovery[lex.ReservedEnd]++

	oldTypeFunctionDepth := typeFunctionDepth
	typeFunctionDepth = len(functionStack)

	fnNameStr := fnName.Name.Value
	body, _ := parseFunctionBody(false, matchFn, &fnNameStr, nil, Attrs{})

	typeFunctionDepth = oldTypeFunctionDepth
	matchRecovery[lex.ReservedEnd]--

	hasErrors := len(parseErrors) > errorsAtStart

	node := &AstStatTypeFunction{
		NodeLoc:      &NodeLoc{lex.Location{Begin: start.Begin, End: body.GetLocation().End}},
		Name:         fnName.Name.Value,
		NameLocation: fnName.Location,
		Body:         body,
		Exported:     exported,
		HasErrors:    hasErrors,
	}

	if storeCstData {
		cstNodes[node] = CstStatTypeFunction{
			TypeKeywordPosition:     typeKeywordPosition,
			FunctionKeywordPosition: matchFn.Location.Begin,
		}
	}

	return node
}

// parseNameOpt tries to parse a NAME token; returns nil if not a name
func parseNameOpt(context *string) *Binding {
	if token_type != lex.Name {
		reportNameError(context)
		return nil
	}

	value := ""
	if token_string != nil {
		value = *token_string
	}

	result := &Binding{
		Name:    lex.AstName{Value: value},
		NodeLoc: &NodeLoc{snapshot()},
	}

	nextLexeme()
	return result
}

// parseName always produces a Binding (using error token if no name available)
func parseName(context *string) Binding {
	name := parseNameOpt(context)
	if name != nil {
		return *name
	}
	return Binding{
		Name:    lex.AstName{Value: nameError},
		NodeLoc: &NodeLoc{snapshot()},
	}
}

var typeFunctionDepth = 0

var (
	pzero int
	pone  = 1
)

func tableSeparator() *int {
	if token_type == ',' {
		return &pzero
	} else if token_type == ';' {
		return &pone
	} else {
		return nil
	}
}

// explist ::= {exp `,'} exp
func parseExprList(result *[]AstExpr, commaPositions *[]lex.Position) {
	*result = append(*result, parseExpr(0))

	for token_type == ',' {
		if commaPositions != nil {
			*commaPositions = append(*commaPositions, token_location.Begin)
		}
		nextLexeme()

		if token_type == ')' {
			report(snapshot(), "Expected expression after ',' but got ')' instead")
			break
		}

		*result = append(*result, parseExpr(0))
	}
}

// parseAssignment handles varlist `=' explist
func parseAssignment(initial AstExpr) *AstStatAssign {
	if !ExprLValues(initial) {
		initial = reportExprError(initial.GetLocation(), []AstExpr{initial}, "Assigned expression must be a variable or a field")
	}

	vars := []AstExpr{initial}
	var varsCommaPositions []lex.Position

	for token_type == ',' {
		if storeCstData {
			varsCommaPositions = append(varsCommaPositions, token_location.Begin)
		}
		nextLexeme()

		expr := parsePrimaryExpr(true)
		if !ExprLValues(expr) {
			expr = reportExprError(expr.GetLocation(), []AstExpr{expr}, "Assigned expression must be a variable or a field")
		}
		vars = append(vars, expr)
	}

	equalsPosition := token_location.Begin
	context := "assignment"
	expectAndConsume('=', &context)

	var values []AstExpr
	var valuesCommaPositions []lex.Position

	if storeCstData {
		parseExprList(&values, &valuesCommaPositions)
	} else {
		parseExprList(&values, nil)
	}

	endLoc := initial.GetLocation()
	if len(values) > 0 {
		endLoc = values[len(values)-1].GetLocation()
	}

	node := &AstStatAssign{
		NodeLoc: &NodeLoc{lex.Location{Begin: initial.GetLocation().Begin, End: endLoc.End}},
		Vars:    vars,
		Values:  values,
	}

	if storeCstData {
		cstNodes[node] = CstStatAssign{
			VarsCommaPositions:   varsCommaPositions,
			EqualsPosition:       equalsPosition,
			ValuesCommaPositions: valuesCommaPositions,
		}
	}

	return node
}

// parseCompoundAssignment handles compound assignment operators
func parseCompoundAssignment(initial AstExpr, op BinaryOp) *AstStatCompoundAssign {
	if !ExprLValues(initial) {
		initial = reportExprError(initial.GetLocation(), []AstExpr{initial}, "Assigned expression must be a variable or a field")
	}

	opPosition := token_location.Begin
	nextLexeme()

	value := parseExpr(0)

	node := &AstStatCompoundAssign{
		NodeLoc: &NodeLoc{lex.Location{Begin: initial.GetLocation().Begin, End: value.GetLocation().End}},
		Op:      int(op),
		Var:     initial,
		Value:   value,
	}

	if storeCstData {
		cstNodes[node] = CstStatCompoundAssign{
			OpPosition: opPosition,
		}
	}

	return node
}

// prepareFunctionArguments sets up self and regular args as locals
func prepareFunctionArguments(start lex.Location, hasself bool, args []Binding) (*AstLocal, []*AstLocal) {
	var selfLocal *AstLocal
	if hasself {
		selfLocal = pushLocal(Binding{
			Name:    lex.AstName{Value: nameSelf},
			NodeLoc: &NodeLoc{start},
		})
	}

	var vars []*AstLocal
	for _, arg := range args {
		vars = append(vars, pushLocal(arg))
	}

	return selfLocal, vars
}

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

// parseFunctionBody parses funcbody ::= `(' [parlist] `)' [`:' ReturnType] block end
func parseFunctionBody(hasself bool, matchFunction lex.Lexeme, debugname *string, localName *string, attributes Attrs) (AstExprFunction, *AstLocal) {
	start := matchFunction.Location
	if len(attributes) > 0 {
		start = attributes[0].Location
	}

	var cstExprFunc *CstExprFunction
	if storeCstData {
		cstExprFunc = &CstExprFunction{
			FunctionKeywordPosition: matchFunction.Location.Begin,
		}
	}

	// Parse generic type list
	var openGenPos lex.Position
	var closeGenPos lex.Position
	var genCommaPos []lex.Position

	var openGenPosRef *lex.Position
	var closeGenPosRef *lex.Position
	if cstExprFunc != nil {
		openGenPosRef = &openGenPos
		closeGenPosRef = &closeGenPos
	}

	generics, genericPacks := parseGenericTypeList(false, openGenPosRef, &genCommaPos, closeGenPosRef)
	if cstExprFunc != nil {
		if openGenPosRef != nil {
			cstExprFunc.OpenGenericsPosition = openGenPosRef
		}
		if closeGenPosRef != nil {
			cstExprFunc.CloseGenericsPosition = closeGenPosRef
		}
		cstExprFunc.GenericsCommaPositions = genCommaPos
	}

	parenType := token_type
	parenBegin := token_location.Begin

	context := "function"
	expectAndConsume('(', &context)

	matchRecovery[')']++

	var args []Binding
	var vararg bool
	var varargLocation *lex.Location
	var varargAnnotation AstTypePack

	if token_type != ')' {
		var commaPositions *[]lex.Position
		if cstExprFunc != nil {
			commaPositions = &cstExprFunc.ArgsCommaPositions
		}

		var vaAnnotPosSlice []*lex.Position
		var vaAnnotPosSliceRef *[]*lex.Position
		if cstExprFunc != nil {
			vaAnnotPosSlice = []*lex.Position{nil}
			vaAnnotPosSliceRef = &vaAnnotPosSlice
		}

		vararg, varargLocation, varargAnnotation = parseBindingList(&args, true, commaPositions, nil, vaAnnotPosSliceRef)

		if cstExprFunc != nil && len(vaAnnotPosSlice) > 0 {
			cstExprFunc.VarargAnnotationColonPosition = vaAnnotPosSlice[0]
		}
	}

	var argLocation *lex.Location
	if parenType == '(' && token_type == ')' {
		loc := lex.Location{
			Begin: parenBegin,
			End:   token_location.End,
		}
		argLocation = &loc
	}

	searchTrue := true
	expectMatchAndConsume(')', parenType, parenBegin, &searchTrue)
	matchRecovery[')']--

	// Return type
	var retSpecPos lex.Position
	var retSpecPosRef *lex.Position
	if cstExprFunc != nil {
		retSpecPosRef = &retSpecPos
	}
	typelist := parseOptionalReturnType(retSpecPosRef)
	if cstExprFunc != nil {
		cstExprFunc.ReturnSpecifierPosition = retSpecPosRef
	}

	// Push the named function local (localName != nil means local function)
	var funLocal *AstLocal
	if localName != nil {
		funLocal = pushLocal(Binding{
			Name:    lex.AstName{Value: *localName},
			NodeLoc: &NodeLoc{start},
		})
	}

	localsBegin := len(localStack)

	functionStack = append(functionStack, FunctionState{Vararg: vararg, LoopDepth: 0})

	selfLocal, vars := prepareFunctionArguments(start, hasself, args)

	body := parseBlock()

	functionStack = functionStack[:len(functionStack)-1]
	restoreLocals(localsBegin)

	hasEnd := expectMatchEndAndConsume(lex.ReservedEnd, matchFunction.Type, matchFunction.Location.Begin)
	body.HasEnd = hasEnd

	// Convert []*AstLocal to []AstLocal
	argLocals := make([]AstLocal, len(vars))
	for i, v := range vars {
		if v != nil {
			argLocals[i] = *v
		}
	}

	var varargAnn *AstTypePack
	if varargAnnotation != nil {
		varargAnn = &varargAnnotation
	}

	var varargLoc lex.Location
	if varargLocation != nil {
		varargLoc = *varargLocation
	}

	node := AstExprFunction{
		NodeLoc:          &NodeLoc{lex.Location{Begin: start.Begin, End: prev_location.End}},
		Attributes:       []AstAttr(attributes),
		Generics:         generics,
		GenericPacks:     genericPacks,
		Self:             selfLocal,
		Args:             argLocals,
		Vararg:           vararg,
		VarargLocation:   varargLoc,
		Body:             *body,
		FunctionDepth:    len(functionStack),
		ReturnAnnotation: typelist,
		VarargAnnotation: varargAnn,
		ArgLocation:      argLocation,
	}

	if debugname != nil {
		node.Debugname = *debugname
	}

	if storeCstData && cstExprFunc != nil {
		cstExprFunc.ArgsAnnotationColonPositions = extractAnnotationColonPositions(args)
		cstNodes[node] = *cstExprFunc
	}

	return node, funLocal
}

// parseGenericTypeList parses `<' TypeList `>'
func parseGenericTypeList(withDefaultValues bool, openPosRef *lex.Position, commaPosRef *[]lex.Position, closePosRef *lex.Position) ([]AstGenericType, []AstGenericTypePack) {
	var names []AstGenericType
	var namePacks []AstGenericTypePack
	var localCommaPositions []lex.Position

	if token_type == '<' {
		beginType := token_type
		beginPos := token_location.Begin

		if openPosRef != nil {
			*openPosRef = beginPos
		}

		nextLexeme()

		seenPack := false
		seenDefault := false

		for {
			nameLoc := snapshot()
			ctx := ""
			nameBinding := parseName(&ctx)
			name := nameBinding.Name.Value

			if token_type == lex.Dot3 || seenPack {
				seenPack = true
				ellipsisPosition := token_location.Begin

				if token_type != lex.Dot3 {
					report(snapshot(), "Generic types come before generic type packs")
				} else {
					nextLexeme()
				}

				if withDefaultValues && token_type == '=' {
					seenDefault = true
					equalsPosition := token_location.Begin
					nextLexeme()

					var typePack AstTypePack
					if shouldParseTypePack() {
						typePack = parseTypePack()
					} else {
						_, pack_ := parseSimpleTypeOrPack()
						typePack = pack_
					}

					node := AstGenericTypePack{
						NodeLoc:      &NodeLoc{nameLoc},
						Name:         name,
						DefaultValue: &typePack,
					}

					namePacks = append(namePacks, node)
					_ = ellipsisPosition
					_ = equalsPosition
				} else {
					if seenDefault {
						report(snapshot(), "Expected default type pack after type pack name")
					}

					node := AstGenericTypePack{
						NodeLoc:      &NodeLoc{nameLoc},
						Name:         name,
						DefaultValue: nil,
					}

					namePacks = append(namePacks, node)
					_ = ellipsisPosition
				}
			} else {
				if withDefaultValues && token_type == '=' {
					seenDefault = true
					equalsPosition := token_location.Begin
					nextLexeme()

					defaultType := parseType(false)

					node := AstGenericType{
						NodeLoc:      &NodeLoc{nameLoc},
						Name:         name,
						DefaultValue: &defaultType,
					}
					names = append(names, node)
					_ = equalsPosition
				} else {
					if seenDefault {
						report(snapshot(), "Expected default type after type name")
					}

					node := AstGenericType{
						NodeLoc:      &NodeLoc{nameLoc},
						Name:         name,
						DefaultValue: nil,
					}
					names = append(names, node)
				}
			}

			if token_type == ',' {
				localCommaPositions = append(localCommaPositions, token_location.Begin)
				nextLexeme()

				if token_type == '>' {
					report(snapshot(), "Expected type after ',' but got '>' instead")
					break
				}
			} else {
				break
			}
		}

		if closePosRef != nil {
			*closePosRef = token_location.Begin
		}

		expectMatchAndConsume('>', beginType, beginPos, nil)
	}

	if commaPosRef != nil {
		*commaPosRef = append(*commaPosRef, localCommaPositions...)
	}

	return names, namePacks
}

// parseOptionalType parses an optional`: Type' annotation
func parseOptionalType() AstType {
	if token_type == ':' {
		nextLexeme()
		return parseType(false)
	}
	return nil
}

// parseTypeList parses TypeList in function/tuple types
func parseTypeList(result *[]AstType, resultNames *[]*AstArgumentName, commaPositions *[]lex.Position, nameColonPositions *[]*lex.Position) AstTypePack {
	for {
		if shouldParseTypePack() {
			return parseTypePack()
		}

		if token_type == lex.Name && next_type == ':' {
			// Named argument
			for len(*resultNames) < len(*result) {
				*resultNames = append(*resultNames, nil)
				if nameColonPositions != nil {
					*nameColonPositions = append(*nameColonPositions, nil)
				}
			}

			nameStr := ""
			if token_string != nil {
				nameStr = *token_string
			}
			argName := &AstArgumentName{
				Name:     nameStr,
				Location: snapshot(),
			}

			*resultNames = append(*resultNames, argName)
			nextLexeme()

			if nameColonPositions != nil {
				colonPos := token_location.Begin
				*nameColonPositions = append(*nameColonPositions, &colonPos)
			}

			context := ""
			expectAndConsume(':', &context)
		} else if len(*resultNames) > 0 {
			*resultNames = append(*resultNames, nil)
			if nameColonPositions != nil {
				*nameColonPositions = append(*nameColonPositions, nil)
			}
		}

		*result = append(*result, parseType(false))

		if token_type != ',' {
			break
		}

		if commaPositions != nil {
			*commaPositions = append(*commaPositions, token_location.Begin)
		}
		nextLexeme()

		if token_type == ')' {
			report(snapshot(), "Expected type after ',' but got ')' instead")
			break
		}
	}
	return nil
}

// parseOptionalReturnType parses optional return type after `:'
func parseOptionalReturnType(returnSpecifierPosRef *lex.Position) *AstTypePack {
	if token_type == ':' || token_type == lex.SkinnyArrow {
		if token_type == lex.SkinnyArrow {
			report(snapshot(), "Function return type annotations are written after ':' instead of '->'")
		}

		if returnSpecifierPosRef != nil {
			*returnSpecifierPosRef = token_location.Begin
		}

		nextLexeme()

		oldRecursion := recursionCounter
		res := parseReturnType()
		recursionCounter = oldRecursion

		if token_type == ',' {
			report(snapshot(), "Expected a statement, got ','; did you forget to wrap the list of return types in parentheses?")
			nextLexeme()
		}

		return &res
	}

	return nil
}

// parseReturnType ::= Type | `(' TypeList `)'
func parseReturnType() AstTypePack {
	incrementRecursionCounter("type annotation")

	begin := get_lexeme()
	beginType := token_type
	beginPos := token_location.Begin

	if token_type != '(' {
		if shouldParseTypePack() {
			return parseTypePack()
		}

		type_ := parseType(false)
		typeLoc := type_.GetLocation()

		var openPos *lex.Position
		var closePos *lex.Position
		node := AstTypePackExplicit{
			NodeLoc: &NodeLoc{typeLoc},
			Types:   AstTypeList{Types: []AstType{type_}},
		}

		if storeCstData {
			cstNodes[node] = CstTypePackExplicit{
				OpenParenthesesPosition:  openPos,
				CloseParenthesesPosition: closePos,
			}
		}
		return node
	}

	nextLexeme()
	matchRecovery[lex.SkinnyArrow]++

	var result []AstType
	var resultNames []*AstArgumentName
	var commaPositions []lex.Position
	var nameColonPositions []*lex.Position
	var varargAnnotation AstTypePack

	if token_type != ')' {
		if storeCstData {
			varargAnnotation = parseTypeList(&result, &resultNames, &commaPositions, &nameColonPositions)
		} else {
			varargAnnotation = parseTypeList(&result, &resultNames, nil, nil)
		}
	}

	closeParenPos := token_location.Begin

	searchTrue := true
	expectMatchAndConsume(')', beginType, beginPos, &searchTrue)

	matchRecovery[lex.SkinnyArrow]--

	if token_type != lex.SkinnyArrow && len(resultNames) == 0 {
		if len(result) == 1 {
			var inner AstType
			if varargAnnotation == nil {
				inner = AstTypeGroup{
					NodeLoc: &NodeLoc{lex.Location{Begin: begin.Location.Begin, End: closeParenPos}},
					Type:    result[0],
				}
			} else {
				inner = result[0]
			}

			returnType := parseTypeSuffix(inner, begin.Location)
			retLoc := returnType.GetLocation()

			endPos := retLoc.End

			var tailTypePtr *AstTypePack
			if varargAnnotation != nil {
				tailTypePtr = &varargAnnotation
			}

			openPos := begin.Location.Begin
			node := AstTypePackExplicit{
				NodeLoc: &NodeLoc{lex.Location{Begin: begin.Location.Begin, End: endPos}},
				Types:   AstTypeList{Types: []AstType{returnType}, TailType: tailTypePtr},
			}

			if storeCstData {
				cp := commaPositions
				cstNodes[node] = CstTypePackExplicit{
					OpenParenthesesPosition:  &openPos,
					CloseParenthesesPosition: &closeParenPos,
					CommaPositions:           &cp,
				}
			}
			return node
		}

		var tailPtr *AstTypePack
		if varargAnnotation != nil {
			tailPtr = &varargAnnotation
		}

		openPos := begin.Location.Begin
		endPos := closeParenPos

		if len(result) > 0 {
			endPos = result[len(result)-1].GetLocation().End
		}

		node := AstTypePackExplicit{
			NodeLoc: &NodeLoc{lex.Location{Begin: begin.Location.Begin, End: endPos}},
			Types:   AstTypeList{Types: result, TailType: tailPtr},
		}

		if storeCstData {
			cp := commaPositions
			cstNodes[node] = CstTypePackExplicit{
				OpenParenthesesPosition:  &openPos,
				CloseParenthesesPosition: &closeParenPos,
				CommaPositions:           &cp,
			}
		}
		return node
	}

	returnArrowPosition := token_location.Begin

	var tailPtr *AstTypePack
	if varargAnnotation != nil {
		tailPtr = &varargAnnotation
	}

	tail := parseFunctionTypeTail(begin, Attrs{}, []AstGenericType{}, []AstGenericTypePack{}, result, resultNames, tailPtr)
	tailLoc := tail.GetLocation()

	openPos := begin.Location.Begin
	node := AstTypePackExplicit{
		NodeLoc: &NodeLoc{lex.Location{Begin: begin.Location.Begin, End: tailLoc.End}},
		Types:   AstTypeList{Types: []AstType{tail}},
	}

	if storeCstData {
		cp := commaPositions
		cstNodes[node] = CstTypePackExplicit{
			OpenParenthesesPosition:  &openPos,
			CloseParenthesesPosition: &closeParenPos,
			CommaPositions:           &cp,
		}

		// Override function type CST with return-type position info
		cstNodes[tail] = CstTypeFunction{
			OpenArgsPosition:           begin.Location.Begin,
			ArgumentNameColonPositions: nameColonPositions,
			ArgumentsCommaPositions:    commaPositions,
			CloseArgsPosition:          closeParenPos,
			ReturnArrowPosition:        returnArrowPosition,
		}
	}

	return node
}

func extractStringDetails() (style CstQuotes, depth int) {
	if token_type == lex.QuotedString {
		if token_aux != nil && *token_aux == 1 {
			style = CstQuotes_Single
		} else {
			style = CstQuotes_Double
		}
	} else if token_type == lex.InterpStringSimple {
		style = CstQuotes_Interp
	} else if token_type == lex.RawString {
		style = CstQuotes_Raw
		if token_aux != nil {
			depth = *token_aux
		}
	}

	return
}

// parseTableIndexer parses `[' Type `]' `:' Type
func parseTableIndexer(access string, accessLoc *lex.Location, begin lex.Lexeme) parseTableIndexerResult {
	index := parseType(false)

	indexerClosePos := token_location.Begin
	expectMatchAndConsume(']', begin.Type, begin.Location.Begin, nil)

	colonPos := token_location.Begin
	context := "table field"
	expectAndConsume(':', &context)

	result := parseType(false)
	resultLoc := result.GetLocation()

	node := AstTableIndexer{
		Location:       lex.Location{Begin: begin.Location.Begin, End: resultLoc.End},
		IndexType:      index,
		ResultType:     result,
		Access:         access,
		AccessLocation: accessLoc,
	}

	return parseTableIndexerResult{
		node:                 node,
		indexerOpenPosition:  begin.Location.Begin,
		indexerClosePosition: indexerClosePos,
		colonPosition:        colonPos,
	}
}

// parseTableType parses `{' PropList `}'
func parseTableType(inDeclarationContext bool) AstTypeTable {
	incrementRecursionCounter("type annotation")

	var props []AstTableProp
	var cstItems []CstTypeTableItem
	var indexer *AstTableIndexer

	start := snapshot()
	matchBrace := get_lexeme()
	context := "table type"
	expectAndConsume('{', &context)

	for token_type != '}' {
		access := "ReadWrite"
		var accessLoc *lex.Location

		if token_type == lex.Name && next_type != ':' {
			if token_string != nil {
				if *token_string == "read" {
					loc := snapshot()
					accessLoc = &loc
					access = "Read"
					nextLexeme()
				} else if *token_string == "write" {
					loc := snapshot()
					accessLoc = &loc
					access = "Write"
					nextLexeme()
				}
			}
		}

		if token_type == '[' {
			begin := get_lexeme()
			nextLexeme()

			if (token_type == lex.RawString || token_type == lex.QuotedString) && next_type == ']' {
				var cstStr *CstExprConstantString
				var stringPos *lex.Location
				if storeCstData {
					style, depth := extractStringDetails()
					sp := snapshot()
					stringPos = &sp
					cstStr = &CstExprConstantString{
						SourceString: token_string,
						QuoteStyle:   int(style),
						BlockDepth:   depth,
					}
				}

				chars := parseCharArray()

				indexerClosePos := token_location.Begin
				expectMatchAndConsume(']', begin.Type, begin.Location.Begin, nil)

				colonPos := token_location.Begin
				context2 := "table field"
				expectAndConsume(':', &context2)

				type_ := parseType(inDeclarationContext)
				typeLoc := type_.GetLocation()

				if chars != nil {
					props = append(props, AstTableProp{
						Name:           lex.AstName{Value: *chars},
						NodeLoc:        &NodeLoc{begin.Location},
						Type:           type_,
						Access:         access,
						AccessLocation: accessLoc,
					})

					if storeCstData {
						sepPos := token_location.Begin
						openPos := begin.Location.Begin
						closePos := indexerClosePos
						cstItems = append(cstItems, CstTypeTableItem{
							Kind:                 "StringProperty",
							IndexerOpenPosition:  &openPos,
							IndexerClosePosition: &closePos,
							ColonPosition:        &colonPos,
							Separator:            tableSeparator(),
							SeparatorPosition:    &sepPos,
							StringInfo:           cstStr,
							StringPosition:       stringPos,
						})
					}
					_ = typeLoc
				} else {
					report(begin.Location, "String literal contains malformed escape sequence or \\0")
				}
			} else {
				if indexer != nil {
					badIdxRes := parseTableIndexer(access, accessLoc, begin)
					report(badIdxRes.node.Location, "Cannot have more than one table indexer")
				} else {
					idxRes := parseTableIndexer(access, accessLoc, begin)
					indexer = &idxRes.node

					if storeCstData {
						sepPos := token_location.Begin
						openPos := idxRes.indexerOpenPosition
						closePos := idxRes.indexerClosePosition
						colonPos := idxRes.colonPosition
						cstItems = append(cstItems, CstTypeTableItem{
							Kind:                 "Indexer",
							IndexerOpenPosition:  &openPos,
							IndexerClosePosition: &closePos,
							ColonPosition:        &colonPos,
							Separator:            tableSeparator(),
							SeparatorPosition:    &sepPos,
						})
					}
				}
			}
		} else if len(props) == 0 && indexer == nil && !(token_type == lex.Name && next_type == ':') {
			// Array-style table type
			type_ := parseType(false)
			typeLoc := type_.GetLocation()

			indexLocation := typeLoc
			if DesugaredArrayTypeReferenceIsEmpty {
				indexLocation = lex.Location{Begin: start.Begin, End: start.Begin}
			}

			index := AstTypeReference{
				NodeLoc:          &NodeLoc{indexLocation},
				HasParameterList: false,
				Name:             nameNumber,
				NameLocation:     indexLocation,
			}

			idxVal := AstTableIndexer{
				Location:       typeLoc,
				IndexType:      index,
				ResultType:     type_,
				Access:         access,
				AccessLocation: accessLoc,
			}
			indexer = &idxVal
			break
		} else {
			ctx := "table field"
			nameOpt := parseNameOpt(&ctx)
			if nameOpt == nil {
				break
			}

			colonPos := token_location.Begin
			ctx2 := "table field"
			expectAndConsume(':', &ctx2)

			type_ := parseType(inDeclarationContext)
			typeLoc := type_.GetLocation()
			_ = typeLoc

			props = append(props, AstTableProp{
				Name:           nameOpt.Name,
				NodeLoc:        nameOpt.NodeLoc,
				Type:           type_,
				Access:         access,
				AccessLocation: accessLoc,
			})

			if storeCstData {
				sepPos := token_location.Begin
				cstItems = append(cstItems, CstTypeTableItem{
					Kind:              "Property",
					ColonPosition:     &colonPos,
					Separator:         tableSeparator(),
					SeparatorPosition: &sepPos,
				})
			}
		}

		if token_type == ',' || token_type == ';' {
			nextLexeme()
		} else if token_type != '}' {
			break
		}
	}

	endLoc := snapshot()
	searchTrue := true
	expectMatchAndConsume('}', matchBrace.Type, matchBrace.Location.Begin, &searchTrue)

	node := AstTypeTable{
		NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: endLoc.End}},
		Props:   props,
		Indexer: indexer,
	}

	if storeCstData {
		cstNodes[node] = CstTypeTable{
			Items:   cstItems,
			IsArray: indexer != nil && len(props) == 0,
		}
	}

	return node
}

// parseFunctionType parses function types
func parseFunctionType(allowPack bool, attributes Attrs) (AstType, AstTypePack) {
	incrementRecursionCounter("type annotation")

	forceFunctionType := token_type == '<'
	begin := get_lexeme()

	var openGenPos lex.Position
	var genCommaPos []lex.Position
	var closeGenPos lex.Position
	var openGenPosRef *lex.Position
	var closeGenPosRef *lex.Position

	if storeCstData {
		openGenPosRef = &openGenPos
		closeGenPosRef = &closeGenPos
	}

	generics, genericPacks := parseGenericTypeList(false, openGenPosRef, &genCommaPos, closeGenPosRef)

	paramStart := get_lexeme()
	context := "function parameters"
	expectAndConsume('(', &context)

	matchRecovery[lex.SkinnyArrow]++

	var params []AstType
	var names []*AstArgumentName
	var argCommaPos []lex.Position
	var nameColonPos []*lex.Position
	var varargAnnotation AstTypePack

	if token_type != ')' {
		if storeCstData {
			varargAnnotation = parseTypeList(&params, &names, &argCommaPos, &nameColonPos)
		} else {
			varargAnnotation = parseTypeList(&params, &names, nil, nil)
		}
	}

	closeArgsPos := token_location.Begin
	searchTrue := true
	expectMatchAndConsume(')', paramStart.Type, paramStart.Location.Begin, &searchTrue)

	matchRecovery[lex.SkinnyArrow]--

	if len(names) > 0 {
		forceFunctionType = true
	}

	returnTypeIntroducer := token_type == lex.SkinnyArrow || token_type == ':'

	if len(params) == 1 && varargAnnotation == nil && !forceFunctionType && !returnTypeIntroducer {
		if allowPack {
			openPos := paramStart.Location.Begin
			cp := argCommaPos
			node := AstTypePackExplicit{
				NodeLoc: &NodeLoc{lex.Location{Begin: paramStart.Location.Begin, End: closeArgsPos}},
				Types:   AstTypeList{Types: params},
			}

			if storeCstData {
				cstNodes[node] = CstTypePackExplicit{
					OpenParenthesesPosition:  &openPos,
					CloseParenthesesPosition: &closeArgsPos,
					CommaPositions:           &cp,
				}
			}

			return nil, node
		}

		return AstTypeGroup{
			NodeLoc: &NodeLoc{lex.Location{Begin: paramStart.Location.Begin, End: closeArgsPos}},
			Type:    params[0],
		}, nil
	}

	if !forceFunctionType && !returnTypeIntroducer && allowPack {
		var tailPtr *AstTypePack
		if varargAnnotation != nil {
			tailPtr = &varargAnnotation
		}

		openPos := paramStart.Location.Begin
		cp := argCommaPos
		node := AstTypePackExplicit{
			NodeLoc: &NodeLoc{lex.Location{Begin: paramStart.Location.Begin, End: closeArgsPos}},
			Types:   AstTypeList{Types: params, TailType: tailPtr},
		}

		if storeCstData {
			cstNodes[node] = CstTypePackExplicit{
				OpenParenthesesPosition:  &openPos,
				CloseParenthesesPosition: &closeArgsPos,
				CommaPositions:           &cp,
			}
		}

		return nil, node
	}

	returnArrowPosition := token_location.Begin

	var tailPtr *AstTypePack
	if varargAnnotation != nil {
		tailPtr = &varargAnnotation
	}

	node := parseFunctionTypeTail(begin, attributes, generics, genericPacks, params, names, tailPtr)

	if storeCstData {
		cstNodes[node] = CstTypeFunction{
			OpenGenericsPosition:       openGenPosRef,
			GenericsCommaPositions:     genCommaPos,
			CloseGenericsPosition:      closeGenPosRef,
			OpenArgsPosition:           paramStart.Location.Begin,
			ArgumentNameColonPositions: nameColonPos,
			ArgumentsCommaPositions:    argCommaPos,
			CloseArgsPosition:          closeArgsPos,
			ReturnArrowPosition:        returnArrowPosition,
		}
	}

	return node, nil
}

// parseFunctionTypeTail completes a function type after params are parsed
func parseFunctionTypeTail(begin lex.Lexeme, attributes Attrs, generics []AstGenericType, genericPacks []AstGenericTypePack, params []AstType, paramNames []*AstArgumentName, varargAnnotation *AstTypePack) AstType {
	incrementRecursionCounter("type annotation")

	if token_type == ':' {
		report(snapshot(), "Return types in function type annotations are written after '->' instead of ':'")
		nextLexeme()
	} else if token_type != lex.SkinnyArrow && len(generics) == 0 && len(genericPacks) == 0 && len(params) == 0 {
		report(lex.Location{Begin: begin.Location.Begin, End: prev_location.End},
			"Expected '->' after '()' when parsing function type; did you mean 'nil'?")

		return AstTypeReference{
			NodeLoc:          &NodeLoc{begin.Location},
			HasParameterList: false,
			Name:             nameNil,
			NameLocation:     begin.Location,
		}
	} else {
		context := "function type"
		expectAndConsume(lex.SkinnyArrow, &context)
	}

	returnType := parseReturnType()
	retTypeLoc := returnType.GetLocation()

	retTypePack, ok := returnType.(AstTypePackExplicit)
	if !ok {
		retTypePack = AstTypePackExplicit{
			NodeLoc: &NodeLoc{retTypeLoc},
			Types:   AstTypeList{},
		}
	}

	return AstTypeFunction{
		NodeLoc:      &NodeLoc{lex.Location{Begin: begin.Location.Begin, End: retTypeLoc.End}},
		Attributes:   []AstAttr(attributes),
		Generics:     generics,
		GenericPacks: genericPacks,
		ArgTypes:     AstTypeList{Types: params, TailType: varargAnnotation},
		ArgNames:     paramNames,
		ReturnTypes:  retTypePack,
	}
}

type parseTableIndexerResult struct {
	node                                                     AstTableIndexer
	indexerOpenPosition, indexerClosePosition, colonPosition lex.Position
}

// parseTypeSuffix parses union (`|'), intersection (`&') and optional (`?') suffixes
func parseTypeSuffix(type_ AstType, begin lex.Location) AstType {
	var parts []AstType
	if type_ != nil {
		parts = append(parts, type_)
	}

	incrementRecursionCounter("type annotation")

	isUnion := false
	isIntersection := false
	optionalCount := 0

	var separatorPositions []lex.Position
	var leadingPosition *lex.Position

	for {
		t := token_type
		separatorPosition := token_location.Begin

		if t == '|' {
			nextLexeme()

			oldRecursion := recursionCounter
			typePart, _ := parseSimpleType(false, false)
			recursionCounter = oldRecursion

			if typePart != nil {
				parts = append(parts, typePart)
			}

			isUnion = true

			if storeCstData {
				if type_ == nil && leadingPosition == nil {
					leadingPosition = &separatorPosition
				} else {
					separatorPositions = append(separatorPositions, separatorPosition)
				}
			}
		} else if t == '?' {
			loc := snapshot()
			nextLexeme()

			parts = append(parts, AstTypeOptional{NodeLoc: &NodeLoc{loc}})
			optionalCount++
			isUnion = true
		} else if t == '&' {
			nextLexeme()

			oldRecursion := recursionCounter
			typePart, _ := parseSimpleType(false, false)
			recursionCounter = oldRecursion

			if typePart != nil {
				parts = append(parts, typePart)
			}

			isIntersection = true

			if storeCstData {
				if type_ == nil && leadingPosition == nil {
					leadingPosition = &separatorPosition
				} else {
					separatorPositions = append(separatorPositions, separatorPosition)
				}
			}
		} else if t == lex.Dot3 {
			report(snapshot(), "Unexpected '...' after type annotation")
			nextLexeme()
		} else {
			break
		}

		if len(parts) > TypeLengthLimit+optionalCount {
			report(parts[len(parts)-1].GetLocation(), "Exceeded allowed type length; simplify your type annotation to make the code compile")
		}
	}

	if len(parts) == 1 && !isUnion && !isIntersection {
		return parts[0]
	}

	if isUnion && isIntersection {
		reportTypeError(
			lex.Location{Begin: begin.Begin, End: parts[len(parts)-1].GetLocation().End},
			parts,
			"Mixing union and intersection types is not allowed; consider wrapping in parentheses.",
		)
	}

	if len(parts) == 0 {
		return AstTypeError{
			NodeLoc:      &NodeLoc{begin},
			IsMissing:    true,
			MessageIndex: len(parseErrors),
		}
	}

	loc := lex.Location{Begin: begin.Begin, End: parts[len(parts)-1].GetLocation().End}

	if isUnion {
		node := AstTypeUnion{NodeLoc: &NodeLoc{loc}, Types: parts}
		if storeCstData {
			cstNodes[node] = CstTypeUnion{
				LeadingPosition:    leadingPosition,
				SeparatorPositions: separatorPositions,
			}
		}
		return node
	}

	node := AstTypeIntersection{NodeLoc: &NodeLoc{loc}, Types: parts}
	if storeCstData {
		cstNodes[node] = CstTypeIntersection{
			LeadingPosition:    leadingPosition,
			SeparatorPositions: separatorPositions,
		}
	}
	return node
}

// parseSimpleTypeOrPack parses a single type (possibly pack if followed by `...')
func parseSimpleTypeOrPack() (AstType, AstTypePack) {
	begin := snapshot()
	type_, typePack := parseSimpleType(true, false)
	if typePack != nil {
		return nil, typePack
	}
	return parseTypeSuffix(type_, begin), nil
}

// parseType parses a full type expression
func parseType(inDeclarationContext bool) AstType {
	begin := snapshot()

	var type_ AstType
	if token_type != '|' && token_type != '&' {
		type_, _ = parseSimpleType(false, inDeclarationContext)
	}

	return parseTypeSuffix(type_, begin)
}

// parseSimpleType parses an atomic type, possibly returning a type pack
func parseSimpleType(allowPack bool, inDeclarationContext bool) (AstType, AstTypePack) {
	incrementRecursionCounter("type annotation")

	start := snapshot()

	if token_type == lex.Attribute || token_type == lex.AttributeOpen {
		attributes := parseAttributes()
		return parseFunctionType(allowPack, attributes)
	} else if token_type == lex.ReservedNil {
		nextLexeme()
		return AstTypeReference{
			NodeLoc:          &NodeLoc{start},
			HasParameterList: false,
			Name:             nameNil,
			NameLocation:     start,
		}, nil
	} else if token_type == lex.ReservedTrue {
		nextLexeme()
		return AstTypeSingletonBool{NodeLoc: &NodeLoc{start}, Value: true}, nil
	} else if token_type == lex.ReservedFalse {
		nextLexeme()
		return AstTypeSingletonBool{NodeLoc: &NodeLoc{start}, Value: false}, nil
	} else if token_type == lex.RawString || token_type == lex.QuotedString {
		chars := parseCharArray()
		if chars != nil {
			return AstTypeSingletonString{NodeLoc: &NodeLoc{start}, Value: *chars}, nil
		}
		return reportTypeError(start, nil, "String literal contains malformed escape sequence"), nil
	} else if token_type == lex.InterpStringBegin || token_type == lex.InterpStringSimple {
		parseInterpString()
		return reportTypeError(start, nil, "Interpolated string literals cannot be used as types"), nil
	} else if token_type == lex.BrokenString {
		nextLexeme()
		return reportTypeError(start, nil, "Malformed string; did you forget to finish it?"), nil
	} else if token_type == lex.Name {
		ctx := "type name"
		name := parseName(&ctx)
		var prefix *string
		var prefixLoc *lex.Location
		var prefixPointPos *lex.Position

		if token_type == '.' {
			pos := token_location.Begin
			prefixPointPos = &pos
			nextLexeme()
			nameCopy := name.Name.Value
			prefix = &nameCopy
			loc := name.Location
			prefixLoc = &loc
			ctx2 := "field name"
			name = parseIndexName(&ctx2, pos)
		} else if token_type == lex.Dot3 {
			report(snapshot(), "Unexpected '...' after type name; type pack is not allowed in this context")
			nextLexeme()
		} else if name.Name.Value == "typeof" {
			typeofBegin := get_lexeme()
			ctx3 := "typeof type"
			expectAndConsume('(', &ctx3)
			expr := parseExpr(0)
			endLoc := token_location
			expectMatchAndConsume(')', typeofBegin.Type, typeofBegin.Location.Begin, nil)

			node := AstTypeTypeof{
				NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: endLoc.End}},
				Expr:    expr,
			}

			if storeCstData {
				cstNodes[node] = CstTypeTypeof{
					OpenPosition:  typeofBegin.Location.Begin,
					ClosePosition: endLoc.Begin,
				}
			}
			return node, nil
		}

		hasParams := false
		var params []AstTypeOrPack

		var openPos lex.Position
		var commaPos []lex.Position
		var closePos lex.Position
		var openPosRef *lex.Position
		var closePosRef *lex.Position
		if storeCstData {
			openPosRef = &openPos
			closePosRef = &closePos
		}

		if token_type == '<' {
			hasParams = true
			params = parseTypeParams(openPosRef, &commaPos, closePosRef)
		}

		node := AstTypeReference{
			NodeLoc:          &NodeLoc{lex.Location{Begin: start.Begin, End: prev_location.End}},
			HasParameterList: hasParams,
			Prefix:           prefix,
			PrefixLocation:   prefixLoc,
			Name:             name.Name.Value,
			NameLocation:     name.Location,
			Parameters:       params,
		}

		if storeCstData {
			cstNodes[node] = CstTypeReference{
				PrefixPointPosition:      prefixPointPos,
				OpenParametersPosition:   openPosRef,
				ParametersCommaPositions: commaPos,
				CloseParametersPosition:  closePosRef,
			}
		}

		_ = inDeclarationContext
		return node, nil
	} else if token_type == '{' {
		return parseTableType(inDeclarationContext), nil
	} else if token_type == '(' || token_type == '<' {
		return parseFunctionType(allowPack, Attrs{})
	} else if token_type == lex.ReservedFunction {
		nextLexeme()
		return reportTypeError(start, nil, "Using 'function' as a type annotation is not supported, consider using a typed function decorator instead"), nil
	} else {
		currLex := lex.Lexeme{Type: token_type, Codepoint: token_codepoint}
		if token_string != nil {
			currLex.Data = []byte(*token_string)
		}
		report(start, fmt.Sprintf("Expected type, got %s", currLex.String()))

		return AstTypeError{
			NodeLoc:      &NodeLoc{start},
			Types:        nil,
			IsMissing:    true,
			MessageIndex: len(parseErrors),
		}, nil
	}
}

// parseVariadicArgumentTypePack parses T... or Name...
func parseVariadicArgumentTypePack() AstTypePackVariadicOrGeneric {
	if token_type == lex.Name && next_type == lex.Dot3 {
		ctx := "generic name"
		name := parseName(&ctx)
		ellipsisPos := token_location.Begin
		nextLexeme() // consume ...

		node := AstTypePackGeneric{
			NodeLoc:     &NodeLoc{lex.Location{Begin: name.Location.Begin, End: prev_location.End}},
			GenericName: name.Name.Value,
		}

		if storeCstData {
			cstNodes[node] = CstTypePackGeneric{
				EllipsisPosition: ellipsisPos,
			}
		}

		return node
	}

	varTy := parseType(false)
	varLoc := varTy.GetLocation()
	return AstTypePackVariadic{
		NodeLoc:      &NodeLoc{varLoc},
		VariadicType: varTy,
	}
}

// parseTypePack parses `...' Type or Name `...'
func parseTypePack() AstTypePackVariadicOrGeneric {
	if token_type == lex.Dot3 {
		start := snapshot()
		nextLexeme()
		varTy := parseType(false)
		varLoc := varTy.GetLocation()
		return AstTypePackVariadic{
			NodeLoc:      &NodeLoc{lex.Location{Begin: start.Begin, End: varLoc.End}},
			VariadicType: varTy,
		}
	} else if token_type == lex.Name && next_type == lex.Dot3 {
		ctx := "generic name"
		name := parseName(&ctx)
		ellipsisPos := token_location.Begin
		nextLexeme() // consume ...

		node := AstTypePackGeneric{
			NodeLoc:     &NodeLoc{lex.Location{Begin: name.Location.Begin, End: prev_location.End}},
			GenericName: name.Name.Value,
		}

		if storeCstData {
			cstNodes[node] = CstTypePackGeneric{
				EllipsisPosition: ellipsisPos,
			}
		}

		return node
	}

	panic("parseTypePack called when shouldParseTypePack() is false")
}

// parseTypeParams parses `<' TypeOrPack `>'
func parseTypeParams(openingPosRef *lex.Position, commaPosRef *[]lex.Position, closingPosRef *lex.Position) []AstTypeOrPack {
	var params []AstTypeOrPack

	if token_type == '<' {
		begin := get_lexeme()
		if openingPosRef != nil {
			*openingPosRef = begin.Location.Begin
		}

		nextLexeme()

		for {
			if shouldParseTypePack() {
				pack := parseTypePack()
				typePack := AstTypePack(pack)
				params = append(params, AstTypeOrPack{Pack: &typePack})
			} else if token_type == '(' {
				beginParen := snapshot()
				type_, typePack := parseSimpleType(true, false)

				if typePack != nil {
					if explicit, ok := typePack.(AstTypePackExplicit); ok &&
						len(explicit.Types.Types) == 1 &&
						explicit.Types.TailType == nil &&
						(token_type == '|' || token_type == '?' || token_type == '&') {
						parenTy := explicit.Types.Types[0]

						inner := AstTypeGroup{
							NodeLoc: &NodeLoc{parenTy.GetLocation()},
							Type:    parenTy,
						}

						t2 := parseTypeSuffix(inner, beginParen)
						params = append(params, AstTypeOrPack{Type: &t2})
					} else {
						params = append(params, AstTypeOrPack{Pack: &typePack})
					}
				} else {
					t2 := parseTypeSuffix(type_, beginParen)
					params = append(params, AstTypeOrPack{Type: &t2})
				}
			} else if token_type == '>' && len(params) == 0 {
				break
			} else {
				t := parseType(false)
				params = append(params, AstTypeOrPack{Type: &t})
			}

			if token_type == ',' {
				if commaPosRef != nil {
					*commaPosRef = append(*commaPosRef, token_location.Begin)
				}
				nextLexeme()
			} else {
				break
			}
		}

		if closingPosRef != nil {
			*closingPosRef = token_location.Begin
		}

		expectMatchAndConsume('>', begin.Type, begin.Location.Begin, nil)
	}
	return params
}

var unaryOpNot = UnaryOp_Not

func checkUnaryConfusables() *UnaryOp {
	// early-out: need to check if this is a possible confusable quickly
	if token_type != '!' {
		return nil
	}

	report(snapshot(), "Unexpected '!'; did you mean 'not'?")

	return &unaryOpNot
}

// checkBinaryConfusables checks for `&&', `||', `!=' confusables
func checkBinaryConfusables(limit int) *BinaryOp {
	curr := get_lexeme()

	if curr.Type != '&' && curr.Type != '|' && curr.Type != '!' {
		return nil
	}

	start := curr.Location

	if curr.Type == '&' && next_type == '&' &&
		curr.Location.End.Column == next_location.End.Column &&
		BinaryPriority[BinaryOp_And][0] > limit {
		nextLexeme()
		report(lex.Location{Begin: start.Begin, End: next_location.End}, "Unexpected '&&'; did you mean 'and'?")
		op := BinaryOp_And
		return &op
	} else if curr.Type == '|' && next_type == '|' &&
		curr.Location.End.Column == next_location.End.Column &&
		BinaryPriority[BinaryOp_Or][0] > limit {
		nextLexeme()
		report(lex.Location{Begin: start.Begin, End: next_location.End}, "Unexpected '||'; did you mean 'or'?")
		op := BinaryOp_Or
		return &op
	} else if curr.Type == '!' && next_type == '=' &&
		curr.Location.End.Column == next_location.End.Column &&
		BinaryPriority[BinaryOp_CompareNe][0] > limit {
		nextLexeme()
		report(lex.Location{Begin: start.Begin, End: next_location.End}, "Unexpected '!='; did you mean '~='?")
		op := BinaryOp_CompareNe
		return &op
	}

	return nil
}

// parseExpr parses binary expressions at priority > limit
func parseExpr(limit int) AstExpr {
	oldRecursion := recursionCounter
	incrementRecursionCounter("expression")

	start := snapshot()
	var expr AstExpr

	uop, hasUop := UnaryOpLookup[token_type]
	if !hasUop {
		if confusable := checkUnaryConfusables(); confusable != nil {
			uop = *confusable
			hasUop = true
		}
	}

	if hasUop {
		opPosition := token_location.Begin
		nextLexeme()

		subexpr := parseExpr(8)

		node := AstExprUnary{
			NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: subexpr.GetLocation().End}},
			Op:      uop,
			Expr:    subexpr,
		}

		if storeCstData {
			cstNodes[node] = CstExprOp{OpPosition: opPosition}
		}

		expr = node
	} else {
		expr = parseAssertionExpr()
	}

	op, hasOp := BinaryOpLookup[token_type]
	if !hasOp {
		if confusable := checkBinaryConfusables(limit); confusable != nil {
			op = *confusable
			hasOp = true
		}
	}

	for hasOp && BinaryPriority[op][0] > limit {
		opPosition := token_location.Begin
		nextLexeme()

		nextExpr := parseExpr(BinaryPriority[op][1])

		node := AstExprBinary{
			NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: nextExpr.GetLocation().End}},
			Op:      int(op),
			Left:    expr,
			Right:   nextExpr,
		}

		if storeCstData {
			cstNodes[node] = CstExprOp{OpPosition: opPosition}
		}

		expr = node

		op, hasOp = BinaryOpLookup[token_type]
		if !hasOp {
			if confusable := checkBinaryConfusables(limit); confusable != nil {
				op = *confusable
				hasOp = true
			}
		}

		incrementRecursionCounter("expression")
	}

	recursionCounter = oldRecursion
	return expr
}

// parseNameExpr parses a NAME reference, resolving to local, global, or error
func parseNameExpr(context string) AstExprLocalOrGlobalOrError {
	nameOpt := parseNameOpt(&context)

	if nameOpt == nil {
		return AstExprError{
			NodeLoc:      &NodeLoc{snapshot()},
			Expressions:  nil,
			MessageIndex: len(parseErrors),
		}
	}

	name := nameOpt
	local_ := localMap[name.Name.Value]

	if local_ != nil {
		if local_.FunctionDepth < typeFunctionDepth {
			return reportExprError(snapshot(), nil, fmt.Sprintf("Type function cannot reference outer local '%s'", local_.Name))
		}

		return AstExprLocal{
			NodeLoc: &NodeLoc{name.Location},
			Local:   *local_,
			Upvalue: local_.FunctionDepth != len(functionStack)-1,
		}
	}

	return AstExprGlobal{
		NodeLoc: &NodeLoc{name.Location},
		Name:    name.Name.Value,
	}
}

// parsePrefixExpr parses NAME | `(' expr `)'
func parsePrefixExpr() AstExpr {
	if token_type == '(' {
		start := token_location.Begin
		parenType := token_type
		parenBegin := token_location.Begin
		nextLexeme()

		expr := parseExpr(0)

		end := token_location.End
		if token_type != ')' {
			var extra string
			if token_type == '=' {
				extra = "; did you mean to use '{' when defining a table?"
			}
			expectMatchAndConsumeFail(')', parenType, parenBegin, extra)
			end = prev_location.End
		} else {
			end = token_location.End
			nextLexeme()
		}

		return AstExprGroup{
			NodeLoc: &NodeLoc{lex.Location{Begin: start, End: end}},
			Expr:    expr,
		}
	}

	return parseNameExpr("expression")
}

// parseTypeInstantiationExpr parses `<<' type params `>>'
func parseTypeInstantiationExpr() ([]AstTypeOrPack, CstTypeInstantiation) {
	leftArrow1 := token_location.Begin
	beginType := token_type
	beginPos := token_location.Begin
	nextLexeme()

	var leftArrow2 lex.Position
	var commaPositions []lex.Position
	var rightArrow1 lex.Position

	typesOrPacks := parseTypeParams(&leftArrow2, &commaPositions, &rightArrow1)

	rightArrow2 := token_location.Begin
	expectMatchAndConsume('>', beginType, beginPos, nil)

	cstData := CstTypeInstantiation{
		LeftArrow1Position:  leftArrow1,
		LeftArrow2Position:  leftArrow2,
		CommaPositions:      commaPositions,
		RightArrow1Position: rightArrow1,
		RightArrow2Position: rightArrow2,
	}

	return typesOrPacks, cstData
}

// parseExplicitTypeInstantiationExpr parses expr `<<' TypeParams `>>'
func parseExplicitTypeInstantiationExpr(start lex.Position, basedOnExpr AstExpr) AstExprInstantiate {
	typesOrPacks, cstInstantiation := parseTypeInstantiationExpr()

	expr := AstExprInstantiate{
		NodeLoc:       &NodeLoc{lex.Location{Begin: start, End: prev_location.End}},
		Expr:          basedOnExpr,
		TypeArguments: typesOrPacks,
	}

	if storeCstData {
		cstNodes[expr] = CstExprExplicitTypeInstantiation{
			Instantiation: cstInstantiation,
		}
	}

	return expr
}

func reportAmbiguousCallError() {
	report(snapshot(), "Ambiguous syntax: this looks like an argument list for a function call, but could also be a start of new statement; use ';' to separate statements")
}

// parsePrimaryExpr parses primary expression (field access, indexing, calls)
func parsePrimaryExpr(asStatement bool) AstExpr {
	start := token_location.Begin
	expr := AstExpr(parsePrefixExpr())

	oldRecursion := recursionCounter

	for {
		if token_type == '.' {
			opPosition := token_location.Begin
			nextLexeme()

			ctx := "field name"
			index := parseIndexName(&ctx, opPosition)

			expr = AstExprIndexName{
				NodeLoc:       &NodeLoc{lex.Location{Begin: start, End: index.Location.End}},
				Expr:          expr,
				Index:         index.Name.Value,
				IndexLocation: index.Location,
				OpPosition:    opPosition,
				Op:            '.',
			}
		} else if token_type == '[' {
			bracketType := token_type
			bracketBegin := token_location.Begin
			openBracket := token_location.Begin
			nextLexeme()

			index := parseExpr(0)
			closeBracket := token_location.Begin
			expectMatchAndConsume(']', bracketType, bracketBegin, nil)

			e := AstExprIndexExpr{
				NodeLoc: &NodeLoc{lex.Location{Begin: start, End: prev_location.End}},
				Expr:    expr,
				Index:   index,
			}

			if storeCstData {
				cstNodes[e] = CstExprIndexExpr{
					OpenBracketPosition:  openBracket,
					CloseBracketPosition: closeBracket,
				}
			}

			expr = e
		} else if token_type == ':' {
			opPosition := token_location.Begin
			nextLexeme()

			ctx := "method name"
			index := parseIndexName(&ctx, opPosition)

			funcExpr := AstExprIndexName{
				NodeLoc:       &NodeLoc{lex.Location{Begin: start, End: index.Location.End}},
				Expr:          expr,
				Index:         index.Name.Value,
				IndexLocation: index.Location,
				OpPosition:    opPosition,
				Op:            ':',
			}

			if LuauExplicitTypeInstantiationSyntax {
				var typeArgs []AstTypeOrPack
				var cstInstantiation *CstTypeInstantiation

				if token_type == '<' && next_type == '<' {
					args, cst := parseTypeInstantiationExpr()
					typeArgs = args
					cstInstantiation = &cst
				}

				callExpr := parseFunctionArgs(AstExpr(funcExpr), true)
				if len(typeArgs) > 0 {
					if ce, ok := callExpr.(AstExprCall); ok {
						ce.TypeArguments = &typeArgs
						callExpr = ce
					}
				}
				if storeCstData && cstInstantiation != nil {
					if ce, ok := callExpr.(AstExprCall); ok {
						if cstCall, ok2 := cstNodes[ce].(CstExprCall); ok2 {
							cstCall.ExplicitTypes = cstInstantiation
							cstNodes[ce] = cstCall
						}
					}
				}
				expr = callExpr
			} else {
				expr = parseFunctionArgs(AstExpr(funcExpr), true)
			}
		} else if token_type == '(' {
			if !asStatement && expr.GetLocation().End.Line != token_location.Begin.Line {
				reportAmbiguousCallError()
				break
			}
			expr = parseFunctionArgs(expr, false)
		} else if token_type == '{' || token_type == lex.RawString || token_type == lex.QuotedString {
			expr = parseFunctionArgs(expr, false)
		} else if LuauExplicitTypeInstantiationSyntax && token_type == '<' && next_type == '<' {
			expr = parseExplicitTypeInstantiationExpr(start, expr)
		} else {
			break
		}

		incrementRecursionCounter("expression")
	}

	recursionCounter = oldRecursion
	return expr
}

// parseAssertionExpr parses expr [`::' Type]
func parseAssertionExpr() AstExpr {
	start := snapshot()
	expr := parseSimpleExpr()

	if token_type == lex.DoubleColon {
		opPos := token_location.Begin
		nextLexeme()
		annotation := parseType(false)
		annotLoc := annotation.GetLocation()

		node := AstExprTypeAssertion{
			NodeLoc:    &NodeLoc{lex.Location{Begin: start.Begin, End: annotLoc.End}},
			Expr:       expr,
			Annotation: annotation,
		}

		if storeCstData {
			cstNodes[node] = CstExprTypeAssertion{OpPosition: opPos}
		}

		return node
	}

	return expr
}

// parseSimpleExpr parses atoms: literals, `...', constructor, function, primary
func parseSimpleExpr() AstExpr {
	start := snapshot()

	var attributes Attrs
	if token_type == lex.Attribute || token_type == lex.AttributeOpen {
		attributes = parseAttributes()

		if token_type != lex.ReservedFunction {
			currLex := lex.Lexeme{Type: token_type, Codepoint: token_codepoint}
			if token_string != nil {
				currLex.Data = []byte(*token_string)
			}
			return reportExprError(start, nil, fmt.Sprintf("Expected 'function' declaration after attribute, but got %s instead", currLex.String()))
		}
	}

	switch token_type {
	case lex.ReservedNil:
		nextLexeme()
		return AstExprConstantNil{NodeLoc: &NodeLoc{start}}
	case lex.ReservedTrue:
		nextLexeme()
		return AstExprConstantBool{NodeLoc: &NodeLoc{start}, Value: true}
	case lex.ReservedFalse:
		nextLexeme()
		return AstExprConstantBool{NodeLoc: &NodeLoc{start}, Value: false}
	case lex.ReservedFunction:
		matchFunction := get_lexeme()
		nextLexeme()
		node, _ := parseFunctionBody(false, matchFunction, nil, nil, attributes)
		return node
	case lex.Number:
		return parseNumber()
	case lex.RawString, lex.QuotedString, lex.InterpStringSimple:
		return parseString()
	case lex.InterpStringBegin:
		return parseInterpString()
	case lex.BrokenString:
		nextLexeme()
		return reportExprError(start, nil, "Malformed string; did you forget to finish it?")
	case lex.BrokenInterpDoubleBrace:
		nextLexeme()
		return reportExprError(start, nil, "Double braces are not permitted within interpolated strings; did you mean '\\{'?")
	case lex.Dot3:
		if len(functionStack) > 0 && functionStack[len(functionStack)-1].Vararg {
			nextLexeme()
			return AstExprVarargs{NodeLoc: &NodeLoc{start}}
		}
		nextLexeme()
		return reportExprError(start, nil, "Cannot use '...' outside of a vararg function")
	case '{':
		return parseTableConstructor()
	case lex.ReservedIf:
		return parseIfElseExpr()
	default:
		return parsePrimaryExpr(false)
	}
}

// parseFunctionArgs parses `(' [explist] `)' | tableconstructor | String
func parseFunctionArgs(funcExpr AstExpr, selfCall bool) AstExpr {
	if token_type == '(' {
		if funcExpr.GetLocation().End.Line != token_location.Begin.Line {
			reportAmbiguousCallError()
		}

		argStart := token_location.End
		parenType := token_type
		parenBegin := token_location.Begin
		nextLexeme()

		var args []AstExpr
		var commaPositions []lex.Position
		if token_type != ')' {
			parseExprList(&args, &commaPositions)
		}

		closeParen := token_location.Begin
		end := snapshot()
		expectMatchAndConsume(')', parenType, parenBegin, nil)

		result := AstExprCall{
			NodeLoc:     &NodeLoc{lex.Location{Begin: funcExpr.GetLocation().Begin, End: end.End}},
			Func:        funcExpr,
			Args:        args,
			Self:        selfCall,
			ArgLocation: lex.Location{Begin: argStart, End: end.End},
		}

		if storeCstData {
			cstNodes[result] = CstExprCall{
				OpenParens:     &parenBegin,
				CloseParens:    &closeParen,
				CommaPositions: commaPositions,
			}
		}

		return result
	} else if token_type == '{' {
		argStart := token_location.End
		tableExpr := parseTableConstructor()
		argEnd := prev_location.End

		result := AstExprCall{
			NodeLoc:     &NodeLoc{lex.Location{Begin: funcExpr.GetLocation().Begin, End: tableExpr.GetLocation().End}},
			Func:        funcExpr,
			Args:        []AstExpr{tableExpr},
			Self:        selfCall,
			ArgLocation: lex.Location{Begin: argStart, End: argEnd},
		}

		if storeCstData {
			cstNodes[result] = CstExprCall{CommaPositions: []lex.Position{}}
		}

		return result
	} else if token_type == lex.RawString || token_type == lex.QuotedString {
		argLocation := snapshot()
		strExpr := parseString()

		result := AstExprCall{
			NodeLoc:     &NodeLoc{lex.Location{Begin: funcExpr.GetLocation().Begin, End: strExpr.GetLocation().End}},
			Func:        funcExpr,
			Args:        []AstExpr{strExpr},
			Self:        selfCall,
			ArgLocation: argLocation,
		}

		if storeCstData {
			cstNodes[result] = CstExprCall{CommaPositions: []lex.Position{}}
		}

		return result
	}

	return reportFunctionArgsError(funcExpr, selfCall)
}

// reportFunctionArgsError reports error for bad function call syntax
func reportFunctionArgsError(funcExpr AstExpr, selfCall bool) AstExpr {
	if selfCall && token_location.Begin.Line != funcExpr.GetLocation().End.Line {
		return reportExprError(funcExpr.GetLocation(), []AstExpr{funcExpr}, "Expected function call arguments after '('")
	}

	currLex := lex.Lexeme{Type: token_type, Codepoint: token_codepoint}
	if token_string != nil {
		currLex.Data = []byte(*token_string)
	}
	return reportExprError(
		lex.Location{Begin: funcExpr.GetLocation().Begin, End: token_location.Begin},
		[]AstExpr{funcExpr},
		fmt.Sprintf("Expected '(', '{' or <string> when parsing function call, got %s", currLex.String()),
	)
}

// parseIndexName parses a field name, accepting keywords if on same line
func parseIndexName(context *string, prev lex.Position) Binding {
	nameOpt := parseNameOpt(context)
	if nameOpt != nil {
		return *nameOpt
	}

	if token_type >= lex.Reserved_BEGIN && token_type < lex.Reserved_END &&
		token_location.Begin.Line == prev.Line {
		nameStr := ""
		if token_string != nil {
			nameStr = *token_string
		}
		result := Binding{
			Name:    lex.AstName{Value: nameStr},
			NodeLoc: &NodeLoc{snapshot()},
		}

		nextLexeme()
		return result
	}

	return Binding{
		Name:    lex.AstName{Value: nameError},
		NodeLoc: &NodeLoc{snapshot()},
	}
}

// parseCallList parses function call arguments (used by intepstring etc.)
func parseCallList(commaPositions *[]lex.Position) ([]AstExpr, lex.Location, lex.Location) {
	if token_type == '(' {
		argStart := token_location.End
		parenType := token_type
		parenBegin := token_location.Begin

		nextLexeme()

		var args []AstExpr

		if token_type != ')' {
			parseExprList(&args, commaPositions)
		}

		end := snapshot()
		expectMatchAndConsume(')', parenType, parenBegin, nil)

		return args,
			lex.Location{Begin: argStart, End: end.End},
			lex.Location{Begin: parenBegin, End: prev_location.End}
	} else if token_type == '{' {
		argStart := token_location.End
		expr := parseTableConstructor()

		return []AstExpr{expr},
			lex.Location{Begin: argStart, End: prev_location.End},
			expr.GetLocation()
	}

	argLoc := snapshot()
	expr := parseString()
	return []AstExpr{expr}, argLoc, expr.GetLocation()
}

// parseTableConstructor parses `{' [fieldlist] `}'
func parseTableConstructor() AstExprTable {
	var items []AstExprTableItem
	var cstItems []CstExprTableItem

	start := snapshot()

	braceType := token_type
	braceBegin := token_location.Begin
	context := "table literal"
	expectAndConsume('{', &context)

	lastElementIndent := uint32(0)

	for token_type != '}' {
		lastElementIndent = token_location.Begin.Column

		if token_type == '[' {
			indexerOpenPos := token_location.Begin
			bracketType := token_type
			bracketBegin := token_location.Begin
			nextLexeme()

			key := parseExpr(0)

			indexerClosePos := token_location.Begin
			expectMatchAndConsume(']', bracketType, bracketBegin, nil)

			equalsPos := token_location.Begin
			ctx := "table field"
			expectAndConsume('=', &ctx)

			value := parseExpr(0)

			items = append(items, AstExprTableItem{
				NodeLoc: &NodeLoc{lex.Location{}},
				Kind:    "General",
				Key:     &key,
				Value:   value,
			})

			if storeCstData {
				sepPos := token_location.Begin
				cstItems = append(cstItems, CstExprTableItem{
					Kind:                 "General",
					IndexerOpenPosition:  &indexerOpenPos,
					IndexerClosePosition: &indexerClosePos,
					EqualsPosition:       &equalsPos,
					Separator:            tableSeparator(),
					SeparatorPosition:    sepPos,
				})
			}
		} else if token_type == lex.Name && next_type == '=' {
			ctx := "table field"
			name := parseName(&ctx)

			equalsPos := token_location.Begin
			ctx2 := "table field"
			expectAndConsume('=', &ctx2)

			keyExpr := AstExpr(AstExprConstantString{
				NodeLoc: &NodeLoc{name.Location},
				Value:   name.Name.Value,
			})

			value := parseExpr(0)

			if fe, ok := value.(AstExprFunction); ok {
				fe.Debugname = name.Name.Value
				value = fe
			}

			items = append(items, AstExprTableItem{
				NodeLoc: &NodeLoc{lex.Location{}},
				Kind:    "Record",
				Key:     &keyExpr,
				Value:   value,
			})

			if storeCstData {
				sepPos := token_location.Begin
				cstItems = append(cstItems, CstExprTableItem{
					Kind:              "Record",
					EqualsPosition:    &equalsPos,
					Separator:         tableSeparator(),
					SeparatorPosition: sepPos,
				})
			}
		} else {
			expr := parseExpr(0)
			items = append(items, AstExprTableItem{
				NodeLoc: &NodeLoc{lex.Location{}},
				Kind:    "List",
				Value:   expr,
			})

			if storeCstData {
				sepPos := token_location.Begin
				cstItems = append(cstItems, CstExprTableItem{
					Kind:              "List",
					Separator:         tableSeparator(),
					SeparatorPosition: sepPos,
				})
			}
		}

		if token_type == ',' || token_type == ';' {
			nextLexeme()
		} else if (token_type == '[' || token_type == lex.Name) && token_location.Begin.Column == lastElementIndent {
			report(snapshot(), "Expected ',' after table constructor element")
		} else if token_type != '}' {
			break
		}
	}

	end := snapshot()
	if !expectMatchAndConsume('}', braceType, braceBegin, nil) {
		end = getprev()
	}

	node := AstExprTable{
		NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: end.End}},
		Items:   items,
	}

	if storeCstData {
		cstNodes[node] = CstExprTable{Items: cstItems}
	}

	return node
}

// parseIfElseExpr parses if-then-else expression
func parseIfElseExpr() AstExprIfElse {
	start := snapshot()
	nextLexeme() // consume 'if' or 'elseif'

	condition := parseExpr(0)

	thenPosition := token_location.Begin
	hasThen := expectAndConsume(lex.ReservedThen, nil)

	trueExpr := parseExpr(0)
	var falseExpr AstExpr

	elsePosition := token_location.Begin
	isElseIf := false
	hasElse := false

	if token_type == lex.ReservedElseif {
		oldRecursion := recursionCounter
		incrementRecursionCounter("expression")
		hasElse = true
		result := parseIfElseExpr()
		falseExpr = result
		recursionCounter = oldRecursion
		isElseIf = true
	} else {
		hasElse = expectAndConsume(lex.ReservedElse, nil)
		falseExpr = parseExpr(0)
	}

	var falseEnd lex.Position
	if falseExpr != nil {
		falseEnd = falseExpr.GetLocation().End
	}

	node := AstExprIfElse{
		NodeLoc:   &NodeLoc{lex.Location{Begin: start.Begin, End: falseEnd}},
		Condition: condition,
		HasThen:   hasThen,
		TrueExpr:  trueExpr,
		HasElse:   hasElse,
		FalseExpr: falseExpr,
	}

	if storeCstData {
		cstNodes[node] = CstExprIfElse{
			ThenPosition: thenPosition,
			ElsePosition: elsePosition,
			IsElseIf:     isElseIf,
		}
	}

	return node
}

// parseInterpString parses an interpolated string expression
func parseInterpString() AstExprInterpStringOrError {
	var strs []string
	var sourceStrings []string
	var stringPositions []lex.Position
	var expressions []AstExpr

	startLocation := snapshot()
	var endLocation lex.Location

	for {
		currentLexeme := get_lexeme()
		endLocation = currentLexeme.Location

		data := ""
		if token_string != nil {
			data = *token_string
		}

		if storeCstData {
			sourceStrings = append(sourceStrings, data)
			stringPositions = append(stringPositions, currentLexeme.Location.Begin)
		}

		ok, fixedData := lexer.FixupQuotedString([]byte(data))
		if !ok {
			nextLexeme()
			return reportExprError(
				lex.Location{Begin: startLocation.Begin, End: endLocation.End},
				nil,
				"Interpolated string literal contains malformed escape sequence",
			)
		}

		nextLexeme()
		strs = append(strs, string(fixedData))

		if currentLexeme.Type == lex.InterpStringEnd || currentLexeme.Type == lex.InterpStringSimple {
			break
		}

		t := token_type

		if t == lex.InterpStringMid || t == lex.InterpStringEnd {
			nextLexeme()
			expressions = append(expressions, reportExprError(endLocation, nil, "Malformed interpolated string, expected expression inside '{}'"))
			break
		} else if t == lex.BrokenString {
			nextLexeme()
			expressions = append(expressions, reportExprError(endLocation, nil, "Malformed interpolated string; did you forget to add a '`'?"))
			break
		} else {
			expressions = append(expressions, parseExpr(0))
		}

		t = token_type

		if t == lex.InterpStringBegin || t == lex.InterpStringMid || t == lex.InterpStringEnd {
			// continue reading
		} else if t == lex.BrokenInterpDoubleBrace {
			nextLexeme()
			return reportExprError(endLocation, nil, "Double braces are not permitted within interpolated strings; did you mean '\\{'?")
		} else if t == lex.BrokenString || t == lex.Eof {
			if t == lex.BrokenString {
				nextLexeme()
			}

			node := AstExprInterpString{
				NodeLoc:     &NodeLoc{lex.Location{Begin: startLocation.Begin, End: prev_location.End}},
				Strings:     strs,
				Expressions: expressions,
			}

			if storeCstData {
				cstNodes[node] = CstExprInterpString{
					SourceStrings:   sourceStrings,
					StringPositions: stringPositions,
				}
			}

			if len(braceStack) > 0 && braceStack[len(braceStack)-1] == lex.InterpolatedString {
				report(getprev(), "Malformed interpolated string; did you forget to add a '}'?")
			} else {
				report(getprev(), "Malformed interpolated string; did you forget to add a '`'?")
			}

			return node
		} else {
			currLex := lex.Lexeme{Type: token_type, Codepoint: token_codepoint}
			if token_string != nil {
				currLex.Data = []byte(*token_string)
			}
			return reportExprError(endLocation, nil, fmt.Sprintf("Malformed interpolated string, got %s", currLex.String()))
		}
	}

	node := AstExprInterpString{
		NodeLoc:     &NodeLoc{lex.Location{Begin: startLocation.Begin, End: endLocation.End}},
		Strings:     strs,
		Expressions: expressions,
	}

	if storeCstData {
		cstNodes[node] = CstExprInterpString{
			SourceStrings:   sourceStrings,
			StringPositions: stringPositions,
		}
	}

	return node
}

// parseCharArray parses string token and returns unescaped bytes, or nil on error
func parseCharArray() *string {
	t := token_type
	data := ""
	if token_string != nil {
		data = *token_string
	}

	var result string

	if t == lex.QuotedString || t == lex.InterpStringSimple {
		ok, fixed := lexer.FixupQuotedString([]byte(data))
		if !ok {
			nextLexeme()
			return nil
		}
		result = string(fixed)
	} else {
		result = string(lexer.FixupMultilineString([]byte(data)))
	}

	nextLexeme()
	return &result
}

// parseString parses a string literal expression
func parseString() AstExprConstantStringOrError {
	location := snapshot()

	var fullStyle CstQuotes
	var blockDepth int
	if storeCstData {
		fullStyle, blockDepth = extractStringDetails()
	}

	var originalString *string
	if storeCstData {
		originalString = token_string
	}

	value := parseCharArray()

	if value != nil {
		node := AstExprConstantString{
			NodeLoc: &NodeLoc{location},
			Value:   *value,
		}

		if storeCstData {
			cstNodes[node] = CstExprConstantString{
				SourceString: originalString,
				QuoteStyle:   int(fullStyle),
				BlockDepth:   blockDepth,
			}
		}

		return node
	}

	return reportExprError(location, nil, "String literal contains malformed escape sequence")
}

// parseNumber parses a number literal expression
func parseNumber() AstExprConstantNumberOrError {
	start := snapshot()
	data := ""
	if token_string != nil {
		data = *token_string
	}

	var sourceData string
	if storeCstData {
		sourceData = data
	}

	cleanData := strings.ReplaceAll(data, "_", "")

	value := 0.0
	malformed := false

	if strings.HasPrefix(cleanData, "0x") || strings.HasPrefix(cleanData, "0X") {
		v, err := strconv.ParseUint(cleanData[2:], 16, 64)
		if err != nil {
			malformed = true
		} else {
			value = float64(v)
		}
	} else if strings.HasPrefix(cleanData, "0b") || strings.HasPrefix(cleanData, "0B") {
		v, err := strconv.ParseUint(cleanData[2:], 2, 64)
		if err != nil {
			malformed = true
		} else {
			value = float64(v)
		}
	} else {
		v, err := strconv.ParseFloat(cleanData, 64)
		if err != nil {
			malformed = true
		} else {
			value = v
		}
	}

	nextLexeme()

	if malformed {
		return reportExprError(start, nil, "Malformed number")
	}

	node := AstExprConstantNumber{
		NodeLoc: &NodeLoc{start},
		Value:   value,
	}

	if storeCstData {
		cstNodes[node] = CstExprConstantNumber{Value: sourceData}
	}

	return node
}
