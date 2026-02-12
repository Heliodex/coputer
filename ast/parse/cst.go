package main

import "github.com/Heliodex/coputer/ast/lex"

// --------------------------------------------------------------------------------
// -- CST NODE UNION TYPES
// --------------------------------------------------------------------------------

// Union type representing all possible Concrete Syntax Tree nodes.
// CST nodes store positional data for syntactic elements (commas, keywords, etc.)
// that are not semantically significant but needed for source-accurate transformations.

type CstNode interface {
	isCstNode()
}

var (
	_ CstNode = CstStatBlock{}
	_ CstNode = CstStatRepeat{}
	_ CstNode = CstStatDo{}
	_ CstNode = CstStatFor{}
	_ CstNode = CstStatForIn{}
	_ CstNode = CstStatFunction{}
	_ CstNode = CstStatLocalFunction{}
	_ CstNode = CstStatLocal{}
	_ CstNode = CstStatAssign{}
	_ CstNode = CstStatCompoundAssign{}
	_ CstNode = CstStatReturn{}
	_ CstNode = CstStatTypeAlias{}
	_ CstNode = CstStatTypeFunction{}
	_ CstNode = CstExprFunction{}
	_ CstNode = CstExprTable{}
	_ CstNode = CstExprIfElse{}
	_ CstNode = CstExprInterpString{}
	_ CstNode = CstExprConstantString{}
	_ CstNode = CstExprConstantNumber{}
	_ CstNode = CstExprOp{}
	_ CstNode = CstTypeTypeof{}
	_ CstNode = CstTypeReference{}
	_ CstNode = CstTypePackGeneric{}
	_ CstNode = CstTypePackExplicit{}
	_ CstNode = CstTypeFunction{}
	_ CstNode = CstTypeUnion{}
	_ CstNode = CstTypeIntersection{}
	_ CstNode = CstTypeTable{}
	_ CstNode = CstGenericType{}
	_ CstNode = CstGenericTypePack{}
	_ CstNode = CstExprCall{}
	_ CstNode = CstExprIndexExpr{}
	_ CstNode = CstExprTypeAssertion{}
	_ CstNode = CstExprExplicitTypeInstantiation{}
	_ CstNode = CstTypeSingletonString{}
)

type CstStatBlock struct {
	BodyCommaPositions []lex.Position
}

func (CstStatBlock) isCstNode() {}

type CstStatRepeat struct {
	UntilPosition lex.Position
}

func (CstStatRepeat) isCstNode() {}

type CstStatDo struct {
	StatsStart  lex.Position
	EndPosition lex.Position
}

func (CstStatDo) isCstNode() {}

// type CstStatDo_DEPRECATED struct {
// 	EndPosition lex.Position
// }

// func (CstStatDo_DEPRECATED) isCstNode() {}

type CstStatFor struct {
	AnnotationColonPosition *lex.Position
	EqualsPosition          lex.Position
	EndCommaPosition        lex.Position
	StepCommaPosition       *lex.Position
}

func (CstStatFor) isCstNode() {}

type CstStatForIn struct {
	VarsAnnotationColonPositions []*lex.Position
	VarsCommaPositions           []lex.Position
	ValuesCommaPositions         []lex.Position
}

func (CstStatForIn) isCstNode() {}

type CstStatFunction struct {
	FunctionKeywordPosition lex.Position
}

func (CstStatFunction) isCstNode() {}

type CstStatLocalFunction struct {
	LocalKeywordPosition    lex.Position
	FunctionKeywordPosition lex.Position
}

func (CstStatLocalFunction) isCstNode() {}

type CstStatLocal struct {
	VarsAnnotationColonPositions []*lex.Position
	VarsCommaPositions           []lex.Position
	ValuesCommaPositions         []lex.Position
}

func (CstStatLocal) isCstNode() {}

type CstStatAssign struct {
	VarsCommaPositions   []lex.Position
	EqualsPosition       lex.Position
	ValuesCommaPositions []lex.Position
}

func (CstStatAssign) isCstNode() {}

type CstStatCompoundAssign struct {
	OpPosition lex.Position
}

func (CstStatCompoundAssign) isCstNode() {}

type CstStatReturn struct {
	CommaPositions []lex.Position
}

func (CstStatReturn) isCstNode() {}

type CstStatTypeAlias struct {
	TypeKeywordPosition    lex.Position
	GenericsOpenPosition   *lex.Position
	GenericsCommaPositions []lex.Position
	GenericsClosePosition  *lex.Position
	EqualsPosition         lex.Position
}

func (CstStatTypeAlias) isCstNode() {}

type CstStatTypeFunction struct {
	TypeKeywordPosition     lex.Position
	FunctionKeywordPosition lex.Position
}

func (CstStatTypeFunction) isCstNode() {}

type CstExprFunction struct {
	FunctionKeywordPosition       lex.Position
	OpenGenericsPosition          *lex.Position
	GenericsCommaPositions        []lex.Position
	CloseGenericsPosition         *lex.Position
	ArgsAnnotationColonPositions  []*lex.Position
	ArgsCommaPositions            []lex.Position
	VarargAnnotationColonPosition *lex.Position
	ReturnSpecifierPosition       *lex.Position
}

func (CstExprFunction) isCstNode() {}

type CstExprTable struct {
	Items []CstExprTableItem
}

func (CstExprTable) isCstNode() {}

type CstExprTableItem struct {
	Kind                 string
	EqualsPosition       *lex.Position
	Separator            rune
	SeparatorPosition    lex.Position
	IndexerOpenPosition  *lex.Position
	IndexerClosePosition *lex.Position
}

func (CstExprTableItem) isCstNode() {}

type CstExprIfElse struct {
	ThenPosition lex.Position
	ElsePosition lex.Position
	IsElseIf     bool
}

func (CstExprIfElse) isCstNode() {}

type CstExprInterpString struct {
	SourceStrings   []string
	StringPositions []lex.Position
}

func (CstExprInterpString) isCstNode() {}

type CstExprConstantString struct {
	SourceString *string
	QuoteStyle   int
	BlockDepth   int
}

func (CstExprConstantString) isCstNode() {}

type CstTypeInstantiation struct {
	LeftArrow1Position lex.Position
	LeftArrow2Position lex.Position

	RightArrow1Position lex.Position
	RightArrow2Position lex.Position
	CommaPositions      []lex.Position
}

func (CstTypeInstantiation) isCstNode() {}

type CstExprCall struct {
	OpenParens     *lex.Position
	CloseParens    *lex.Position
	CommaPositions []lex.Position
	ExplicitTypes  *CstTypeInstantiation
}

func (CstExprCall) isCstNode() {}

type CstExprIndexExpr struct {
	OpenBracketPosition  lex.Position
	CloseBracketPosition lex.Position
}

func (CstExprIndexExpr) isCstNode() {}

type CstExprTypeAssertion struct {
	OpPosition lex.Position
}

func (CstExprTypeAssertion) isCstNode() {}

type CstExprExplicitTypeInstantiation struct {
	Instantiation CstTypeInstantiation
}

func (CstExprExplicitTypeInstantiation) isCstNode() {}

type CstExprConstantNumber struct {
	Value string
}

func (CstExprConstantNumber) isCstNode() {}

type CstExprOp struct {
	OpPosition lex.Position
}

func (CstExprOp) isCstNode() {}

type CstTypeTypeof struct {
	OpenPosition  lex.Position
	ClosePosition lex.Position
}

func (CstTypeTypeof) isCstNode() {}

type CstTypeReference struct {
	PrefixPointPosition      *lex.Position
	OpenParametersPosition   *lex.Position
	ParametersCommaPositions []lex.Position
	CloseParametersPosition  *lex.Position
}

func (CstTypeReference) isCstNode() {}

type CstTypePackGeneric struct {
	EllipsisPosition lex.Position
}

func (CstTypePackGeneric) isCstNode() {}

type CstTypePackExplicit struct {
	OpenParenthesesPosition  *lex.Position
	CloseParenthesesPosition *lex.Position
	CommaPositions           *[]lex.Position
}

func (CstTypePackExplicit) isCstNode() {}

type CstTypeFunction struct {
	OpenGenericsPosition       *lex.Position
	GenericsCommaPositions     []lex.Position
	CloseGenericsPosition      *lex.Position
	OpenArgsPosition           lex.Position
	ArgumentNameColonPositions []*lex.Position
	ArgumentsCommaPositions    []lex.Position
	CloseArgsPosition          lex.Position
	ReturnArrowPosition        lex.Position
}

func (CstTypeFunction) isCstNode() {}

type CstTypeUnion struct {
	LeadingPosition    *lex.Position
	SeparatorPositions []lex.Position
}

func (CstTypeUnion) isCstNode() {}

type CstTypeIntersection struct {
	LeadingPosition    *lex.Position
	SeparatorPositions []lex.Position
}

func (CstTypeIntersection) isCstNode() {}

type CstTypeTable struct {
	Items   []CstTypeTableItem
	IsArray bool
}

func (CstTypeTable) isCstNode() {}

type CstTypeTableItem struct {
	Kind                 string // "Property" | "StringProperty" | "Indexer"
	ColonPosition        *lex.Position
	Separator            rune
	SeparatorPosition    *lex.Position
	IndexerOpenPosition  *lex.Position
	IndexerClosePosition *lex.Position
	StringInfo           *CstExprConstantString
	StringPosition       *lex.Location
	EqualsPosition       *lex.Position
}

func (CstTypeTableItem) isCstNode() {}

type CstGenericType struct {
	DefaultEqualsPosition *lex.Position
}

func (CstGenericType) isCstNode() {}

type CstGenericTypePack struct {
	EllipsisPosition      lex.Position
	DefaultEqualsPosition *lex.Position
}

func (CstGenericTypePack) isCstNode() {}

type CstTypeSingletonString struct {
	SourceString []string
	QuoteStyle   int
	BlockDepth   int
}

func (CstTypeSingletonString) isCstNode() {}
