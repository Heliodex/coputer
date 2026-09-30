package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Heliodex/coputer/ast/lex"
)

// These globals aren't that great to have around for now, though they'll stick here until compliance with the reference implementation is ensured.

// Parser Settings

const (
	LuauExplicitTypeInstantiationSyntax = true
	DesugaredArrayTypeReferenceIsEmpty  = true
	// LuauCstStatDoWithStatsStart         = true
)

const (
	TypeLengthLimit = 1000
	RecursionLimit  = 1000
	ErrorLimit      = 100
)

type QuoteStyle uint8

const (
	QuoteStyle_QuotedSimple QuoteStyle = iota
	QuoteStyle_QuotedSingle
	QuoteStyle_QuotedRaw
	QuoteStyle_Unquoted
)

func (q QuoteStyle) String() string {
	switch q {
	case QuoteStyle_QuotedSimple:
		return "QuotedSimple"
	case QuoteStyle_QuotedSingle:
		return "QuotedSingle"
	case QuoteStyle_QuotedRaw:
		return "QuotedRaw"
	case QuoteStyle_Unquoted:
		return "Unquoted"
	}
	return "<unknown>"
}

type CstQuotes uint8

const (
	CstQuotes_Single CstQuotes = iota
	CstQuotes_Double
	CstQuotes_Raw
	CstQuotes_Interp
)

func (c CstQuotes) String() string {
	switch c {
	case CstQuotes_Single:
		return "Single"
	case CstQuotes_Double:
		return "Double"
	case CstQuotes_Raw:
		return "Raw"
	case CstQuotes_Interp:
		return "Interp"
	}
	return "<unknown>"
}

type UnaryOp uint8

const (
	UnaryOp_Not UnaryOp = iota
	UnaryOp_Minus
	UnaryOp_Len
)

func (u UnaryOp) String() string {
	switch u {
	case UnaryOp_Not:
		return "Not"
	case UnaryOp_Minus:
		return "Minus"
	case UnaryOp_Len:
		return "Len"
	}
	return "<unknown>"
}

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

func (b BinaryOp) String() string {
	switch b {
	case BinaryOp_Add:
		return "Add"
	case BinaryOp_Sub:
		return "Sub"
	case BinaryOp_Mul:
		return "Mul"
	case BinaryOp_Div:
		return "Div"
	case BinaryOp_FloorDiv:
		return "FloorDiv"
	case BinaryOp_Mod:
		return "Mod"
	case BinaryOp_Pow:
		return "Pow"
	case BinaryOp_Concat:
		return "Concat"
	case BinaryOp_CompareNe:
		return "CompareNe"
	case BinaryOp_CompareEq:
		return "CompareEq"
	case BinaryOp_CompareLt:
		return "CompareLt"
	case BinaryOp_CompareLe:
		return "CompareLe"
	case BinaryOp_CompareGt:
		return "CompareGt"
	case BinaryOp_CompareGe:
		return "CompareGe"
	case BinaryOp_And:
		return "And"
	case BinaryOp_Or:
		return "Or"
	}
	return "<unknown>"
}

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
	case AstExprConstantNil, AstExprConstantBool, AstExprConstantNumber, AstExprConstantInteger, AstExprConstantString:
		return true
	}
	return false
}

func ExprLValues(expr AstExpr) bool {
	switch e := expr.(type) {
	case AstExprLocal:
		// Constant locals may not be reassigned
		return !e.Local.IsConst
	case AstExprGlobal, AstExprIndexExpr, AstExprIndexName:
		return true
	}
	return false
}

// reportLValueError reports an error for an expression that cannot be assigned to,
// distinguishing constant locals from other non-lvalue expressions.
func (p *Parser) reportLValueError(expr AstExpr) *AstExprError {
	if e, ok := expr.(AstExprLocal); ok && e.Local.IsConst {
		return p.reportExprError(expr.GetLocation(), []AstExpr{expr}, fmt.Sprintf("Variable '%s' is constant and may not be reassigned", e.Local.Name))
	}
	return p.reportExprError(expr.GetLocation(), []AstExpr{expr}, "Assigned expression must be a variable or a field")
}

// isEnoughValues reports whether an expression list definitely provides enough
// values for the expected number of bindings. A trailing call or `...` may
// expand to any number of values.
func isEnoughValues(values []AstExpr, expected int) bool {
	if len(values) > 0 {
		switch values[len(values)-1].(type) {
		case AstExprCall, AstExprVarargs:
			return true
		}
	}
	return len(values) == expected
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
		if item.Kind == General {
			return false
		}
		if item.Kind == Record || item.Kind == List {
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
		if item.Key != nil && item.Kind == Record {
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

// All mutable parser state lives on *Parser (see parser.go); the remaining
// package-level vars below are immutable lookup tables and limits.

// All unlocalized Parser functions

// there would be 52 declarations here if Go supported forward declarations
// honestly they're still better than function hoisting

// Parser Interface

func (p *Parser) fillNext() {
	for {
		next := p.lexer.Next0()

		// fmt.Println("lexed next type", lex.Lexeme{Type: next.Type}.String())

		p.next_type = next.Type
		p.next_location = next.Location
		p.next_codepoint = next.Codepoint
		nstr := string(next.Data)
		p.next_string = &nstr
		p.next_aux = next.Aux

		if next.Type == lex.Comment || next.Type == lex.BlockComment || next.Type == lex.BrokenComment {
			// comments are always collected: they're attached to the blocks
			// they appear in so that Source() can reproduce them
			comment := Comment{
				Type:    next.Type,
				Content: nstr,
				NodeLoc: &NodeLoc{next.Location},
			}
			p.commentLocations = append(p.commentLocations, comment)
			p.pendingComments = append(p.pendingComments, comment)

			if next.Type == lex.Comment && p.next_string != nil && (*p.next_string)[0] == '!' {
				p.hotcomments = append(p.hotcomments, HotComment{
					Header:   p.hotcommentHeader,
					Location: next.Location,
					Content:  *p.next_string,
				})
			}

			if next.Type == lex.BrokenComment {
				return
			}

			continue
		}

		break
	}

	// fmt.Println("filled next with type", lex.Lexeme{Type: next_type}.String())
}

func (p *Parser) nextLexeme() {
	// Save previous current to prev
	p.prev_location = p.token_location

	// Move NEXT to CURRENT
	p.token_type = p.next_type
	// fmt.Println("set token_type to", lex.Lexeme{Type: token_type}.String())
	p.token_location = p.next_location
	// if next_string != nil {
	p.token_string = p.next_string
	// }
	// fmt.Println("set token_string to", *token_string)
	p.token_aux = p.next_aux
	p.token_codepoint = p.next_codepoint

	// Refill NEXT
	p.fillNext()
}

// Parser Commons

func (p *Parser) snapshot() lex.Location {
	return p.token_location
}

func (p *Parser) get_lexeme() lex.Lexeme {
	return lex.Lexeme{
		Type:     p.token_type,
		Location: p.token_location,
	}
}

func (p *Parser) getprev() lex.Location {
	return p.prev_location
}

// Error reports

func (p *Parser) report(loc lex.Location, msg string) {
	if len(p.parseErrors) > 0 && p.parseErrors[len(p.parseErrors)-1].Location == loc {
		return
	}

	p.parseErrors = append(p.parseErrors, ParseError{Location: loc, Message: msg})

	if ErrorLimit == 1 {
		panic(msg)
	}

	if len(p.parseErrors) >= ErrorLimit {
		panic(fmt.Sprintf("Reached error limit (%d)", ErrorLimit))
	}
}

func (p *Parser) expectAndConsumeFail(type_ lex.LexemeType, context *string) {
	typeString := lex.Lexeme{Type: type_}.String()

	lexLex := lex.Lexeme{Codepoint: p.token_codepoint}
	if p.token_string != nil {
		lexLex.Data = []byte(*p.token_string)
	}
	lexString := lexLex.String()

	if context != nil {
		p.report(p.snapshot(), fmt.Sprintf("Expected %s when parsing %s, got %s", typeString, *context, lexString))
	} else {
		p.report(p.snapshot(), fmt.Sprintf("Expected %s, got %s", typeString, lexString))
	}
}

func (p *Parser) expectMatchAndConsumeFail(type_, begin_type lex.LexemeType, position lex.Position, extra ...string) {
	typeString := lex.Lexeme{Type: type_}.String()
	matchString := lex.Lexeme{Type: begin_type}.String()
	currLex := lex.Lexeme{Type: p.token_type, Codepoint: p.token_codepoint}
	if p.token_string != nil {
		currLex.Data = []byte(*p.token_string)
	}
	currString := currLex.String()

	var xtra string
	if len(extra) > 0 {
		xtra = extra[0]
	}

	if p.token_location.Begin.Line == position.Line {
		p.report(p.snapshot(), fmt.Sprintf("Expected %s (to close %s at column) %d, got %s%s", typeString, matchString, position.Column+1, currString, xtra))
	} else {
		p.report(p.snapshot(), fmt.Sprintf("Expected %s (to close %s at line) %d, got %s%s", typeString, matchString, position.Line, currString, xtra))
	}
}

func (p *Parser) expectAndConsume(type_ lex.LexemeType, context *string) bool {
	if p.token_type != type_ {
		p.expectAndConsumeFail(type_, context)

		if p.next_type == type_ {
			p.nextLexeme()
			p.nextLexeme()
		}

		return false
	}

	p.nextLexeme()
	return true
}

func (p *Parser) expectMatchAndConsume(value, begin_type lex.LexemeType, position lex.Position, seachForMissing *bool) bool {
	if p.token_type != value {
		p.expectMatchAndConsumeFail(value, begin_type, position)

		if seachForMissing != nil && (*seachForMissing) {
			currentLine := p.prev_location.End.Line
			type_ := p.token_type

			for currentLine == p.token_location.Begin.Line && type_ != value && p.matchRecovery[type_] == 0 {
				p.nextLexeme()
				type_ = p.token_type
			}

			if type_ == value {
				p.nextLexeme()
				return true
			}
		} else {
			if p.next_type == value {
				p.nextLexeme()
				p.nextLexeme()
				return true
			}
		}

		return false
	}

	p.nextLexeme()
	return true
}

func (p *Parser) expectMatchEndAndConsume(type_, begin_type lex.LexemeType, position lex.Position) bool {
	if p.token_type != type_ {
		if p.suspect_type != lex.Eof && p.suspect_line > position.Line {
			suggestionLex := lex.Lexeme{Type: p.suspect_type, Codepoint: p.next_codepoint}
			if p.token_string != nil {
				suggestionLex.Data = []byte(*p.token_string)
			}
			suggestionString := suggestionLex.String()

			suggestion := fmt.Sprintf("; did you forget to close %s at line %d?", suggestionString, position.Line+1)

			p.expectMatchAndConsumeFail(type_, begin_type, position, suggestion)
		} else {
			p.expectMatchAndConsumeFail(type_, begin_type, position)
		}

		if p.next_type == type_ {
			p.nextLexeme()
			p.nextLexeme()
			return true
		}

		return false
	}

	if p.token_location.Begin.Line != position.Line && p.token_location.Begin.Column != position.Column && p.suspect_line < position.Line {
		p.suspect_line = position.Line
		p.suspect_type = begin_type
	}

	p.nextLexeme()
	return true
}

// Ast reports

func (p *Parser) reportStatError(location lex.Location, exprs []AstExpr, stats []AstStat, msg string) *AstStatError {
	p.report(location, msg)

	return &AstStatError{
		NodeLoc:      &NodeLoc{location},
		Expressions:  exprs,
		Statements:   stats,
		MessageIndex: len(p.parseErrors) - 1,
	}
}

func (p *Parser) reportExprError(location lex.Location, exprs []AstExpr, msg string) *AstExprError {
	p.report(location, msg)

	return &AstExprError{
		NodeLoc:      &NodeLoc{location},
		Expressions:  exprs,
		MessageIndex: len(p.parseErrors) - 1,
	}
}

func (p *Parser) reportTypeError(location lex.Location, types []AstType, msg string) *AstTypeError {
	p.report(location, msg)

	return &AstTypeError{
		NodeLoc:      &NodeLoc{location},
		Types:        types,
		MessageIndex: len(p.parseErrors) - 1,
	}
}

func (p *Parser) reportNameError(context *string) {
	currLex := lex.Lexeme{Type: p.token_type, Codepoint: p.token_codepoint}
	if p.token_string != nil {
		currLex.Data = []byte(*p.token_string)
	}
	currString := currLex.String()

	if context != nil {
		p.report(p.snapshot(), fmt.Sprintf("Expected identifier when parsing %s, got %s", *context, currString))
	} else {
		p.report(p.snapshot(), fmt.Sprintf("Expected identifier, got %s", currString))
	}
}

// Locals helpers

func (p *Parser) restoreLocals(offset int) {
	for i := len(p.localStack) - 1; i >= offset; i-- {
		l := p.localStack[i]
		// l better not be nil bruh
		p.localMap[l.Name] = l.Shadow
	}

	// setting to nil wouldn't change the array length, which I assume we're relying on somewhere...
	p.localStack = p.localStack[:offset]
}

func (p *Parser) pushLocal(binding Binding) *AstLocal {
	name := binding.Name.Value
	shadow := p.localMap[name]

	local := &AstLocal{
		Name:          name,
		NodeLoc:       binding.NodeLoc,
		Shadow:        shadow,
		FunctionDepth: len(p.functionStack) - 1,
		LoopDepth:     p.functionStack[len(p.functionStack)-1].LoopDepth,
		Annotation:    binding.Annotation,
		IsConst:       binding.IsConst,
	}

	p.localMap[name] = local
	p.localStack = append(p.localStack, local)

	return local
}

func (p *Parser) incrementRecursionCounter(context string) {
	p.recursionCounter++

	if p.recursionCounter > RecursionLimit {
		msg := fmt.Sprintf("Exceeded allowed recursion depth; simplify your %s to make the code compile", context)
		p.report(p.snapshot(), msg) // lol y
		panic(msg)
	}
}

// The core of the code

func (p *Parser) parseBinding(isConst bool) Binding {
	nameOpt := p.parseNameOpt(new("variable name"))

	var bindingName Binding
	if nameOpt != nil {
		bindingName = *nameOpt
	} else {
		bindingName = Binding{
			Name:    lex.AstName{Value: nameError},
			NodeLoc: &NodeLoc{p.snapshot()},
		}
	}

	colonPos := p.token_location.Begin
	annotation := p.parseOptionalType()

	return Binding{
		Name:          bindingName.Name,
		NodeLoc:       bindingName.NodeLoc,
		Annotation:    annotation,
		ColonPosition: &colonPos,
		IsConst:       isConst,
	}
}

// bindinglist ::= (binding | `...') [`,' bindinglist]
func (p *Parser) parseBindingList(result *[]Binding, allowDot3 bool, commaPositions *[]lex.Position, initialComma *lex.Position, varargAnnotColonPos *[]*lex.Position, isConst bool) (bool, *lex.Location, AstTypePack) {
	localCommaPositions := []lex.Position{}

	if commaPositions != nil && initialComma != nil {
		localCommaPositions = append(localCommaPositions, *initialComma)
	}

	for {
		if p.token_type == lex.Dot3 && allowDot3 {
			varargLocation := p.snapshot()
			p.nextLexeme()

			var tailAnnotation AstTypePack

			if p.token_type == ':' {
				if varargAnnotColonPos != nil {
					(*varargAnnotColonPos)[0] = &p.token_location.Begin
				}

				p.nextLexeme()
				tailAnnotation = p.parseVariadicArgumentTypePack()
			}

			if commaPositions != nil {
				*commaPositions = append(*commaPositions, localCommaPositions...)
			}

			return true, &varargLocation, tailAnnotation
		}

		*result = append(*result, p.parseBinding(isConst))

		if p.token_type != ',' {
			break
		}

		if commaPositions != nil {
			localCommaPositions = append(localCommaPositions, p.token_location.Begin)
		}

		p.nextLexeme()
	}

	if commaPositions != nil {
		*commaPositions = append(*commaPositions, localCommaPositions...)
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
func (p *Parser) parseStat() AstStat {
	type_ := p.token_type

	switch type_ {
	case lex.ReservedIf:
		return p.parseIf()
	case lex.ReservedWhile:
		return p.parseWhile()
	case lex.ReservedDo:
		return p.parseDo()
	case lex.ReservedFor:
		return p.parseFor()
	case lex.ReservedRepeat:
		return p.parseRepeat()
	case lex.ReservedFunction:
		return p.parseFunctionStat(nil)
	case lex.ReservedLocal:
		return p.parseLocal(p.snapshot(), p.token_location.Begin, nil, false)
	case lex.ReservedReturn:
		return p.parseReturn()
	case lex.ReservedBreak:
		return p.parseBreak()
	case lex.Attribute, lex.AttributeOpen:
		return p.parseAttributeStat()
	}

	start_line := p.token_location.Begin.Line
	start_column := p.token_location.Begin.Column
	expr := p.parsePrimaryExpr(true)

	if e, ok := expr.(AstExprCall); ok {
		return &AstStatExpr{
			NodeLoc: e.NodeLoc,
			Expr:    e,
		}
	}

	if p.token_type == ',' || p.token_type == '=' {
		return p.parseAssignment(expr)
	}

	operator, ok := CompoundLookup[p.token_type]
	if ok {
		return p.parseCompoundAssignment(expr, operator)
	}

	var ident *string
	if e, ok := expr.(AstExprGlobal); ok {
		ident = &e.Name
	} else if e, ok := expr.(AstExprLocal); ok {
		ident = &e.Local.Name
	}

	if ident != nil && *ident == "type" {
		loc := expr.GetLocation()
		return p.parseTypeAlias(loc, false, loc.Begin)
	}

	if ident != nil && *ident == "export" {
		if p.token_type == lex.ReservedLocal || p.token_type == lex.ReservedFunction ||
			(p.token_type == lex.Name && p.token_string != nil && *p.token_string == "const") {
			return p.parseExportValue(expr.GetLocation(), expr.GetLocation().Begin, nil)
		}

		if p.token_type == lex.Name && p.token_string != nil && *p.token_string == "type" {
			typeKeywordPos := p.token_location.Begin
			p.nextLexeme()
			return p.parseTypeAlias(expr.GetLocation(), true, typeKeywordPos)
		}
	}

	if ident != nil && *ident == "continue" {
		return p.parseContinue(expr.GetLocation())
	}

	if ident != nil && *ident == "const" {
		// `const` is a contextual keyword; parsePrimaryExpr has already
		// consumed it, so pass its location through to parseLocal.
		return p.parseLocal(expr.GetLocation(), expr.GetLocation().Begin, nil, true)
	}

	if start_line == p.token_location.Begin.Line && start_column == p.token_location.Begin.Column {
		p.nextLexeme()
	}

	return p.reportStatError(expr.GetLocation(), []AstExpr{expr}, nil, "Incomplete statement: expected assignment or a function call")
}

func (p *Parser) parseBlockNoScope() *AstStatBlock {
	var body []AstStat

	prevPos := p.prev_location.End

	// fmt.Println("Current token type at start of block:", token_type)
	for !BlockFollow[p.token_type] {
		oldRecursion := p.recursionCounter
		p.recursionCounter++

		stat := p.parseStat()

		p.recursionCounter = oldRecursion

		if p.token_type == ';' {
			p.nextLexeme()
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

	// fmt.Println("Parsed block with body:", body)

	block := &AstStatBlock{
		NodeLoc: &NodeLoc{
			lex.Location{
				Begin: prevPos,
				End:   p.token_location.Begin,
			},
		},
		Body:   body,
		HasEnd: false,
	}

	// attach the comments contained in this block; nested blocks are finished
	// first, so comments end up on the deepest block containing them
	p.attachBlockComments(block)

	return block
}

// chunk ::= {stat [`;']} [laststat [`;']]
// block ::= chunk
func (p *Parser) parseBlock() *AstStatBlock {
	localsBegin := len(p.localStack)
	result := p.parseBlockNoScope()
	p.restoreLocals(localsBegin)
	return result
}

// attachBlockComments moves the comments contained in block out of the pending
// list and onto the block. Nested blocks are finalized before their parents, so
// every comment ends up attached to the deepest block that contains it.
func (p *Parser) attachBlockComments(block *AstStatBlock) {
	if len(p.pendingComments) == 0 {
		return
	}

	loc := block.GetLocation()

	remaining := p.pendingComments[:0]
	for _, comment := range p.pendingComments {
		if loc.Contains(comment.Location) {
			block.Comments = append(block.Comments, comment)
		} else {
			remaining = append(remaining, comment)
		}
	}
	p.pendingComments = remaining
}

// if exp then block {elseif exp then block} [else block] end
func (p *Parser) parseIf() *AstStatIf {
	start := p.snapshot()

	p.nextLexeme()

	if p.token_type == lex.ReservedLocal {
		return p.parseIfLocalCondition(start)
	}

	if p.token_type == lex.Name && p.token_string != nil && *p.token_string == "const" && p.next_type == lex.Name {
		return p.parseIfLocalCondition(start)
	}

	cond := p.parseExpr(0)

	return p.parseIfTail(start, cond, nil, nil, false, nil, nil)
}

// parseIfLocalCondition parses `if local name = exp then ... end` and
// `if const name = exp then ... end` (LuauExperimentalIfLocalSyntax).
func (p *Parser) parseIfLocalCondition(start lex.Location) *AstStatIf {
	condIsConst := p.token_type == lex.Name && p.token_string != nil && *p.token_string == "const"

	keywordLocation := p.snapshot()
	p.nextLexeme() // consume 'local' or 'const'

	binding := p.parseBinding(condIsConst)

	if p.token_type == ',' {
		p.report(p.token_location, "Expected '=' after variable name in 'if local', got ','; only a single binding is allowed")
	}

	var equalsPosition *lex.Location
	if p.token_type == '=' {
		loc := p.snapshot()
		equalsPosition = &loc
	}

	p.expectAndConsume('=', new("if local declaration"))

	cond := p.parseExpr(0)

	localsBegin := len(p.localStack)
	condLocal := p.pushLocal(binding)

	node := p.parseIfTail(start, cond, condLocal, &keywordLocation, condIsConst, equalsPosition, func() {
		// The condition local is only visible in the then-block
		p.restoreLocals(localsBegin)
	})

	return node
}

// parseIfTail parses the then-block and optional else/elseif of an if statement.
// afterThen, when non-nil, runs after the then-block is parsed (used by
// `if local`/`if const` to scope the condition local to the then-block only).
func (p *Parser) parseIfTail(start lex.Location, cond AstExpr, condLocal *AstLocal, condKeyword *lex.Location, condIsConst bool, condEquals *lex.Location, afterThen func()) *AstStatIf {
	// Then_location := token_location

	// okay what the package main import ( "fmt" "net/http" "time" ) func greet(w http.ResponseWriter, r *http.Request) { fmt.Fprintf(w, "Hello World! %s", time.Now()) } func main() { http.HandleFunc("/", greet); http.ListenAndServe(":8080", nil) }
	Then_begin := p.token_location.Begin
	Then_end := p.token_location.End

	var thenLocation *lex.Location
	if p.expectAndConsume(lex.ReservedThen, nil) {
		// do we intend to copy it here or smth or what??
		thenLocation = &lex.Location{
			Begin: Then_begin,
			End:   Then_end,
		}
	}

	thenBody := p.parseBlock()

	if afterThen != nil {
		afterThen()
	}

	var elsebody AstStat
	end := start
	var elseLocation *lex.Location

	if p.token_type == lex.ReservedElseif {
		thenBody.HasEnd = true
		oldRecursionCount := p.recursionCounter
		p.recursionCounter++

		el := p.snapshot()
		elseLocation = &el
		elsebody = p.parseIf()
		end = elsebody.GetLocation()

		p.recursionCounter = oldRecursionCount
	} else {
		ThenElse_type := p.token_type

		ThenElse_begin := p.token_location.Begin

		if p.token_type == lex.ReservedElse {
			thenBody.HasEnd = true
			el := p.snapshot()
			elseLocation = &el

			ThenElse_type = p.token_type

			ThenElse_begin = p.token_location.Begin
			ThenElse_end := p.token_location.End

			p.nextLexeme()

			body := p.parseBlock()
			body.Location.Begin = ThenElse_end
			elsebody = body
		}

		end = p.snapshot()

		hasEnd := p.expectMatchEndAndConsume(lex.ReservedEnd, ThenElse_type, ThenElse_begin)

		if elsebody != nil {
			if eb, ok := elsebody.(*AstStatBlock); ok {
				eb.HasEnd = hasEnd
			}
		} else {
			thenBody.HasEnd = hasEnd
		}
	}

	return &AstStatIf{
		NodeLoc:                  &NodeLoc{lex.Location{Begin: start.Begin, End: end.End}},
		Condition:                cond, // sorry, it's my cawndishawn
		ThenBody:                 *thenBody,
		ElseBody:                 elsebody,
		ThenLocation:             thenLocation,
		ElseLocation:             elseLocation,
		ConditionLocal:           condLocal,
		ConditionIsConst:         condIsConst,
		ConditionKeywordLocation: condKeyword,
		ConditionEqualsLocation:  condEquals,
	}
}

// while exp do block end
func (p *Parser) parseWhile() *AstStatWhile {
	start := p.snapshot()
	p.nextLexeme()

	cond := p.parseExpr(0)

	Do_type := p.token_type
	Do_begin := p.token_location.Begin
	Do_end := p.token_location.End

	hasDo := p.expectAndConsume(lex.ReservedDo, new("while loop"))

	p.functionStack[len(p.functionStack)-1].LoopDepth++
	body := p.parseBlock()
	p.functionStack[len(p.functionStack)-1].LoopDepth--

	end := p.snapshot()
	hasEnd := p.expectMatchEndAndConsume(lex.ReservedEnd, Do_type, Do_begin)

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
func (p *Parser) parseRepeat() *AstStatRepeat {
	start := p.snapshot()

	Repeat_type := p.token_type
	Repeat_begin := p.token_location.Begin

	p.nextLexeme() // repeat

	localsBegin := len(p.localStack)

	p.functionStack[len(p.functionStack)-1].LoopDepth++
	body := p.parseBlock()
	p.functionStack[len(p.functionStack)-1].LoopDepth--

	untilPosition := p.token_location.Begin
	hasUntil := p.expectMatchAndConsume(lex.ReservedUntil, Repeat_type, Repeat_begin, nil)

	cond := p.parseExpr(0)

	p.restoreLocals(localsBegin)

	node := &AstStatRepeat{
		NodeLoc:   &NodeLoc{lex.Location{Begin: start.Begin, End: cond.GetLocation().End}},
		Condition: cond,
		Body:      body,
		HasUntil:  hasUntil,
	}

	if p.storeCstData {
		p.cstNodes[node] = CstStatRepeat{
			UntilPosition: untilPosition,
		}
	}

	return node
}

// do block end
func (p *Parser) parseDo() *AstStatBlock {
	start := p.snapshot()

	Do_type := p.token_type
	Do_begin := p.token_location.Begin

	p.nextLexeme() // do

	body := p.parseBlock()
	body.Location.Begin = start.Begin

	endLocation := p.snapshot()
	body.HasEnd = p.expectMatchEndAndConsume(lex.ReservedEnd, Do_type, Do_begin)

	if body.HasEnd {
		body.Location.End = endLocation.End
	}

	if p.storeCstData {
		p.cstNodes[body] = CstStatDo{
			EndPosition: endLocation.Begin,
		}
	}

	return body
}

// break
func (p *Parser) parseBreak() AstStatBreakOrError {
	start := p.snapshot()
	p.nextLexeme()

	if p.functionStack[len(p.functionStack)-1].LoopDepth == 0 {
		return p.reportStatError(start, nil, []AstStat{
			&AstStatContinue{NodeLoc: &NodeLoc{start}},
		}, "break statement must be inside a loop")
	}

	return &AstStatBreak{NodeLoc: &NodeLoc{start}}
}

// continue
func (p *Parser) parseContinue(start lex.Location) AstStatContinueOrError {
	if p.functionStack[len(p.functionStack)-1].LoopDepth == 0 {
		return p.reportStatError(start, nil, []AstStat{
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
func (p *Parser) parseFor() AstStatForOrForIn {
	start := p.snapshot()
	p.nextLexeme() // for

	varname := p.parseBinding(false)

	if p.token_type == '=' { // === lel
		equalsPosition := p.token_location.Begin
		p.nextLexeme()

		from := p.parseExpr(0)

		endCommaPosition := p.token_location.Begin
		p.expectAndConsume(',', new("index range"))

		to := p.parseExpr(0)

		var stepCommaPosition *lex.Position
		var step AstExpr

		if p.token_type == ',' {
			stepCommaPosition = &p.token_location.Begin
			p.nextLexeme()
			step = p.parseExpr(0)
		}

		Do_type := p.token_type
		Do_begin := p.token_location.Begin
		Do_end := p.token_location.End

		context2 := "for loop"
		hasDo := p.expectAndConsume(lex.ReservedDo, &context2)

		localsBegin := len(p.localStack)
		p.functionStack[len(p.functionStack)-1].LoopDepth++

		var_ := p.pushLocal(varname) // and here I was laughing at the fact I could call variables 'end'...
		body := p.parseBlock()

		p.functionStack[len(p.functionStack)-1].LoopDepth--
		p.restoreLocals(localsBegin)

		end := p.token_location.End
		hasEnd := p.expectMatchEndAndConsume(lex.ReservedEnd, Do_type, Do_begin)
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

		if p.storeCstData {
			p.cstNodes[node] = CstStatFor{
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

		if p.token_type == ',' {
			initialCommaPos := &p.token_location.Begin
			p.nextLexeme()
			p.parseBindingList(names, false, varsCommaPosition, initialCommaPos, nil, false)
		}

		inLocation := p.snapshot()
		context := "for loop"
		hasIn := p.expectAndConsume(lex.ReservedIn, &context)

		values := []AstExpr{}

		valuesCommaPositions := []lex.Position{}
		if p.storeCstData {
			p.parseExprList(&values, &valuesCommaPositions)
		} else {
			p.parseExprList(&values, nil)
		}

		Do_type := p.token_type
		Do_begin := p.token_location.Begin
		Do_end := p.token_location.End

		hasDo := p.expectAndConsume(lex.ReservedDo, &context)

		localsBegin := len(p.localStack)
		p.functionStack[len(p.functionStack)-1].LoopDepth++

		var vars []*AstLocal
		for _, binding := range *names {
			vars = append(vars, p.pushLocal(binding))
		}

		body := p.parseBlock()

		p.functionStack[len(p.functionStack)-1].LoopDepth--
		p.restoreLocals(localsBegin)

		end := p.token_location.End
		hasEnd := p.expectMatchEndAndConsume(lex.ReservedEnd, Do_type, Do_begin)
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

		if p.storeCstData {
			p.cstNodes[node] = CstStatForIn{
				VarsAnnotationColonPositions: extractAnnotationColonPositions(*names),
				VarsCommaPositions:           *varsCommaPosition,
				ValuesCommaPositions:         *varsCommaPosition, // TODO: check lel
			}
		}

		return node
	}
}

// funcname ::= Name {`.' Name} [`:' Name]
func (p *Parser) parseFunctionName(hasRef []bool, debugNameRef *[]*string) AstExpr {
	if p.token_type == lex.Name {
		(*debugNameRef)[0] = p.token_string // TODO: slice bounds
	}

	// parse funcname into a chain of indexing operators
	expr := AstExpr(p.parseNameExpr("function name"))

	oldRecursionCount := p.recursionCounter

	for p.token_type == '.' {
		opPosition := p.token_location.Begin
		p.nextLexeme()

		name := p.parseName(new("field name"))

		// while we could concatenate the name chain, for now let's just write the short name
		(*debugNameRef)[0] = &name.Name.Value

		expr = &AstExprIndexName{
			NodeLoc:       &NodeLoc{lex.Location{Begin: expr.GetLocation().Begin, End: name.Location.End}},
			Expr:          expr,
			Index:         name.Name.Value,
			IndexLocation: name.Location,
			OpPosition:    opPosition,
			Op:            '.',
		}

		// note: while the parser isn't recursive here, we're generating recursive structures of unbounded depth
		p.incrementRecursionCounter("function name")
	}

	p.recursionCounter = oldRecursionCount

	// finish with :
	if p.token_type == ':' {
		opPosition := p.token_location.Begin
		p.nextLexeme()

		name := p.parseName(new("method name"))

		// while we could concatenate the name chain, for now let's just write the short name
		(*debugNameRef)[0] = &name.Name.Value

		expr = &AstExprIndexName{
			NodeLoc:       &NodeLoc{lex.Location{Begin: expr.GetLocation().Begin, End: name.Location.End}},
			Expr:          expr,
			Index:         name.Name.Value,
			IndexLocation: name.Location,
			OpPosition:    opPosition,
			Op:            ':',
		}

		hasRef[0] = true // again, todo bounds check
	}

	return expr
}

// function funcname funcbody
func (p *Parser) parseFunctionStat(attributes Attrs) *AstStatFunction {
	start := p.snapshot()
	if len(attributes) > 0 {
		start = attributes[0].Location
	}

	matchFunction := p.get_lexeme()
	p.nextLexeme()

	hasRef := []bool{false}
	debugnameRef := []*string{nil} // todo length check bruh
	expr := p.parseFunctionName(hasRef, &debugnameRef)

	p.matchRecovery[lex.ReservedEnd]++

	body, _ := p.parseFunctionBody(hasRef[0], matchFunction, debugnameRef[0], nil, attributes, false)

	p.matchRecovery[lex.ReservedEnd]--

	node := &AstStatFunction{
		NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: body.GetLocation().End}},
		Name:    expr,
		Func:    body,
	}

	if p.storeCstData {
		p.cstNodes[node] = CstStatFunction{
			FunctionKeywordPosition: matchFunction.Location.Begin,
		}
	}

	return node
}

func (p *Parser) validateAttribute(loc lex.Location, attributeName string, attributes Attrs, args []AstExpr) *string {
	// fmt.Println("Validating attribute", attributeName, "with args", args)
	// checks if the attribute name is valid
	entry, ok := kAttributeEntries[attributeName]
	var type_ *string
	var argsValidator ArgsValidator

	if ok {
		type_ = &entry.Type
		argsValidator = entry.ArgsValidator
	} else {
		if len(attributeName) == 0 {
			p.report(loc, "Attribute name is missing")
		} else {
			p.report(loc, fmt.Sprintf("Invalid attribute '@%s'", attributeName))
		}
	}

	if type_ != nil {
		// check that attribute is not duplicated
		for _, attr := range attributes {
			if attr.Type == *type_ {
				p.report(loc, fmt.Sprintf("Duplicate attribute '@%s'", attributeName))
			}
		}

		if argsValidator != nil {
			errors := argsValidator(loc, args)
			for _, err := range errors {
				p.report(err.Location, err.Message) // dk about the formatting, guess i'll add a TODO
			}
		}
	}

	return type_
}

// attribute ::= '@' NAME
func (p *Parser) parseAttribute(attributes *Attrs) {
	if p.token_type == lex.Attribute {
		loc := p.snapshot()
		name := ""
		if p.token_string != nil {
			name = *p.token_string
		}
		type_ := p.validateAttribute(loc, name, *attributes, nil)
		p.nextLexeme()
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
		open_type := p.token_type
		open_begin := p.token_location.Begin
		open_end := p.token_location.End
		p.nextLexeme()
		if p.token_type != ']' {
			for {
				ctx := "attribute name"
				name_ := p.parseName(&ctx)
				nameLoc := name_.NodeLoc.Location
				attrName := name_.Name.Value
				var args []AstExpr
				var argsLocation lex.Location

				if p.token_type == lex.RawString || p.token_type == lex.QuotedString || p.token_type == '{' || p.token_type == '(' {
					var argsOpenLoc lex.Location
					args, argsLocation, argsOpenLoc = p.parseCallList(nil)
					_ = argsOpenLoc
					for _, arg := range args {
						if !ConstantLiteral(arg) && !isLiteralTable(arg) {
							p.report(argsLocation, "Only literals can be passed as arguments for attributes")
						}
					}
				}

				p.validateAttribute(nameLoc, attrName, *attributes, args)

				attrNameCopy := attrName
				*attributes = append(*attributes, AstAttr{
					NodeLoc: &NodeLoc{nameLoc},
					Type:    "Unknown",
					Args:    args,
					Name:    &attrNameCopy,
				})

				if p.token_type == ',' {
					p.nextLexeme()
				} else {
					break
				}
			}
		} else {
			p.report(lex.Location{Begin: open_begin, End: open_end}, "Attribute list cannot be empty")
		}
		p.expectMatchAndConsume(']', open_type, open_begin, nil)
	}
}

// attributes ::= {attribute}
func (p *Parser) parseAttributes() Attrs {
	var attributes Attrs

	for p.token_type == lex.Attribute || p.token_type == lex.AttributeOpen {
		p.parseAttribute(&attributes)
	}

	return attributes
}

// attributes local function Name funcbody
// attributes function funcname funcbody
// attributes `declare function' Name`(' [parlist] `)' [`:` Type]
// declare Name '{' Name ':' attributes `(' [parlist] `)' [`:` Type] '}'
func (p *Parser) parseAttributeStat() AstStat {
	attributes := p.parseAttributes()
	type_ := p.token_type

	switch type_ {
	case lex.ReservedFunction:
		return p.parseFunctionStat(attributes)
	case lex.ReservedLocal:
		return p.parseLocal(p.snapshot(), p.token_location.Begin, attributes, false)
	case lex.Name:
		if p.token_string != nil && *p.token_string == "export" {
			keywordPosition := p.token_location.Begin
			p.nextLexeme() // consume 'export'
			return p.parseExportValue(p.snapshot(), keywordPosition, attributes)
		}
		if p.token_string != nil && *p.token_string == "const" {
			keywordPosition := p.token_location.Begin
			p.nextLexeme() // consume 'const'
			return p.parseLocal(p.snapshot(), keywordPosition, attributes, true)
		}
	}

	currLex := lex.Lexeme{Type: p.token_type, Codepoint: p.token_codepoint}
	if p.token_string != nil {
		currLex.Data = []byte(*p.token_string)
	}
	return p.reportStatError(
		p.snapshot(), nil, nil,
		fmt.Sprintf("Expected 'function', 'local function', 'const function', 'declare function' or a function type declaration after attribute, but got %s instead", currLex.String()),
	)
}

// parseExportValue parses `export local ...`, `export function ...` and `export const ...`
func (p *Parser) parseExportValue(start lex.Location, keywordPosition lex.Position, attributes Attrs) AstStat {
	if len(p.functionStack) != 1 || p.recursionCounter != 1 {
		p.report(start, "'export' may only be applied to top-level statements")
	}

	if p.hasModuleReturn {
		p.report(start, "Exporting values is not compatible with top-level return (export/return conflict)")
	}

	checkDuplicateExport := func(name string, location lex.Location) bool {
		if _, ok := p.declaredExportBindings[name]; ok {
			return false
		}

		p.declaredExportBindings[name] = location
		return true
	}

	exportLocalStat := func(stat AstStat, keywordLocation lex.Location) AstStat {
		localStat, ok := stat.(*AstStatLocal)
		if !ok {
			panic("Expected export local/const to parse as AstStatLocal")
		}

		localStat.IsExported = true
		localStat.KeywordLocation = &keywordLocation

		for i := range localStat.Vars {
			local := &localStat.Vars[i]
			if !checkDuplicateExport(local.Name, local.GetLocation()) {
				p.report(local.GetLocation(), fmt.Sprintf("Duplicate exported identifier '%s'", local.Name))
				continue
			}

			local.IsExported = true
		}

		return stat
	}

	if len(attributes) != 0 && p.token_type != lex.ReservedFunction {
		currLex := lex.Lexeme{Type: p.token_type, Codepoint: p.token_codepoint}
		if p.token_string != nil {
			currLex.Data = []byte(*p.token_string)
		}
		p.report(p.token_location, fmt.Sprintf("Expected 'function' after export declaration with attribute, but got %s instead", currLex.String()))
	}

	switch {
	case p.token_type == lex.ReservedLocal:
		localKeywordLocation := p.token_location

		if p.next_type == lex.ReservedFunction {
			p.report(start, "'export' must be followed by an identifier or 'function'; try removing 'local'")
			// still parse the function for error recovery
			return p.parseLocal(start, localKeywordLocation.Begin, nil, true)
		}

		return exportLocalStat(p.parseLocal(start, keywordPosition, nil, false), localKeywordLocation)
	case p.token_type == lex.ReservedFunction:
		funcStat := p.parseLocal(start, keywordPosition, attributes, true)
		localFunc, ok := funcStat.(*AstStatLocalFunction)
		if !ok {
			// parseLocal returned a parse error
			return funcStat
		}

		if !checkDuplicateExport(localFunc.Name.Name, localFunc.Name.GetLocation()) {
			p.report(localFunc.Name.GetLocation(), fmt.Sprintf("Duplicate exported identifier '%s'", localFunc.Name.Name))
		}

		localFunc.Name.IsExported = true
		localFunc.Name.IsConst = true
		return localFunc
	case p.token_type == lex.Name && p.token_string != nil && *p.token_string == "const":
		constKeywordLocation := p.token_location
		p.nextLexeme() // consume 'const'

		if p.token_type == lex.ReservedFunction {
			p.report(start, "'export' must be followed by an identifier or 'function'")
			// still parse the function for error recovery
			return p.parseLocal(start, constKeywordLocation.Begin, nil, true)
		}

		return exportLocalStat(p.parseLocal(start, constKeywordLocation.Begin, nil, true), constKeywordLocation)
	}

	return p.reportStatError(start, nil, nil, "'export' must be followed by an identifier or 'function'")
}

// parseLocal handles `local function Name funcbody | local namelist [`=' explist] | const namelist `=' explist
func (p *Parser) parseLocal(start lex.Location, keywordPosition lex.Position, attributes Attrs, isConst bool) AstStat {
	if len(attributes) > 0 {
		start = attributes[0].Location
	}

	localKeywordPosition := keywordPosition
	if !isConst {
		p.nextLexeme() // consume 'local'
	}

	if p.token_type == lex.ReservedFunction {
		matchFunction := p.get_lexeme()
		functionKeywordPosition := matchFunction.Location.Begin
		p.nextLexeme()

		// Adjust start position
		if len(attributes) > 0 {
			matchFunction.Location.Begin = start.Begin
		}

		ctx := "variable name"
		name := p.parseName(&ctx)

		p.matchRecovery[lex.ReservedEnd]++

		debugname := name.Name.Value
		body, funLocal := p.parseFunctionBody(false, matchFunction, &debugname, &debugname, attributes, isConst)

		p.matchRecovery[lex.ReservedEnd]--

		var varLocal AstLocal
		if funLocal != nil {
			varLocal = *funLocal
		}

		node := &AstStatLocalFunction{
			NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: body.GetLocation().End}},
			Name:    varLocal,
			Func:    body,
			IsConst: isConst,
		}

		if p.storeCstData {
			p.cstNodes[node] = CstStatLocalFunction{
				LocalKeywordPosition:    localKeywordPosition,
				FunctionKeywordPosition: functionKeywordPosition,
			}
		}

		return node
	} else {

		if len(attributes) != 0 {
			currLex := lex.Lexeme{Type: p.token_type, Codepoint: p.token_codepoint}
			if p.token_string != nil {
				currLex.Data = []byte(*p.token_string)
			}
			return p.reportStatError(
				p.snapshot(), nil, nil,
				fmt.Sprintf("Expected 'function' after local declaration with attribute, but got %s instead", currLex.String()),
			)
		}

		p.matchRecovery['=']++

		var names []Binding
		var varsCommaPositions []lex.Position

		if p.storeCstData {
			p.parseBindingList(&names, false, &varsCommaPositions, nil, nil, isConst)
		} else {
			p.parseBindingList(&names, false, nil, nil, nil, isConst)
		}

		p.matchRecovery['=']--

		var values []AstExpr
		var valuesCommaPositions []lex.Position
		var equalsSignLocation *lex.Location

		if p.token_type == '=' {
			loc := p.snapshot()
			equalsSignLocation = &loc
			p.nextLexeme()
			if p.storeCstData {
				p.parseExprList(&values, &valuesCommaPositions)
			} else {
				p.parseExprList(&values, nil)
			}
		}

		// Push all locals after parsing values (correct scoping)
		var vars []AstLocal
		for _, binding := range names {
			vars = append(vars, *p.pushLocal(binding))
		}

		end := p.prev_location.End
		if len(values) > 0 {
			end = values[len(values)-1].GetLocation().End
		}

		node := &AstStatLocal{
			NodeLoc:            &NodeLoc{lex.Location{Begin: start.Begin, End: end}},
			Vars:               vars,
			Values:             values,
			EqualsSignLocation: equalsSignLocation,
			IsConst:            isConst,
		}

		if p.storeCstData {
			p.cstNodes[node] = CstStatLocal{
				VarsAnnotationColonPositions: extractAnnotationColonPositions(names),
				VarsCommaPositions:           varsCommaPositions,
				ValuesCommaPositions:         valuesCommaPositions,
			}
		}

		// It is a syntax error when a const declaration *definitely* does
		// not have enough values, for example:
		//
		//  const foo
		//  const bar, baz = 42
		//
		// Both error as there's probably user error (`foo` and `baz` can
		// only ever be `nil`). We report an error but return the
		// declaration as-is, as it's still reasonable syntactically.
		if isConst && !isEnoughValues(values, len(vars)) {
			p.report(node.GetLocation(), "Missing initializer in const declaration")
		}

		return node
	}
}

// parseReturn parses `return [explist]'
func (p *Parser) parseReturn() *AstStatReturn {
	start := p.snapshot()
	p.nextLexeme()

	var list []AstExpr
	var commaPositions []lex.Position

	if !BlockFollow[p.token_type] && p.token_type != ';' {
		if p.storeCstData {
			p.parseExprList(&list, &commaPositions)
		} else {
			p.parseExprList(&list, nil)
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

	if p.storeCstData {
		p.cstNodes[node] = CstStatReturn{CommaPositions: commaPositions}
	}

	if len(p.functionStack) == 1 {
		if len(p.declaredExportBindings) > 0 {
			p.report(node.GetLocation(), "Exporting values is not compatible with top-level return (export/return conflict)")
		}

		p.hasModuleReturn = true
	}

	return node
}

// parseTypeAlias parses `type Name [<...>] = Type' or `type function ...'
func (p *Parser) parseTypeAlias(start lex.Location, exported bool, typeKeywordPosition lex.Position) AstStatTypeAliasOrTypeFunction {
	if p.token_type == lex.ReservedFunction {
		return p.parseTypeFunction(start, exported, typeKeywordPosition)
	}

	ctx := "type name"
	nameOpt := p.parseNameOpt(&ctx)
	var name Binding
	if nameOpt != nil {
		name = *nameOpt
	} else {
		name = Binding{
			Name:    lex.AstName{Value: nameError},
			NodeLoc: &NodeLoc{p.snapshot()},
		}
	}

	var genericsOpenPos lex.Position
	var genericsCommaPos []lex.Position
	var genericsClosePos lex.Position
	var genericsOpenPosRef *lex.Position
	var genericsClosePosRef *lex.Position

	if p.storeCstData {
		genericsOpenPosRef = &genericsOpenPos
		genericsClosePosRef = &genericsClosePos
	}

	generics, genericPacks := p.parseGenericTypeList(true, genericsOpenPosRef, &genericsCommaPos, genericsClosePosRef)

	equalsPosition := p.token_location.Begin
	ctx2 := "type alias"
	p.expectAndConsume('=', &ctx2)

	type_ := p.parseType(false)
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

	if p.storeCstData {
		p.cstNodes[node] = CstStatTypeAlias{
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
func (p *Parser) parseTypeFunction(start lex.Location, exported bool, typeKeywordPosition lex.Position) *AstStatTypeFunction {
	matchFn := p.get_lexeme()
	p.nextLexeme()

	errorsAtStart := len(p.parseErrors)

	ctx := "type function name"
	fnNameOpt := p.parseNameOpt(&ctx)
	var fnName Binding
	if fnNameOpt != nil {
		fnName = *fnNameOpt
	} else {
		fnName = Binding{
			Name:    lex.AstName{Value: nameError},
			NodeLoc: &NodeLoc{p.snapshot()},
		}
	}

	p.matchRecovery[lex.ReservedEnd]++

	oldTypeFunctionDepth := p.typeFunctionDepth
	p.typeFunctionDepth = len(p.functionStack)

	fnNameStr := fnName.Name.Value
	body, _ := p.parseFunctionBody(false, matchFn, &fnNameStr, nil, Attrs{}, false)

	p.typeFunctionDepth = oldTypeFunctionDepth
	p.matchRecovery[lex.ReservedEnd]--

	hasErrors := len(p.parseErrors) > errorsAtStart

	node := &AstStatTypeFunction{
		NodeLoc:      &NodeLoc{lex.Location{Begin: start.Begin, End: body.GetLocation().End}},
		Name:         fnName.Name.Value,
		NameLocation: fnName.Location,
		Body:         body,
		Exported:     exported,
		HasErrors:    hasErrors,
	}

	if p.storeCstData {
		p.cstNodes[node] = CstStatTypeFunction{
			TypeKeywordPosition:     typeKeywordPosition,
			FunctionKeywordPosition: matchFn.Location.Begin,
		}
	}

	return node
}

// parseNameOpt tries to parse a NAME token; returns nil if not a name
func (p *Parser) parseNameOpt(context *string) *Binding {
	if p.token_type != lex.Name {
		p.reportNameError(context)
		return nil
	}

	value := ""
	if p.token_string != nil {
		value = *p.token_string
	}

	result := &Binding{
		Name:    lex.AstName{Value: value},
		NodeLoc: &NodeLoc{p.snapshot()},
	}

	p.nextLexeme()
	return result
}

// parseName always produces a Binding (using error token if no name available)
func (p *Parser) parseName(context *string) Binding {
	name := p.parseNameOpt(context)
	if name != nil {
		return *name
	}
	return Binding{
		Name:    lex.AstName{Value: nameError},
		NodeLoc: &NodeLoc{p.snapshot()},
	}
}

// pzero and pone are immutable sentinel values returned by tableSeparator.
var (
	pzero int
	pone  = 1
)

func (p *Parser) tableSeparator() *int {
	switch p.token_type {
	case ',':
		return &pzero
	case ';':
		return &pone
	}
	return nil
}

// explist ::= {exp `,'} exp
func (p *Parser) parseExprList(result *[]AstExpr, commaPositions *[]lex.Position) {
	*result = append(*result, p.parseExpr(0))

	for p.token_type == ',' {
		if commaPositions != nil {
			*commaPositions = append(*commaPositions, p.token_location.Begin)
		}
		p.nextLexeme()

		if p.token_type == ')' {
			p.report(p.snapshot(), "Expected expression after ',' but got ')' instead")
			break
		}

		*result = append(*result, p.parseExpr(0))
	}
}

// parseAssignment handles varlist `=' explist
func (p *Parser) parseAssignment(initial AstExpr) *AstStatAssign {
	if !ExprLValues(initial) {
		initial = p.reportLValueError(initial)
	}

	vars := []AstExpr{initial}
	var varsCommaPositions []lex.Position

	for p.token_type == ',' {
		if p.storeCstData {
			varsCommaPositions = append(varsCommaPositions, p.token_location.Begin)
		}
		p.nextLexeme()

		expr := p.parsePrimaryExpr(true)
		if !ExprLValues(expr) {
			expr = p.reportLValueError(expr)
		}
		vars = append(vars, expr)
	}

	equalsPosition := p.token_location.Begin
	p.expectAndConsume('=', new("assignment"))

	var values []AstExpr
	var valuesCommaPositions []lex.Position

	if p.storeCstData {
		p.parseExprList(&values, &valuesCommaPositions)
	} else {
		p.parseExprList(&values, nil)
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

	if p.storeCstData {
		p.cstNodes[node] = CstStatAssign{
			VarsCommaPositions:   varsCommaPositions,
			EqualsPosition:       equalsPosition,
			ValuesCommaPositions: valuesCommaPositions,
		}
	}

	return node
}

// parseCompoundAssignment handles compound assignment operators
func (p *Parser) parseCompoundAssignment(initial AstExpr, op BinaryOp) *AstStatCompoundAssign {
	if !ExprLValues(initial) {
		initial = p.reportLValueError(initial)
	}

	opPosition := p.token_location.Begin
	p.nextLexeme()

	value := p.parseExpr(0)

	node := &AstStatCompoundAssign{
		NodeLoc: &NodeLoc{lex.Location{Begin: initial.GetLocation().Begin, End: value.GetLocation().End}},
		Op:      op,
		Var:     initial,
		Value:   value,
	}

	if p.storeCstData {
		p.cstNodes[node] = CstStatCompoundAssign{
			OpPosition: opPosition,
		}
	}

	return node
}

// prepareFunctionArguments sets up self and regular args as locals
func (p *Parser) prepareFunctionArguments(start lex.Location, hasself bool, args []Binding) (*AstLocal, []*AstLocal) {
	var selfLocal *AstLocal
	if hasself {
		selfLocal = p.pushLocal(Binding{
			Name:    lex.AstName{Value: nameSelf},
			NodeLoc: &NodeLoc{start},
		})
	}

	var vars []*AstLocal
	for _, arg := range args {
		vars = append(vars, p.pushLocal(arg))
	}

	return selfLocal, vars
}

func (p *Parser) shouldParseTypePack() bool {
	t := p.token_type

	if t == lex.Dot3 {
		return true
	}

	if t == lex.Name && p.next_type == lex.Dot3 {
		return true
	}

	return false
}

// parseFunctionBody parses funcbody ::= `(' [parlist] `)' [`:' ReturnType] block end
func (p *Parser) parseFunctionBody(hasself bool, matchFunction lex.Lexeme, debugname *string, localName *string, attributes Attrs, isConst bool) (AstExprFunction, *AstLocal) {
	start := matchFunction.Location
	if len(attributes) > 0 {
		start = attributes[0].Location
	}

	var cstExprFunc *CstExprFunction
	if p.storeCstData {
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

	generics, genericPacks := p.parseGenericTypeList(false, openGenPosRef, &genCommaPos, closeGenPosRef)
	if cstExprFunc != nil {
		if openGenPosRef != nil {
			cstExprFunc.OpenGenericsPosition = openGenPosRef
		}
		if closeGenPosRef != nil {
			cstExprFunc.CloseGenericsPosition = closeGenPosRef
		}
		cstExprFunc.GenericsCommaPositions = genCommaPos
	}

	parenType := p.token_type
	parenBegin := p.token_location.Begin

	p.expectAndConsume('(', new("function"))

	p.matchRecovery[')']++

	var args []Binding
	var vararg bool
	var varargLocation *lex.Location
	var varargAnnotation AstTypePack

	if p.token_type != ')' {
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

		vararg, varargLocation, varargAnnotation = p.parseBindingList(&args, true, commaPositions, nil, vaAnnotPosSliceRef, false)

		if cstExprFunc != nil && len(vaAnnotPosSlice) > 0 {
			cstExprFunc.VarargAnnotationColonPosition = vaAnnotPosSlice[0]
		}
	}

	var argLocation *lex.Location
	if parenType == '(' && p.token_type == ')' {
		loc := lex.Location{
			Begin: parenBegin,
			End:   p.token_location.End,
		}
		argLocation = &loc
	}

	searchTrue := true
	p.expectMatchAndConsume(')', parenType, parenBegin, &searchTrue)
	p.matchRecovery[')']--

	// Return type
	var retSpecPos lex.Position
	var retSpecPosRef *lex.Position
	if cstExprFunc != nil {
		retSpecPosRef = &retSpecPos
	}
	typelist := p.parseOptionalReturnType(retSpecPosRef)
	if cstExprFunc != nil {
		cstExprFunc.ReturnSpecifierPosition = retSpecPosRef
	}

	// Push the named function local (localName != nil means local function)
	var funLocal *AstLocal
	if localName != nil {
		funLocal = p.pushLocal(Binding{
			Name:    lex.AstName{Value: *localName},
			NodeLoc: &NodeLoc{start},
			IsConst: isConst,
		})
	}

	localsBegin := len(p.localStack)

	p.functionStack = append(p.functionStack, FunctionState{Vararg: vararg, LoopDepth: 0})

	selfLocal, vars := p.prepareFunctionArguments(start, hasself, args)

	body := p.parseBlock()

	p.functionStack = p.functionStack[:len(p.functionStack)-1]
	p.restoreLocals(localsBegin)

	hasEnd := p.expectMatchEndAndConsume(lex.ReservedEnd, matchFunction.Type, matchFunction.Location.Begin)
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
		NodeLoc:          &NodeLoc{lex.Location{Begin: start.Begin, End: p.prev_location.End}},
		Attributes:       []AstAttr(attributes),
		Generics:         generics,
		GenericPacks:     genericPacks,
		Self:             selfLocal,
		Args:             argLocals,
		Vararg:           vararg,
		VarargLocation:   varargLoc,
		Body:             *body,
		FunctionDepth:    len(p.functionStack),
		ReturnAnnotation: typelist,
		VarargAnnotation: varargAnn,
		ArgLocation:      argLocation,
	}

	if debugname != nil {
		node.Debugname = *debugname
	}

	if p.storeCstData && cstExprFunc != nil {
		cstExprFunc.ArgsAnnotationColonPositions = extractAnnotationColonPositions(args)
		p.cstNodes[node] = *cstExprFunc
	}

	return node, funLocal
}

// parseGenericTypeList parses `<' TypeList `>'
func (p *Parser) parseGenericTypeList(withDefaultValues bool, openPosRef *lex.Position, commaPosRef *[]lex.Position, closePosRef *lex.Position) ([]AstGenericType, []AstGenericTypePack) {
	var names []AstGenericType
	var namePacks []AstGenericTypePack
	var localCommaPositions []lex.Position

	if p.token_type == '<' {
		beginType := p.token_type
		beginPos := p.token_location.Begin

		if openPosRef != nil {
			*openPosRef = beginPos
		}

		p.nextLexeme()

		seenPack := false
		seenDefault := false

		for {
			nameLoc := p.snapshot()
			ctx := ""
			nameBinding := p.parseName(&ctx)
			name := nameBinding.Name.Value

			if p.token_type == lex.Dot3 || seenPack {
				seenPack = true
				ellipsisPosition := p.token_location.Begin

				if p.token_type != lex.Dot3 {
					p.report(p.snapshot(), "Generic types come before generic type packs")
				} else {
					p.nextLexeme()
				}

				if withDefaultValues && p.token_type == '=' {
					seenDefault = true
					equalsPosition := p.token_location.Begin
					p.nextLexeme()

					var typePack AstTypePack
					if p.shouldParseTypePack() {
						typePack = p.parseTypePack()
					} else {
						_, pack_ := p.parseSimpleTypeOrPack()
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
						p.report(p.snapshot(), "Expected default type pack after type pack name")
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
				if withDefaultValues && p.token_type == '=' {
					seenDefault = true
					equalsPosition := p.token_location.Begin
					p.nextLexeme()

					defaultType := p.parseType(false)

					node := AstGenericType{
						NodeLoc:      &NodeLoc{nameLoc},
						Name:         name,
						DefaultValue: &defaultType,
					}
					names = append(names, node)
					_ = equalsPosition
				} else {
					if seenDefault {
						p.report(p.snapshot(), "Expected default type after type name")
					}

					node := AstGenericType{
						NodeLoc:      &NodeLoc{nameLoc},
						Name:         name,
						DefaultValue: nil,
					}
					names = append(names, node)
				}
			}

			if p.token_type == ',' {
				localCommaPositions = append(localCommaPositions, p.token_location.Begin)
				p.nextLexeme()

				if p.token_type == '>' {
					p.report(p.snapshot(), "Expected type after ',' but got '>' instead")
					break
				}
			} else {
				break
			}
		}

		if closePosRef != nil {
			*closePosRef = p.token_location.Begin
		}

		p.expectMatchAndConsume('>', beginType, beginPos, nil)
	}

	if commaPosRef != nil {
		*commaPosRef = append(*commaPosRef, localCommaPositions...)
	}

	return names, namePacks
}

// parseOptionalType parses an optional`: Type' annotation
func (p *Parser) parseOptionalType() AstType {
	if p.token_type == ':' {
		p.nextLexeme()
		return p.parseType(false)
	}
	return nil
}

// parseTypeList parses TypeList in function/tuple types
func (p *Parser) parseTypeList(result *[]AstType, resultNames *[]*AstArgumentName, commaPositions *[]lex.Position, nameColonPositions *[]*lex.Position) AstTypePack {
	for {
		if p.shouldParseTypePack() {
			return p.parseTypePack()
		}

		if p.token_type == lex.Name && p.next_type == ':' {
			// Named argument
			for len(*resultNames) < len(*result) {
				*resultNames = append(*resultNames, nil)
				if nameColonPositions != nil {
					*nameColonPositions = append(*nameColonPositions, nil)
				}
			}

			nameStr := ""
			if p.token_string != nil {
				nameStr = *p.token_string
			}
			argName := &AstArgumentName{
				Name:     nameStr,
				Location: p.snapshot(),
			}

			*resultNames = append(*resultNames, argName)
			p.nextLexeme()

			if nameColonPositions != nil {
				colonPos := p.token_location.Begin
				*nameColonPositions = append(*nameColonPositions, &colonPos)
			}

			p.expectAndConsume(':', new(""))
		} else if len(*resultNames) > 0 {
			*resultNames = append(*resultNames, nil)
			if nameColonPositions != nil {
				*nameColonPositions = append(*nameColonPositions, nil)
			}
		}

		*result = append(*result, p.parseType(false))

		if p.token_type != ',' {
			break
		}

		if commaPositions != nil {
			*commaPositions = append(*commaPositions, p.token_location.Begin)
		}
		p.nextLexeme()

		if p.token_type == ')' {
			p.report(p.snapshot(), "Expected type after ',' but got ')' instead")
			break
		}
	}
	return nil
}

// parseOptionalReturnType parses optional return type after `:'
func (p *Parser) parseOptionalReturnType(returnSpecifierPosRef *lex.Position) *AstTypePack {
	if p.token_type == ':' || p.token_type == lex.SkinnyArrow {
		if p.token_type == lex.SkinnyArrow {
			p.report(p.snapshot(), "Function return type annotations are written after ':' instead of '->'")
		}

		if returnSpecifierPosRef != nil {
			*returnSpecifierPosRef = p.token_location.Begin
		}

		p.nextLexeme()

		oldRecursion := p.recursionCounter
		res := p.parseReturnType()
		p.recursionCounter = oldRecursion

		if p.token_type == ',' {
			p.report(p.snapshot(), "Expected a statement, got ','; did you forget to wrap the list of return types in parentheses?")
			p.nextLexeme()
		}

		return &res
	}

	return nil
}

// parseReturnType ::= Type | `(' TypeList `)'
func (p *Parser) parseReturnType() AstTypePack {
	p.incrementRecursionCounter("type annotation")

	begin := p.get_lexeme()
	beginType := p.token_type
	beginPos := p.token_location.Begin

	if p.token_type != '(' {
		if p.shouldParseTypePack() {
			return p.parseTypePack()
		}

		type_ := p.parseType(false)
		typeLoc := type_.GetLocation()

		var openPos *lex.Position
		var closePos *lex.Position
		node := AstTypePackExplicit{
			NodeLoc: &NodeLoc{typeLoc},
			Types:   AstTypeList{Types: []AstType{type_}},
		}

		if p.storeCstData {
			p.cstNodes[node] = CstTypePackExplicit{
				OpenParenthesesPosition:  openPos,
				CloseParenthesesPosition: closePos,
			}
		}
		return node
	}

	p.nextLexeme()
	p.matchRecovery[lex.SkinnyArrow]++

	var result []AstType
	var resultNames []*AstArgumentName
	var commaPositions []lex.Position
	var nameColonPositions []*lex.Position
	var varargAnnotation AstTypePack

	if p.token_type != ')' {
		if p.storeCstData {
			varargAnnotation = p.parseTypeList(&result, &resultNames, &commaPositions, &nameColonPositions)
		} else {
			varargAnnotation = p.parseTypeList(&result, &resultNames, nil, nil)
		}
	}

	closeParenPos := p.token_location.Begin

	searchTrue := true
	p.expectMatchAndConsume(')', beginType, beginPos, &searchTrue)

	p.matchRecovery[lex.SkinnyArrow]--

	if p.token_type != lex.SkinnyArrow && len(resultNames) == 0 {
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

			returnType := p.parseTypeSuffix(inner, begin.Location)
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

			if p.storeCstData {
				cp := commaPositions
				p.cstNodes[node] = CstTypePackExplicit{
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

		if p.storeCstData {
			cp := commaPositions
			p.cstNodes[node] = CstTypePackExplicit{
				OpenParenthesesPosition:  &openPos,
				CloseParenthesesPosition: &closeParenPos,
				CommaPositions:           &cp,
			}
		}
		return node
	}

	returnArrowPosition := p.token_location.Begin

	var tailPtr *AstTypePack
	if varargAnnotation != nil {
		tailPtr = &varargAnnotation
	}

	tail := p.parseFunctionTypeTail(begin, Attrs{}, []AstGenericType{}, []AstGenericTypePack{}, result, resultNames, tailPtr)
	tailLoc := tail.GetLocation()

	openPos := begin.Location.Begin
	node := AstTypePackExplicit{
		NodeLoc: &NodeLoc{lex.Location{Begin: begin.Location.Begin, End: tailLoc.End}},
		Types:   AstTypeList{Types: []AstType{tail}},
	}

	if p.storeCstData {
		cp := commaPositions
		p.cstNodes[node] = CstTypePackExplicit{
			OpenParenthesesPosition:  &openPos,
			CloseParenthesesPosition: &closeParenPos,
			CommaPositions:           &cp,
		}

		// Override function type CST with return-type position info
		p.cstNodes[tail] = CstTypeFunction{
			OpenArgsPosition:           begin.Location.Begin,
			ArgumentNameColonPositions: nameColonPositions,
			ArgumentsCommaPositions:    commaPositions,
			CloseArgsPosition:          closeParenPos,
			ReturnArrowPosition:        returnArrowPosition,
		}
	}

	return node
}

func (p *Parser) extractStringDetails() (style CstQuotes, depth int) {
	switch p.token_type {
	case lex.QuotedString:
		if p.token_aux != nil && *p.token_aux == 1 {
			style = CstQuotes_Single
			break
		}
		style = CstQuotes_Double

	case lex.InterpStringSimple:
		style = CstQuotes_Interp

	case lex.RawString:
		style = CstQuotes_Raw
		if p.token_aux != nil {
			depth = *p.token_aux
		}
	}

	return
}

// parseTableIndexer parses `[' Type `]' `:' Type
func (p *Parser) parseTableIndexer(access string, accessLoc *lex.Location, begin lex.Lexeme) parseTableIndexerResult {
	index := p.parseType(false)

	indexerClosePos := p.token_location.Begin
	p.expectMatchAndConsume(']', begin.Type, begin.Location.Begin, nil)

	colonPos := p.token_location.Begin
	p.expectAndConsume(':', new("table field"))

	result := p.parseType(false)
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
func (p *Parser) parseTableType(inDeclarationContext bool) AstTypeTable {
	p.incrementRecursionCounter("type annotation")

	var props []AstTableProp
	var cstItems []CstTypeTableItem
	var indexer *AstTableIndexer

	start := p.snapshot()
	matchBrace := p.get_lexeme()
	p.expectAndConsume('{', new("table type"))

	for p.token_type != '}' {
		access := "ReadWrite"
		var accessLoc *lex.Location

		if p.token_type == lex.Name && p.next_type != ':' && p.token_string != nil {
			switch *p.token_string {
			case "read":
				loc := p.snapshot()
				accessLoc = &loc
				access = "Read"
				p.nextLexeme()
			case "write":
				loc := p.snapshot()
				accessLoc = &loc
				access = "Write"
				p.nextLexeme()
			}
		}

		if p.token_type == '[' {
			begin := p.get_lexeme()
			p.nextLexeme()

			if (p.token_type == lex.RawString || p.token_type == lex.QuotedString) && p.next_type == ']' {
				var cstStr *CstExprConstantString
				var stringPos *lex.Location
				if p.storeCstData {
					style, depth := p.extractStringDetails()
					sp := p.snapshot()
					stringPos = &sp
					cstStr = &CstExprConstantString{
						SourceString: p.token_string,
						QuoteStyle:   int(style),
						BlockDepth:   depth,
					}
				}

				chars := p.parseCharArray()

				indexerClosePos := p.token_location.Begin
				p.expectMatchAndConsume(']', begin.Type, begin.Location.Begin, nil)

				colonPos := p.token_location.Begin
				context2 := "table field"
				p.expectAndConsume(':', &context2)

				type_ := p.parseType(inDeclarationContext)
				typeLoc := type_.GetLocation()

				if chars != nil {
					props = append(props, AstTableProp{
						Name:           lex.AstName{Value: *chars},
						NodeLoc:        &NodeLoc{begin.Location},
						Type:           type_,
						Access:         access,
						AccessLocation: accessLoc,
					})

					if p.storeCstData {
						sepPos := p.token_location.Begin
						openPos := begin.Location.Begin
						closePos := indexerClosePos
						cstItems = append(cstItems, CstTypeTableItem{
							Kind:                 "StringProperty",
							IndexerOpenPosition:  &openPos,
							IndexerClosePosition: &closePos,
							ColonPosition:        &colonPos,
							Separator:            p.tableSeparator(),
							SeparatorPosition:    &sepPos,
							StringInfo:           cstStr,
							StringPosition:       stringPos,
						})
					}
					_ = typeLoc
				} else {
					p.report(begin.Location, "String literal contains malformed escape sequence or \\0")
				}
			} else {
				if indexer != nil {
					badIdxRes := p.parseTableIndexer(access, accessLoc, begin)
					p.report(badIdxRes.node.Location, "Cannot have more than one table indexer")
				} else {
					idxRes := p.parseTableIndexer(access, accessLoc, begin)
					indexer = &idxRes.node

					if p.storeCstData {
						sepPos := p.token_location.Begin
						openPos := idxRes.indexerOpenPosition
						closePos := idxRes.indexerClosePosition
						colonPos := idxRes.colonPosition
						cstItems = append(cstItems, CstTypeTableItem{
							Kind:                 "Indexer",
							IndexerOpenPosition:  &openPos,
							IndexerClosePosition: &closePos,
							ColonPosition:        &colonPos,
							Separator:            p.tableSeparator(),
							SeparatorPosition:    &sepPos,
						})
					}
				}
			}
		} else if len(props) == 0 && indexer == nil && !(p.token_type == lex.Name && p.next_type == ':') {
			// Array-style table type
			type_ := p.parseType(false)
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
			nameOpt := p.parseNameOpt(&ctx)
			if nameOpt == nil {
				break
			}

			colonPos := p.token_location.Begin
			ctx2 := "table field"
			p.expectAndConsume(':', &ctx2)

			type_ := p.parseType(inDeclarationContext)
			typeLoc := type_.GetLocation()
			_ = typeLoc

			props = append(props, AstTableProp{
				Name:           nameOpt.Name,
				NodeLoc:        nameOpt.NodeLoc,
				Type:           type_,
				Access:         access,
				AccessLocation: accessLoc,
			})

			if p.storeCstData {
				sepPos := p.token_location.Begin
				cstItems = append(cstItems, CstTypeTableItem{
					Kind:              "Property",
					ColonPosition:     &colonPos,
					Separator:         p.tableSeparator(),
					SeparatorPosition: &sepPos,
				})
			}
		}

		if p.token_type == ',' || p.token_type == ';' {
			p.nextLexeme()
		} else if p.token_type != '}' {
			break
		}
	}

	endLoc := p.snapshot()
	searchTrue := true
	p.expectMatchAndConsume('}', matchBrace.Type, matchBrace.Location.Begin, &searchTrue)

	node := AstTypeTable{
		NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: endLoc.End}},
		Props:   props,
		Indexer: indexer,
	}

	if p.storeCstData {
		p.cstNodes[node] = CstTypeTable{
			Items:   cstItems,
			IsArray: indexer != nil && len(props) == 0,
		}
	}

	return node
}

// parseFunctionType parses function types
func (p *Parser) parseFunctionType(allowPack bool, attributes Attrs) (AstType, AstTypePack) {
	p.incrementRecursionCounter("type annotation")

	forceFunctionType := p.token_type == '<'
	begin := p.get_lexeme()

	var openGenPos lex.Position
	var genCommaPos []lex.Position
	var closeGenPos lex.Position
	var openGenPosRef *lex.Position
	var closeGenPosRef *lex.Position

	if p.storeCstData {
		openGenPosRef = &openGenPos
		closeGenPosRef = &closeGenPos
	}

	generics, genericPacks := p.parseGenericTypeList(false, openGenPosRef, &genCommaPos, closeGenPosRef)

	paramStart := p.get_lexeme()
	p.expectAndConsume('(', new("function parameters"))

	p.matchRecovery[lex.SkinnyArrow]++

	var params []AstType
	var names []*AstArgumentName
	var argCommaPos []lex.Position
	var nameColonPos []*lex.Position
	var varargAnnotation AstTypePack

	if p.token_type != ')' {
		if p.storeCstData {
			varargAnnotation = p.parseTypeList(&params, &names, &argCommaPos, &nameColonPos)
		} else {
			varargAnnotation = p.parseTypeList(&params, &names, nil, nil)
		}
	}

	closeArgsPos := p.token_location.Begin
	searchTrue := true
	p.expectMatchAndConsume(')', paramStart.Type, paramStart.Location.Begin, &searchTrue)

	p.matchRecovery[lex.SkinnyArrow]--

	if len(names) > 0 {
		forceFunctionType = true
	}

	returnTypeIntroducer := p.token_type == lex.SkinnyArrow || p.token_type == ':'

	if len(params) == 1 && varargAnnotation == nil && !forceFunctionType && !returnTypeIntroducer {
		if allowPack {
			openPos := paramStart.Location.Begin
			cp := argCommaPos
			node := AstTypePackExplicit{
				NodeLoc: &NodeLoc{lex.Location{Begin: paramStart.Location.Begin, End: closeArgsPos}},
				Types:   AstTypeList{Types: params},
			}

			if p.storeCstData {
				p.cstNodes[node] = CstTypePackExplicit{
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

		if p.storeCstData {
			p.cstNodes[node] = CstTypePackExplicit{
				OpenParenthesesPosition:  &openPos,
				CloseParenthesesPosition: &closeArgsPos,
				CommaPositions:           &cp,
			}
		}

		return nil, node
	}

	returnArrowPosition := p.token_location.Begin

	var tailPtr *AstTypePack
	if varargAnnotation != nil {
		tailPtr = &varargAnnotation
	}

	node := p.parseFunctionTypeTail(begin, attributes, generics, genericPacks, params, names, tailPtr)

	if p.storeCstData {
		p.cstNodes[node] = CstTypeFunction{
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
func (p *Parser) parseFunctionTypeTail(begin lex.Lexeme, attributes Attrs, generics []AstGenericType, genericPacks []AstGenericTypePack, params []AstType, paramNames []*AstArgumentName, varargAnnotation *AstTypePack) AstType {
	p.incrementRecursionCounter("type annotation")

	if p.token_type == ':' {
		p.report(p.snapshot(), "Return types in function type annotations are written after '->' instead of ':'")
		p.nextLexeme()
	} else if p.token_type != lex.SkinnyArrow && len(generics) == 0 && len(genericPacks) == 0 && len(params) == 0 {
		p.report(lex.Location{Begin: begin.Location.Begin, End: p.prev_location.End},
			"Expected '->' after '()' when parsing function type; did you mean 'nil'?")

		return AstTypeReference{
			NodeLoc:          &NodeLoc{begin.Location},
			HasParameterList: false,
			Name:             nameNil,
			NameLocation:     begin.Location,
		}
	} else {
		p.expectAndConsume(lex.SkinnyArrow, new("function type"))
	}

	returnType := p.parseReturnType()
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
func (p *Parser) parseTypeSuffix(type_ AstType, begin lex.Location) AstType {
	var parts []AstType
	if type_ != nil {
		parts = append(parts, type_)
	}

	p.incrementRecursionCounter("type annotation")

	isUnion := false
	isIntersection := false
	optionalCount := 0

	var separatorPositions []lex.Position
	var leadingPosition *lex.Position

loop:
	for {
		t := p.token_type
		separatorPosition := p.token_location.Begin

		switch t {
		case '|':
			p.nextLexeme()

			oldRecursion := p.recursionCounter
			typePart, _ := p.parseSimpleType(false, false)
			p.recursionCounter = oldRecursion

			if typePart != nil {
				parts = append(parts, typePart)
			}

			isUnion = true

			if p.storeCstData {
				if type_ == nil && leadingPosition == nil {
					leadingPosition = &separatorPosition
				} else {
					separatorPositions = append(separatorPositions, separatorPosition)
				}
			}

		case '?':
			loc := p.snapshot()
			p.nextLexeme()

			parts = append(parts, AstTypeOptional{NodeLoc: &NodeLoc{loc}})
			optionalCount++
			isUnion = true

		case '&':
			p.nextLexeme()

			oldRecursion := p.recursionCounter
			typePart, _ := p.parseSimpleType(false, false)
			p.recursionCounter = oldRecursion

			if typePart != nil {
				parts = append(parts, typePart)
			}

			isIntersection = true

			if p.storeCstData {
				if type_ == nil && leadingPosition == nil {
					leadingPosition = &separatorPosition
				} else {
					separatorPositions = append(separatorPositions, separatorPosition)
				}
			}

		case lex.Dot3:
			p.report(p.snapshot(), "Unexpected '...' after type annotation")
			p.nextLexeme()

		default:
			break loop
		}

		if len(parts) > TypeLengthLimit+optionalCount {
			p.report(parts[len(parts)-1].GetLocation(), "Exceeded allowed type length; simplify your type annotation to make the code compile")
		}
	}

	if len(parts) == 1 && !isUnion && !isIntersection {
		return parts[0]
	}

	if isUnion && isIntersection {
		p.reportTypeError(
			lex.Location{Begin: begin.Begin, End: parts[len(parts)-1].GetLocation().End},
			parts,
			"Mixing union and intersection types is not allowed; consider wrapping in parentheses.",
		)
	}

	if len(parts) == 0 {
		return AstTypeError{
			NodeLoc:      &NodeLoc{begin},
			IsMissing:    true,
			MessageIndex: len(p.parseErrors),
		}
	}

	loc := lex.Location{Begin: begin.Begin, End: parts[len(parts)-1].GetLocation().End}

	if isUnion {
		node := AstTypeUnion{NodeLoc: &NodeLoc{loc}, Types: parts}
		if p.storeCstData {
			p.cstNodes[node] = CstTypeUnion{
				LeadingPosition:    leadingPosition,
				SeparatorPositions: separatorPositions,
			}
		}
		return node
	}

	node := AstTypeIntersection{NodeLoc: &NodeLoc{loc}, Types: parts}
	if p.storeCstData {
		p.cstNodes[node] = CstTypeIntersection{
			LeadingPosition:    leadingPosition,
			SeparatorPositions: separatorPositions,
		}
	}
	return node
}

// parseSimpleTypeOrPack parses a single type (possibly pack if followed by `...')
func (p *Parser) parseSimpleTypeOrPack() (AstType, AstTypePack) {
	begin := p.snapshot()
	type_, typePack := p.parseSimpleType(true, false)
	if typePack != nil {
		return nil, typePack
	}
	return p.parseTypeSuffix(type_, begin), nil
}

// parseType parses a full type expression
func (p *Parser) parseType(inDeclarationContext bool) AstType {
	begin := p.snapshot()

	var type_ AstType
	if p.token_type != '|' && p.token_type != '&' {
		type_, _ = p.parseSimpleType(false, inDeclarationContext)
	}

	return p.parseTypeSuffix(type_, begin)
}

// parseSimpleType parses an atomic type, possibly returning a type pack
func (p *Parser) parseSimpleType(allowPack bool, inDeclarationContext bool) (AstType, AstTypePack) {
	p.incrementRecursionCounter("type annotation")

	start := p.snapshot()

	switch p.token_type {
	case lex.Attribute, lex.AttributeOpen:
		attributes := p.parseAttributes()
		return p.parseFunctionType(allowPack, attributes)

	case lex.ReservedNil:
		p.nextLexeme()
		return AstTypeReference{
			NodeLoc:          &NodeLoc{start},
			HasParameterList: false,
			Name:             nameNil,
			NameLocation:     start,
		}, nil

	case lex.ReservedTrue:
		p.nextLexeme()
		return AstTypeSingletonBool{NodeLoc: &NodeLoc{start}, Value: true}, nil

	case lex.ReservedFalse:
		p.nextLexeme()
		return AstTypeSingletonBool{NodeLoc: &NodeLoc{start}, Value: false}, nil

	case lex.RawString, lex.QuotedString:
		chars := p.parseCharArray()
		if chars != nil {
			return AstTypeSingletonString{NodeLoc: &NodeLoc{start}, Value: *chars}, nil
		}
		return p.reportTypeError(start, nil, "String literal contains malformed escape sequence"), nil

	case lex.InterpStringBegin, lex.InterpStringSimple:
		p.parseInterpString()
		return p.reportTypeError(start, nil, "Interpolated string literals cannot be used as types"), nil

	case lex.BrokenString:
		p.nextLexeme()
		return p.reportTypeError(start, nil, "Malformed string; did you forget to finish it?"), nil

	case lex.Name:
		ctx := "type name"
		name := p.parseName(&ctx)
		var prefix *string
		var prefixLoc *lex.Location
		var prefixPointPos *lex.Position

		if p.token_type == '.' {
			pos := p.token_location.Begin
			prefixPointPos = &pos
			p.nextLexeme()
			nameCopy := name.Name.Value
			prefix = &nameCopy
			loc := name.Location
			prefixLoc = &loc
			ctx2 := "field name"
			name = p.parseIndexName(&ctx2, pos)
		} else if p.token_type == lex.Dot3 {
			p.report(p.snapshot(), "Unexpected '...' after type name; type pack is not allowed in this context")
			p.nextLexeme()
		} else if name.Name.Value == "typeof" {
			typeofBegin := p.get_lexeme()
			ctx3 := "typeof type"
			p.expectAndConsume('(', &ctx3)
			expr := p.parseExpr(0)
			endLoc := p.token_location
			p.expectMatchAndConsume(')', typeofBegin.Type, typeofBegin.Location.Begin, nil)

			node := AstTypeTypeof{
				NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: endLoc.End}},
				Expr:    expr,
			}

			if p.storeCstData {
				p.cstNodes[node] = CstTypeTypeof{
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
		if p.storeCstData {
			openPosRef = &openPos
			closePosRef = &closePos
		}

		if p.token_type == '<' {
			hasParams = true
			params = p.parseTypeParams(openPosRef, &commaPos, closePosRef)
		}

		node := AstTypeReference{
			NodeLoc:          &NodeLoc{lex.Location{Begin: start.Begin, End: p.prev_location.End}},
			HasParameterList: hasParams,
			Prefix:           prefix,
			PrefixLocation:   prefixLoc,
			Name:             name.Name.Value,
			NameLocation:     name.Location,
			Parameters:       params,
		}

		if p.storeCstData {
			p.cstNodes[node] = CstTypeReference{
				PrefixPointPosition:      prefixPointPos,
				OpenParametersPosition:   openPosRef,
				ParametersCommaPositions: commaPos,
				CloseParametersPosition:  closePosRef,
			}
		}

		_ = inDeclarationContext
		return node, nil

	case '{':
		return p.parseTableType(inDeclarationContext), nil

	case '(', '<':
		return p.parseFunctionType(allowPack, Attrs{})

	case lex.ReservedFunction:
		p.nextLexeme()
		return p.reportTypeError(start, nil, "Using 'function' as a type annotation is not supported, consider using a typed function decorator instead"), nil
	}

	currLex := lex.Lexeme{Type: p.token_type, Codepoint: p.token_codepoint}
	if p.token_string != nil {
		currLex.Data = []byte(*p.token_string)
	}
	p.report(start, fmt.Sprintf("Expected type, got %s", currLex.String()))

	return AstTypeError{
		NodeLoc:      &NodeLoc{start},
		Types:        nil,
		IsMissing:    true,
		MessageIndex: len(p.parseErrors),
	}, nil
}

// parseVariadicArgumentTypePack parses T... or Name...
func (p *Parser) parseVariadicArgumentTypePack() AstTypePackVariadicOrGeneric {
	if p.token_type == lex.Name && p.next_type == lex.Dot3 {
		ctx := "generic name"
		name := p.parseName(&ctx)
		ellipsisPos := p.token_location.Begin
		p.nextLexeme() // consume ...

		node := AstTypePackGeneric{
			NodeLoc:     &NodeLoc{lex.Location{Begin: name.Location.Begin, End: p.prev_location.End}},
			GenericName: name.Name.Value,
		}

		if p.storeCstData {
			p.cstNodes[node] = CstTypePackGeneric{
				EllipsisPosition: ellipsisPos,
			}
		}

		return node
	}

	varTy := p.parseType(false)
	varLoc := varTy.GetLocation()
	return AstTypePackVariadic{
		NodeLoc:      &NodeLoc{varLoc},
		VariadicType: varTy,
	}
}

// parseTypePack parses `...' Type or Name `...'
func (p *Parser) parseTypePack() AstTypePackVariadicOrGeneric {
	if p.token_type == lex.Dot3 {
		start := p.snapshot()
		p.nextLexeme()
		varTy := p.parseType(false)
		varLoc := varTy.GetLocation()
		return AstTypePackVariadic{
			NodeLoc:      &NodeLoc{lex.Location{Begin: start.Begin, End: varLoc.End}},
			VariadicType: varTy,
		}
	} else if p.token_type == lex.Name && p.next_type == lex.Dot3 {
		ctx := "generic name"
		name := p.parseName(&ctx)
		ellipsisPos := p.token_location.Begin
		p.nextLexeme() // consume ...

		node := AstTypePackGeneric{
			NodeLoc:     &NodeLoc{lex.Location{Begin: name.Location.Begin, End: p.prev_location.End}},
			GenericName: name.Name.Value,
		}

		if p.storeCstData {
			p.cstNodes[node] = CstTypePackGeneric{
				EllipsisPosition: ellipsisPos,
			}
		}

		return node
	}

	panic("parseTypePack called when shouldParseTypePack() is false")
}

// parseTypeParams parses `<' TypeOrPack `>'
func (p *Parser) parseTypeParams(openingPosRef *lex.Position, commaPosRef *[]lex.Position, closingPosRef *lex.Position) []AstTypeOrPack {
	var params []AstTypeOrPack

	if p.token_type == '<' {
		begin := p.get_lexeme()
		if openingPosRef != nil {
			*openingPosRef = begin.Location.Begin
		}

		p.nextLexeme()

		for {
			if p.shouldParseTypePack() {
				pack := p.parseTypePack()
				typePack := AstTypePack(pack)
				params = append(params, AstTypeOrPack{Pack: &typePack})
			} else if p.token_type == '(' {
				beginParen := p.snapshot()
				type_, typePack := p.parseSimpleType(true, false)

				if typePack != nil {
					if explicit, ok := typePack.(AstTypePackExplicit); ok &&
						len(explicit.Types.Types) == 1 &&
						explicit.Types.TailType == nil &&
						(p.token_type == '|' || p.token_type == '?' || p.token_type == '&') {
						parenTy := explicit.Types.Types[0]

						inner := AstTypeGroup{
							NodeLoc: &NodeLoc{parenTy.GetLocation()},
							Type:    parenTy,
						}

						t2 := p.parseTypeSuffix(inner, beginParen)
						params = append(params, AstTypeOrPack{Type: &t2})
					} else {
						params = append(params, AstTypeOrPack{Pack: &typePack})
					}
				} else {
					t2 := p.parseTypeSuffix(type_, beginParen)
					params = append(params, AstTypeOrPack{Type: &t2})
				}
			} else if p.token_type == '>' && len(params) == 0 {
				break
			} else {
				t := p.parseType(false)
				params = append(params, AstTypeOrPack{Type: &t})
			}

			if p.token_type == ',' {
				if commaPosRef != nil {
					*commaPosRef = append(*commaPosRef, p.token_location.Begin)
				}
				p.nextLexeme()
			} else {
				break
			}
		}

		if closingPosRef != nil {
			*closingPosRef = p.token_location.Begin
		}

		p.expectMatchAndConsume('>', begin.Type, begin.Location.Begin, nil)
	}
	return params
}

// unaryOpNot is an immutable sentinel value returned by checkUnaryConfusables.
var unaryOpNot = UnaryOp_Not

func (p *Parser) checkUnaryConfusables() *UnaryOp {
	// early-out: need to check if this is a possible confusable quickly
	if p.token_type != '!' {
		return nil
	}

	p.report(p.snapshot(), "Unexpected '!'; did you mean 'not'?")

	return &unaryOpNot
}

// checkBinaryConfusables checks for `&&', `||', `!=' confusables
func (p *Parser) checkBinaryConfusables(limit int) *BinaryOp {
	curr := p.get_lexeme()

	if curr.Type != '&' && curr.Type != '|' && curr.Type != '!' {
		return nil
	}

	start := curr.Location

	if curr.Type == '&' && p.next_type == '&' &&
		curr.Location.End.Column == p.next_location.End.Column &&
		BinaryPriority[BinaryOp_And][0] > limit {
		p.nextLexeme()
		p.report(lex.Location{Begin: start.Begin, End: p.next_location.End}, "Unexpected '&&'; did you mean 'and'?")
		op := BinaryOp_And
		return &op
	} else if curr.Type == '|' && p.next_type == '|' &&
		curr.Location.End.Column == p.next_location.End.Column &&
		BinaryPriority[BinaryOp_Or][0] > limit {
		p.nextLexeme()
		p.report(lex.Location{Begin: start.Begin, End: p.next_location.End}, "Unexpected '||'; did you mean 'or'?")
		op := BinaryOp_Or
		return &op
	} else if curr.Type == '!' && p.next_type == '=' &&
		curr.Location.End.Column == p.next_location.End.Column &&
		BinaryPriority[BinaryOp_CompareNe][0] > limit {
		p.nextLexeme()
		p.report(lex.Location{Begin: start.Begin, End: p.next_location.End}, "Unexpected '!='; did you mean '~='?")
		op := BinaryOp_CompareNe
		return &op
	}

	return nil
}

// parseExpr parses binary expressions at priority > limit
func (p *Parser) parseExpr(limit int) AstExpr {
	oldRecursion := p.recursionCounter
	p.incrementRecursionCounter("expression")

	start := p.snapshot()
	var expr AstExpr

	uop, hasUop := UnaryOpLookup[p.token_type]
	if !hasUop {
		if confusable := p.checkUnaryConfusables(); confusable != nil {
			uop = *confusable
			hasUop = true
		}
	}

	if hasUop {
		opPosition := p.token_location.Begin
		p.nextLexeme()

		subexpr := p.parseExpr(8)

		node := AstExprUnary{
			NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: subexpr.GetLocation().End}},
			Op:      uop,
			Expr:    subexpr,
		}

		if p.storeCstData {
			p.cstNodes[node] = CstExprOp{OpPosition: opPosition}
		}

		expr = node
	} else {
		expr = p.parseAssertionExpr()
	}

	op, hasOp := BinaryOpLookup[p.token_type]
	if !hasOp {
		if confusable := p.checkBinaryConfusables(limit); confusable != nil {
			op = *confusable
			hasOp = true
		}
	}

	for hasOp && BinaryPriority[op][0] > limit {
		opPosition := p.token_location.Begin
		p.nextLexeme()

		nextExpr := p.parseExpr(BinaryPriority[op][1])

		node := AstExprBinary{
			NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: nextExpr.GetLocation().End}},
			Op:      int(op),
			Left:    expr,
			Right:   nextExpr,
		}

		if p.storeCstData {
			p.cstNodes[node] = CstExprOp{OpPosition: opPosition}
		}

		expr = node

		op, hasOp = BinaryOpLookup[p.token_type]
		if !hasOp {
			if confusable := p.checkBinaryConfusables(limit); confusable != nil {
				op = *confusable
				hasOp = true
			}
		}

		p.incrementRecursionCounter("expression")
	}

	p.recursionCounter = oldRecursion
	return expr
}

// parseNameExpr parses a NAME reference, resolving to local, global, or error
func (p *Parser) parseNameExpr(context string) AstExprLocalOrGlobalOrError {
	nameOpt := p.parseNameOpt(&context)

	if nameOpt == nil {
		return AstExprError{
			NodeLoc:      &NodeLoc{p.snapshot()},
			Expressions:  nil,
			MessageIndex: len(p.parseErrors),
		}
	}

	name := nameOpt
	local_ := p.localMap[name.Name.Value]

	if local_ != nil {
		if local_.FunctionDepth < p.typeFunctionDepth {
			return p.reportExprError(p.snapshot(), nil, fmt.Sprintf("Type function cannot reference outer local '%s'", local_.Name))
		}

		return AstExprLocal{
			NodeLoc: &NodeLoc{name.Location},
			Local:   *local_,
			Upvalue: local_.FunctionDepth != len(p.functionStack)-1,
		}
	}

	return AstExprGlobal{
		NodeLoc: &NodeLoc{name.Location},
		Name:    name.Name.Value,
	}
}

// parsePrefixExpr parses NAME | `(' expr `)'
func (p *Parser) parsePrefixExpr() AstExpr {
	if p.token_type == '(' {
		start := p.token_location.Begin
		parenType := p.token_type
		parenBegin := p.token_location.Begin
		p.nextLexeme()

		expr := p.parseExpr(0)

		end := p.token_location.End
		if p.token_type != ')' {
			var extra string
			if p.token_type == '=' {
				extra = "; did you mean to use '{' when defining a table?"
			}
			p.expectMatchAndConsumeFail(')', parenType, parenBegin, extra)
			end = p.prev_location.End
		} else {
			p.nextLexeme()
		}

		return AstExprGroup{
			NodeLoc: &NodeLoc{lex.Location{Begin: start, End: end}},
			Expr:    expr,
		}
	}

	return p.parseNameExpr("expression")
}

// parseTypeInstantiationExpr parses `<<' type params `>>'
func (p *Parser) parseTypeInstantiationExpr() ([]AstTypeOrPack, CstTypeInstantiation) {
	leftArrow1 := p.token_location.Begin
	beginType := p.token_type
	beginPos := p.token_location.Begin
	p.nextLexeme()

	var leftArrow2 lex.Position
	var commaPositions []lex.Position
	var rightArrow1 lex.Position

	typesOrPacks := p.parseTypeParams(&leftArrow2, &commaPositions, &rightArrow1)

	rightArrow2 := p.token_location.Begin
	p.expectMatchAndConsume('>', beginType, beginPos, nil)

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
func (p *Parser) parseExplicitTypeInstantiationExpr(start lex.Position, basedOnExpr AstExpr) AstExprInstantiate {
	typesOrPacks, cstInstantiation := p.parseTypeInstantiationExpr()

	expr := AstExprInstantiate{
		NodeLoc:       &NodeLoc{lex.Location{Begin: start, End: p.prev_location.End}},
		Expr:          basedOnExpr,
		TypeArguments: typesOrPacks,
	}

	if p.storeCstData {
		p.cstNodes[expr] = CstExprExplicitTypeInstantiation{
			Instantiation: cstInstantiation,
		}
	}

	return expr
}

func (p *Parser) reportAmbiguousCallError() {
	p.report(p.snapshot(), "Ambiguous syntax: this looks like an argument list for a function call, but could also be a start of new statement; use ';' to separate statements")
}

// parsePrimaryExpr parses primary expression (field access, indexing, calls)
func (p *Parser) parsePrimaryExpr(asStatement bool) AstExpr {
	start := p.token_location.Begin
	expr := AstExpr(p.parsePrefixExpr())

	oldRecursion := p.recursionCounter

	for {
		if p.token_type == '.' {
			opPosition := p.token_location.Begin
			p.nextLexeme()

			ctx := "field name"
			index := p.parseIndexName(&ctx, opPosition)

			expr = AstExprIndexName{
				NodeLoc:       &NodeLoc{lex.Location{Begin: start, End: index.Location.End}},
				Expr:          expr,
				Index:         index.Name.Value,
				IndexLocation: index.Location,
				OpPosition:    opPosition,
				Op:            '.',
			}
		} else if p.token_type == '[' {
			bracketType := p.token_type
			bracketBegin := p.token_location.Begin
			openBracket := p.token_location.Begin
			p.nextLexeme()

			index := p.parseExpr(0)
			closeBracket := p.token_location.Begin
			p.expectMatchAndConsume(']', bracketType, bracketBegin, nil)

			e := AstExprIndexExpr{
				NodeLoc: &NodeLoc{lex.Location{Begin: start, End: p.prev_location.End}},
				Expr:    expr,
				Index:   index,
			}

			if p.storeCstData {
				p.cstNodes[e] = CstExprIndexExpr{
					OpenBracketPosition:  openBracket,
					CloseBracketPosition: closeBracket,
				}
			}

			expr = e
		} else if p.token_type == ':' {
			opPosition := p.token_location.Begin
			p.nextLexeme()

			ctx := "method name"
			index := p.parseIndexName(&ctx, opPosition)

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

				if p.token_type == '<' && p.next_type == '<' {
					args, cst := p.parseTypeInstantiationExpr()
					typeArgs = args
					cstInstantiation = &cst
				}

				callExpr := p.parseFunctionArgs(AstExpr(funcExpr), true)
				if len(typeArgs) > 0 {
					if ce, ok := callExpr.(AstExprCall); ok {
						ce.TypeArguments = &typeArgs
						callExpr = ce
					}
				}
				if p.storeCstData && cstInstantiation != nil {
					if ce, ok := callExpr.(AstExprCall); ok {
						if cstCall, ok2 := p.cstNodes[ce].(CstExprCall); ok2 {
							cstCall.ExplicitTypes = cstInstantiation
							p.cstNodes[ce] = cstCall
						}
					}
				}
				expr = callExpr
			} else {
				expr = p.parseFunctionArgs(AstExpr(funcExpr), true)
			}
		} else if p.token_type == '(' {
			if !asStatement && expr.GetLocation().End.Line != p.token_location.Begin.Line {
				p.reportAmbiguousCallError()
				break
			}
			expr = p.parseFunctionArgs(expr, false)
		} else if p.token_type == '{' || p.token_type == lex.RawString || p.token_type == lex.QuotedString {
			expr = p.parseFunctionArgs(expr, false)
		} else if LuauExplicitTypeInstantiationSyntax && p.token_type == '<' && p.next_type == '<' {
			expr = p.parseExplicitTypeInstantiationExpr(start, expr)
		} else {
			break
		}

		p.incrementRecursionCounter("expression")
	}

	p.recursionCounter = oldRecursion
	return expr
}

// parseAssertionExpr parses expr [`::' Type]
func (p *Parser) parseAssertionExpr() AstExpr {
	start := p.snapshot()
	expr := p.parseSimpleExpr()

	if p.token_type == lex.DoubleColon {
		opPos := p.token_location.Begin
		p.nextLexeme()
		annotation := p.parseType(false)
		annotLoc := annotation.GetLocation()

		node := AstExprTypeAssertion{
			NodeLoc:    &NodeLoc{lex.Location{Begin: start.Begin, End: annotLoc.End}},
			Expr:       expr,
			Annotation: annotation,
		}

		if p.storeCstData {
			p.cstNodes[node] = CstExprTypeAssertion{OpPosition: opPos}
		}

		return node
	}

	return expr
}

// parseSimpleExpr parses atoms: literals, `...', constructor, function, primary
func (p *Parser) parseSimpleExpr() AstExpr {
	start := p.snapshot()

	var attributes Attrs
	if p.token_type == lex.Attribute || p.token_type == lex.AttributeOpen {
		attributes = p.parseAttributes()

		if p.token_type != lex.ReservedFunction {
			currLex := lex.Lexeme{Type: p.token_type, Codepoint: p.token_codepoint}
			if p.token_string != nil {
				currLex.Data = []byte(*p.token_string)
			}
			return p.reportExprError(start, nil, fmt.Sprintf("Expected 'function' declaration after attribute, but got %s instead", currLex.String()))
		}
	}

	switch p.token_type {
	case lex.ReservedNil:
		p.nextLexeme()
		return AstExprConstantNil{NodeLoc: &NodeLoc{start}}
	case lex.ReservedTrue:
		p.nextLexeme()
		return AstExprConstantBool{NodeLoc: &NodeLoc{start}, Value: true}
	case lex.ReservedFalse:
		p.nextLexeme()
		return AstExprConstantBool{NodeLoc: &NodeLoc{start}, Value: false}
	case lex.ReservedFunction:
		matchFunction := p.get_lexeme()
		p.nextLexeme()
		node, _ := p.parseFunctionBody(false, matchFunction, nil, nil, attributes, false)
		return node
	case lex.Number:
		return p.parseNumber()
	case lex.RawString, lex.QuotedString, lex.InterpStringSimple:
		return p.parseString()
	case lex.InterpStringBegin:
		return p.parseInterpString()
	case lex.BrokenString:
		p.nextLexeme()
		return p.reportExprError(start, nil, "Malformed string; did you forget to finish it?")
	case lex.BrokenInterpDoubleBrace:
		p.nextLexeme()
		return p.reportExprError(start, nil, "Double braces are not permitted within interpolated strings; did you mean '\\{'?")
	case lex.Dot3:
		if len(p.functionStack) > 0 && p.functionStack[len(p.functionStack)-1].Vararg {
			p.nextLexeme()
			return AstExprVarargs{NodeLoc: &NodeLoc{start}}
		}
		p.nextLexeme()
		return p.reportExprError(start, nil, "Cannot use '...' outside of a vararg function")
	case '{':
		return p.parseTableConstructor()
	case lex.ReservedIf:
		return p.parseIfElseExpr()
	default:
		return p.parsePrimaryExpr(false)
	}
}

// parseFunctionArgs parses `(' [explist] `)' | tableconstructor | String
func (p *Parser) parseFunctionArgs(funcExpr AstExpr, selfCall bool) AstExpr {
	switch p.token_type {
	case '(':
		if funcExpr.GetLocation().End.Line != p.token_location.Begin.Line {
			p.reportAmbiguousCallError()
		}

		argStart := p.token_location.End
		parenType := p.token_type
		parenBegin := p.token_location.Begin
		p.nextLexeme()

		var args []AstExpr
		var commaPositions []lex.Position
		if p.token_type != ')' {
			p.parseExprList(&args, &commaPositions)
		}

		closeParen := p.token_location.Begin
		end := p.snapshot()
		p.expectMatchAndConsume(')', parenType, parenBegin, nil)

		result := AstExprCall{
			NodeLoc:     &NodeLoc{lex.Location{Begin: funcExpr.GetLocation().Begin, End: end.End}},
			Func:        funcExpr,
			Args:        args,
			Self:        selfCall,
			ArgLocation: lex.Location{Begin: argStart, End: end.End},
		}

		if p.storeCstData {
			p.cstNodes[result] = CstExprCall{
				OpenParens:     &parenBegin,
				CloseParens:    &closeParen,
				CommaPositions: commaPositions,
			}
		}

		return result

	case '{':
		argStart := p.token_location.End
		tableExpr := p.parseTableConstructor()
		argEnd := p.prev_location.End

		result := AstExprCall{
			NodeLoc:     &NodeLoc{lex.Location{Begin: funcExpr.GetLocation().Begin, End: tableExpr.GetLocation().End}},
			Func:        funcExpr,
			Args:        []AstExpr{tableExpr},
			Self:        selfCall,
			ArgLocation: lex.Location{Begin: argStart, End: argEnd},
		}

		if p.storeCstData {
			p.cstNodes[result] = CstExprCall{CommaPositions: []lex.Position{}}
		}

		return result

	case lex.RawString, lex.QuotedString:
		argLocation := p.snapshot()
		strExpr := p.parseString()

		result := AstExprCall{
			NodeLoc:     &NodeLoc{lex.Location{Begin: funcExpr.GetLocation().Begin, End: strExpr.GetLocation().End}},
			Func:        funcExpr,
			Args:        []AstExpr{strExpr},
			Self:        selfCall,
			ArgLocation: argLocation,
		}

		if p.storeCstData {
			p.cstNodes[result] = CstExprCall{CommaPositions: []lex.Position{}}
		}

		return result
	}

	return p.reportFunctionArgsError(funcExpr, selfCall)
}

// reportFunctionArgsError reports error for bad function call syntax
func (p *Parser) reportFunctionArgsError(funcExpr AstExpr, selfCall bool) AstExpr {
	if selfCall && p.token_location.Begin.Line != funcExpr.GetLocation().End.Line {
		return p.reportExprError(funcExpr.GetLocation(), []AstExpr{funcExpr}, "Expected function call arguments after '('")
	}

	currLex := lex.Lexeme{Type: p.token_type, Codepoint: p.token_codepoint}
	if p.token_string != nil {
		currLex.Data = []byte(*p.token_string)
	}
	return p.reportExprError(
		lex.Location{Begin: funcExpr.GetLocation().Begin, End: p.token_location.Begin},
		[]AstExpr{funcExpr},
		fmt.Sprintf("Expected '(', '{' or <string> when parsing function call, got %s", currLex.String()),
	)
}

// parseIndexName parses a field name, accepting keywords if on same line
func (p *Parser) parseIndexName(context *string, prev lex.Position) Binding {
	nameOpt := p.parseNameOpt(context)
	if nameOpt != nil {
		return *nameOpt
	}

	if p.token_type >= lex.Reserved_BEGIN && p.token_type < lex.Reserved_END &&
		p.token_location.Begin.Line == prev.Line {
		nameStr := ""
		if p.token_string != nil {
			nameStr = *p.token_string
		}
		result := Binding{
			Name:    lex.AstName{Value: nameStr},
			NodeLoc: &NodeLoc{p.snapshot()},
		}

		p.nextLexeme()
		return result
	}

	return Binding{
		Name:    lex.AstName{Value: nameError},
		NodeLoc: &NodeLoc{p.snapshot()},
	}
}

// parseCallList parses function call arguments (used by intepstring etc.)
func (p *Parser) parseCallList(commaPositions *[]lex.Position) ([]AstExpr, lex.Location, lex.Location) {
	switch p.token_type {
	case '(':
		argStart := p.token_location.End
		parenType := p.token_type
		parenBegin := p.token_location.Begin

		p.nextLexeme()

		var args []AstExpr

		if p.token_type != ')' {
			p.parseExprList(&args, commaPositions)
		}

		end := p.snapshot()
		p.expectMatchAndConsume(')', parenType, parenBegin, nil)

		return args,
			lex.Location{Begin: argStart, End: end.End},
			lex.Location{Begin: parenBegin, End: p.prev_location.End}

	case '{':
		argStart := p.token_location.End
		expr := p.parseTableConstructor()

		return []AstExpr{expr},
			lex.Location{Begin: argStart, End: p.prev_location.End},
			expr.GetLocation()
	}

	argLoc := p.snapshot()
	expr := p.parseString()
	return []AstExpr{expr}, argLoc, expr.GetLocation()
}

// parseTableConstructor parses `{' [fieldlist] `}'
func (p *Parser) parseTableConstructor() AstExprTable {
	var items []AstExprTableItem
	var cstItems []CstExprTableItem

	start := p.snapshot()

	braceType := p.token_type
	braceBegin := p.token_location.Begin
	p.expectAndConsume('{', new("table literal"))

	lastElementIndent := uint32(0)

	for p.token_type != '}' {
		lastElementIndent = p.token_location.Begin.Column

		if p.token_type == '[' {
			indexerOpenPos := p.token_location.Begin
			bracketType := p.token_type
			bracketBegin := p.token_location.Begin
			p.nextLexeme()

			key := p.parseExpr(0)

			indexerClosePos := p.token_location.Begin
			p.expectMatchAndConsume(']', bracketType, bracketBegin, nil)

			equalsPos := p.token_location.Begin
			ctx := "table field"
			p.expectAndConsume('=', &ctx)

			value := p.parseExpr(0)

			items = append(items, AstExprTableItem{
				NodeLoc: &NodeLoc{lex.Location{}},
				Kind:    General,
				Key:     &key,
				Value:   value,
			})

			if p.storeCstData {
				sepPos := p.token_location.Begin
				cstItems = append(cstItems, CstExprTableItem{
					Kind:                 General,
					IndexerOpenPosition:  &indexerOpenPos,
					IndexerClosePosition: &indexerClosePos,
					EqualsPosition:       &equalsPos,
					Separator:            p.tableSeparator(),
					SeparatorPosition:    sepPos,
				})
			}
		} else if p.token_type == lex.Name && p.next_type == '=' {
			ctx := "table field"
			name := p.parseName(&ctx)

			equalsPos := p.token_location.Begin
			ctx2 := "table field"
			p.expectAndConsume('=', &ctx2)

			keyExpr := AstExpr(AstExprConstantString{
				NodeLoc: &NodeLoc{name.Location},
				Value:   name.Name.Value,
			})

			value := p.parseExpr(0)

			if fe, ok := value.(AstExprFunction); ok {
				fe.Debugname = name.Name.Value
				value = fe
			}

			items = append(items, AstExprTableItem{
				NodeLoc: &NodeLoc{lex.Location{}},
				Kind:    Record,
				Key:     &keyExpr,
				Value:   value,
			})

			if p.storeCstData {
				sepPos := p.token_location.Begin
				cstItems = append(cstItems, CstExprTableItem{
					Kind:              Record,
					EqualsPosition:    &equalsPos,
					Separator:         p.tableSeparator(),
					SeparatorPosition: sepPos,
				})
			}
		} else {
			expr := p.parseExpr(0)
			items = append(items, AstExprTableItem{
				NodeLoc: &NodeLoc{lex.Location{}},
				Kind:    List,
				Value:   expr,
			})

			if p.storeCstData {
				sepPos := p.token_location.Begin
				cstItems = append(cstItems, CstExprTableItem{
					Kind:              List,
					Separator:         p.tableSeparator(),
					SeparatorPosition: sepPos,
				})
			}
		}

		if p.token_type == ',' || p.token_type == ';' {
			p.nextLexeme()
		} else if (p.token_type == '[' || p.token_type == lex.Name) && p.token_location.Begin.Column == lastElementIndent {
			p.report(p.snapshot(), "Expected ',' after table constructor element")
		} else if p.token_type != '}' {
			break
		}
	}

	end := p.snapshot()
	if !p.expectMatchAndConsume('}', braceType, braceBegin, nil) {
		end = p.getprev()
	}

	node := AstExprTable{
		NodeLoc: &NodeLoc{lex.Location{Begin: start.Begin, End: end.End}},
		Items:   items,
	}

	if p.storeCstData {
		p.cstNodes[node] = CstExprTable{Items: cstItems}
	}

	return node
}

// parseIfElseExpr parses if-then-else expression
func (p *Parser) parseIfElseExpr() AstExprIfElse {
	start := p.snapshot()
	p.nextLexeme() // consume 'if' or 'elseif'

	if p.token_type == lex.ReservedLocal {
		return p.parseIfElseExprLocalCondition(start)
	}

	if p.token_type == lex.Name && p.token_string != nil && *p.token_string == "const" && p.next_type == lex.Name {
		return p.parseIfElseExprLocalCondition(start)
	}

	condition := p.parseExpr(0)

	thenPosition := p.token_location.Begin
	hasThen := p.expectAndConsume(lex.ReservedThen, nil)

	trueExpr := p.parseExpr(0)

	elsePosition := p.token_location.Begin
	falseExpr, hasElse, isElseIf := p.parseIfElseExprTail()

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

	if p.storeCstData {
		p.cstNodes[node] = CstExprIfElse{
			ThenPosition: thenPosition,
			ElsePosition: elsePosition,
			IsElseIf:     isElseIf,
		}
	}

	return node
}

// parseIfElseExprLocalCondition parses the `if local x = e then a else b` (and
// `const`) expression form (LuauExperimentalIfLocalSyntax).
func (p *Parser) parseIfElseExprLocalCondition(start lex.Location) AstExprIfElse {
	condIsConst := p.token_type == lex.Name && p.token_string != nil && *p.token_string == "const"

	keywordLocation := p.snapshot()
	p.nextLexeme() // consume 'local' or 'const'

	binding := p.parseBinding(condIsConst)

	if p.token_type == ',' {
		p.report(p.token_location, "Expected '=' after variable name in 'if local', got ','; only a single binding is allowed")
	}

	var equalsPosition *lex.Location
	if p.token_type == '=' {
		loc := p.snapshot()
		equalsPosition = &loc
	}

	p.expectAndConsume('=', new("if local declaration"))

	condition := p.parseExpr(0)

	thenPosition := p.token_location.Begin
	hasThen := p.expectAndConsume(lex.ReservedThen, new("if then else expression"))

	// Push the binding after the condition so the condition cannot reference it,
	// and restore after the true expression so it isn't visible in else/elseif.
	localsBegin := len(p.localStack)
	condLocal := p.pushLocal(binding)

	trueExpr := p.parseExpr(0)

	p.restoreLocals(localsBegin)

	elsePosition := p.token_location.Begin
	falseExpr, hasElse, isElseIf := p.parseIfElseExprTail()

	var falseEnd lex.Position
	if falseExpr != nil {
		falseEnd = falseExpr.GetLocation().End
	}

	node := AstExprIfElse{
		NodeLoc:                  &NodeLoc{lex.Location{Begin: start.Begin, End: falseEnd}},
		Condition:                condition,
		HasThen:                  hasThen,
		TrueExpr:                 trueExpr,
		HasElse:                  hasElse,
		FalseExpr:                falseExpr,
		ConditionLocal:           condLocal,
		ConditionIsConst:         condIsConst,
		ConditionKeywordLocation: &keywordLocation,
		ConditionEqualsLocation:  equalsPosition,
	}

	if p.storeCstData {
		p.cstNodes[node] = CstExprIfElse{
			ThenPosition: thenPosition,
			ElsePosition: elsePosition,
			IsElseIf:     isElseIf,
		}
	}

	return node
}

// parseIfElseExprTail parses the `elseif ...`/`else ...` part of an if-expression.
func (p *Parser) parseIfElseExprTail() (falseExpr AstExpr, hasElse bool, isElseIf bool) {
	if p.token_type == lex.ReservedElseif {
		oldRecursion := p.recursionCounter
		p.incrementRecursionCounter("expression")
		hasElse = true
		result := p.parseIfElseExpr()
		falseExpr = result
		p.recursionCounter = oldRecursion
		isElseIf = true
	} else {
		hasElse = p.expectAndConsume(lex.ReservedElse, nil)
		falseExpr = p.parseExpr(0)
	}

	return falseExpr, hasElse, isElseIf
}

// parseInterpString parses an interpolated string expression
func (p *Parser) parseInterpString() AstExprInterpStringOrError {
	var strs []string
	var sourceStrings []string
	var stringPositions []lex.Position
	var expressions []AstExpr

	startLocation := p.snapshot()
	var endLocation lex.Location

	for {
		currentLexeme := p.get_lexeme()
		endLocation = currentLexeme.Location

		data := ""
		if p.token_string != nil {
			data = *p.token_string
		}

		if p.storeCstData {
			sourceStrings = append(sourceStrings, data)
			stringPositions = append(stringPositions, currentLexeme.Location.Begin)
		}

		ok, fixedData := p.lexer.FixupQuotedString([]byte(data))
		if !ok {
			p.nextLexeme()
			return p.reportExprError(
				lex.Location{Begin: startLocation.Begin, End: endLocation.End},
				nil,
				"Interpolated string literal contains malformed escape sequence",
			)
		}

		p.nextLexeme()
		strs = append(strs, string(fixedData))

		if currentLexeme.Type == lex.InterpStringEnd || currentLexeme.Type == lex.InterpStringSimple {
			break
		}

		t := p.token_type

		if t == lex.InterpStringMid || t == lex.InterpStringEnd {
			p.nextLexeme()
			expressions = append(expressions, p.reportExprError(endLocation, nil, "Malformed interpolated string, expected expression inside '{}'"))
			break
		} else if t == lex.BrokenString {
			p.nextLexeme()
			expressions = append(expressions, p.reportExprError(endLocation, nil, "Malformed interpolated string; did you forget to add a '`'?"))
			break
		} else {
			expressions = append(expressions, p.parseExpr(0))
		}

		switch t = p.token_type; t {
		case lex.InterpStringBegin, lex.InterpStringMid, lex.InterpStringEnd:
			// continue reading

		case lex.BrokenInterpDoubleBrace:
			p.nextLexeme()
			return p.reportExprError(endLocation, nil, "Double braces are not permitted within interpolated strings; did you mean '\\{'?")

		case lex.BrokenString, lex.Eof:
			if t == lex.BrokenString {
				p.nextLexeme()
			}

			node := AstExprInterpString{
				NodeLoc:     &NodeLoc{lex.Location{Begin: startLocation.Begin, End: p.prev_location.End}},
				Strings:     strs,
				Expressions: expressions,
			}

			if p.storeCstData {
				p.cstNodes[node] = CstExprInterpString{
					SourceStrings:   sourceStrings,
					StringPositions: stringPositions,
				}
			}

			if len(p.braceStack) > 0 && p.braceStack[len(p.braceStack)-1] == lex.InterpolatedString {
				p.report(p.getprev(), "Malformed interpolated string; did you forget to add a '}'?")
			} else {
				p.report(p.getprev(), "Malformed interpolated string; did you forget to add a '`'?")
			}

			return node

		default:
			currLex := lex.Lexeme{Type: p.token_type, Codepoint: p.token_codepoint}
			if p.token_string != nil {
				currLex.Data = []byte(*p.token_string)
			}
			return p.reportExprError(endLocation, nil, fmt.Sprintf("Malformed interpolated string, got %s", currLex.String()))
		}
	}

	node := AstExprInterpString{
		NodeLoc:     &NodeLoc{lex.Location{Begin: startLocation.Begin, End: endLocation.End}},
		Strings:     strs,
		Expressions: expressions,
	}

	if p.storeCstData {
		p.cstNodes[node] = CstExprInterpString{
			SourceStrings:   sourceStrings,
			StringPositions: stringPositions,
		}
	}

	return node
}

// parseCharArray parses string token and returns unescaped bytes, or nil on error
func (p *Parser) parseCharArray() *string {
	t := p.token_type
	// fmt.Println("Parsing char array token type", lex.Lexeme{Type: t}.String())
	data := ""
	if p.token_string != nil {
		data = *p.token_string
	}
	// fmt.Println("Parsing char array data", data)

	var result string

	if t == lex.QuotedString || t == lex.InterpStringSimple {
		ok, fixed := p.lexer.FixupQuotedString([]byte(data))
		if !ok {
			// fmt.Println("Failed to fixup quoted string")
			p.nextLexeme()
			return nil
		}
		result = string(fixed)
	} else {
		result = string(p.lexer.FixupMultilineString([]byte(data)))
	}

	p.nextLexeme()
	return &result
}

// parseString parses a string literal expression
func (p *Parser) parseString() AstExprConstantStringOrError {
	location := p.snapshot()
	quoteStyle := QuoteStyle_QuotedSimple

	switch p.token_type {
	case lex.QuotedString:
		if p.token_aux != nil && *p.token_aux == 0 {
			quoteStyle = QuoteStyle_QuotedSingle
			break
		}
		quoteStyle = QuoteStyle_QuotedSimple

	case lex.InterpStringSimple:
		quoteStyle = QuoteStyle_QuotedSimple

	case lex.RawString:
		quoteStyle = QuoteStyle_QuotedRaw
	}

	var fullStyle CstQuotes
	var blockDepth int
	if p.storeCstData {
		fullStyle, blockDepth = p.extractStringDetails()
	}

	var originalString *string
	if p.storeCstData {
		originalString = p.token_string
	}

	value := p.parseCharArray()

	if value != nil {
		node := AstExprConstantString{
			NodeLoc:    &NodeLoc{location},
			Value:      *value,
			QuoteStyle: quoteStyle,
		}

		if p.storeCstData {
			p.cstNodes[node] = CstExprConstantString{
				SourceString: originalString,
				QuoteStyle:   int(fullStyle),
				BlockDepth:   blockDepth,
			}
		}

		return node
	}

	return p.reportExprError(location, nil, "String literal contains malformed escape sequence")
}

type NumberParseResult uint8

const (
	NumberParseResult_Ok NumberParseResult = iota
	NumberParseResult_Malformed
	NumberParseResult_Imprecise
	NumberParseResult_BinOverflow
	NumberParseResult_HexOverflow
	NumberParseResult_IntOverflow
)

// parseNumber parses a number literal expression
func (p *Parser) parseNumber() AstExprConstantNumberOrError {
	start := p.snapshot()
	data := ""
	if p.token_string != nil {
		data = *p.token_string
	}

	var sourceData string
	if p.storeCstData {
		sourceData = data
	}

	cleanData := strings.ReplaceAll(data, "_", "")

	// Integer literal with an `i` suffix, e.g. `1i` or `0xFFi`
	if strings.HasSuffix(cleanData, "i") {
		integerData := cleanData[:len(cleanData)-1]
		var intValue int64
		var intResult NumberParseResult

		switch {
		case strings.HasPrefix(integerData, "0x") || strings.HasPrefix(integerData, "0X"):
			u, err := strconv.ParseUint(integerData[2:], 16, 64)
			switch {
			case err == nil:
				intValue = int64(u)
			case errors.Is(err, strconv.ErrRange):
				intResult = NumberParseResult_HexOverflow
			default:
				intResult = NumberParseResult_Malformed
			}
		case strings.HasPrefix(integerData, "0b") || strings.HasPrefix(integerData, "0B"):
			u, err := strconv.ParseUint(integerData[2:], 2, 64)
			switch {
			case err == nil:
				intValue = int64(u)
			case errors.Is(err, strconv.ErrRange):
				intResult = NumberParseResult_BinOverflow
			default:
				intResult = NumberParseResult_Malformed
			}
		default:
			v, err := strconv.ParseInt(integerData, 10, 64)
			switch {
			case err == nil:
				intValue = v
			case errors.Is(err, strconv.ErrRange):
				intResult = NumberParseResult_IntOverflow
			default:
				intResult = NumberParseResult_Malformed
			}
		}

		p.nextLexeme()

		if intResult == NumberParseResult_Malformed {
			return p.reportExprError(start, nil, "Malformed integer")
		}
		if intResult != NumberParseResult_Ok {
			return p.reportExprError(start, nil, "Integer overflow")
		}

		node := AstExprConstantInteger{
			NodeLoc:     &NodeLoc{start},
			Value:       intValue,
			ParseResult: intResult,
		}

		if p.storeCstData {
			p.cstNodes[node] = CstExprConstantInteger{Value: sourceData}
		}

		return node
	}

	value := 0.0
	var parseResult NumberParseResult

	// Hexadecimal check (0x...)
	if strings.HasPrefix(cleanData, "0x") || strings.HasPrefix(cleanData, "0X") {
		// v, err := strconv.ParseUint(cleanData[2:], 16, 64)
		// if err != nil {
		// 	malformed = true
		// } else {
		// 	value = float64(v)
		// }
		content := cleanData[2:]
		significant := "0"
		// get all characters after the leading zeros
		for i, c := range content {
			if c != '0' {
				significant = content[i:]
				break
			}
		}

		if len(significant) > 16 {
			parseResult = NumberParseResult_HexOverflow
			value = 0
		} else {
			v, err := strconv.ParseUint(content, 16, 64)
			if err != nil {
				parseResult = NumberParseResult_Malformed
				value = 0
			} else {
				value = float64(v)
				if v >= 9007199254740992 { // 2^53
					parseResult = NumberParseResult_Imprecise
				}
			}
		}
	} else if strings.HasPrefix(cleanData, "0b") || strings.HasPrefix(cleanData, "0B") {
		// v, err := strconv.ParseUint(cleanData[2:], 2, 64)
		// if err != nil {
		// 	malformed = true
		// } else {
		// 	value = float64(v)
		// }
		content := cleanData[2:]
		significant := "0"
		// get all characters after the leading zeros
		for i, c := range content {
			if c != '0' {
				significant = content[i:]
				break
			}
		}

		if len(significant) > 64 {
			parseResult = NumberParseResult_BinOverflow
			value = 0
		} else {
			v, err := strconv.ParseUint(content, 2, 64)
			if err != nil {
				parseResult = NumberParseResult_Malformed
				value = 0
			} else {
				value = float64(v)
				if v >= 9007199254740992 { // 2^53
					parseResult = NumberParseResult_Imprecise
				}
			}
		}
	} else {
		v, err := strconv.ParseFloat(cleanData, 64)
		if err != nil && !errors.Is(err, strconv.ErrRange) { // wdc about value out of range
			parseResult = NumberParseResult_Malformed
			// fmt.Println("Marked number as malformed:", v, cleanData, err)
			value = 0
		} else {
			value = v

			if value >= 9007199254740992 {
				isAllDigits := true
				for _, c := range cleanData {
					if c < '0' || c > '9' {
						isAllDigits = false
						break
					}
				}

				if isAllDigits {
					repr := fmt.Sprintf("%.0f", value)
					if repr != cleanData {
						parseResult = NumberParseResult_Imprecise
					}
				}
			}
		}
	}

	p.nextLexeme()

	if parseResult == NumberParseResult_Malformed {
		return p.reportExprError(start, nil, "Malformed number")
	}

	node := AstExprConstantNumber{
		NodeLoc: &NodeLoc{start},
		Value:   value,
	}

	if p.storeCstData {
		p.cstNodes[node] = CstExprConstantNumber{Value: sourceData}
	}

	return node
}
