package main

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
	_ CstNode = CstExprExplicitTypeIns{}
	_ CstNode = CstTypeSingletonString{}
)

type CstStatBlock struct {
	BodyCommaPositions []Position
}

func (CstStatBlock) isCstNode() {}

type CstStatRepeat struct {
	UntilPosition Position
}

func (CstStatRepeat) isCstNode() {}

type CstStatDo struct {
	EndPosition Position
}

func (CstStatDo) isCstNode() {}

type CstStatFor struct {
	AnnotationColonPosition *Position
	EqualsPosition          Position
	EndCommaPosition        Position
	StepCommaPosition       *Position
}

func (CstStatFor) isCstNode() {}

type CstStatForIn struct {
	VarsAnnotationColonPositions []*Position
	VarsCommaPositions           []Position
	ValuesCommaPositions         []Position
}

func (CstStatForIn) isCstNode() {}

type CstStatFunction struct {
	FunctionKeywordPosition Position
}

func (CstStatFunction) isCstNode() {}

type CstStatLocalFunction struct {
	LocalKeywordPosition    Position
	FunctionKeywordPosition Position
}

func (CstStatLocalFunction) isCstNode() {}

type CstStatLocal struct {
	VarsAnnotationColonPositions []*Position
	VarsCommaPositions           []Position
	ValuesCommaPositions         []Position
}

func (CstStatLocal) isCstNode() {}

type CstStatAssign struct {
	VarsCommaPositions   []Position
	EqualsPosition       Position
	ValuesCommaPositions []Position
}

func (CstStatAssign) isCstNode() {}

type CstStatCompoundAssign struct {
	OpPosition Position
}

func (CstStatCompoundAssign) isCstNode() {}

type CstStatReturn struct {
	CommaPositions []Position
}

func (CstStatReturn) isCstNode() {}

type CstStatTypeAlias struct {
	TypeKeywordPosition    Position
	GenericsOpenPosition   *Position
	GenericsCommaPositions []Position
	GenericsClosePosition  *Position
	EqualsPosition         Position
}

func (CstStatTypeAlias) isCstNode() {}

type CstStatTypeFunction struct {
	TypeKeywordPosition     Position
	FunctionKeywordPosition Position
}

func (CstStatTypeFunction) isCstNode() {}

type CstExprFunction struct {
	FunctionKeywordPosition       Position
	OpenGenericsPosition          *Position
	GenericsCommaPositions        []Position
	CloseGenericsPosition         *Position
	ArgsAnnotationColonPositions  []*Position
	ArgsCommaPositions            []Position
	VarargAnnotationColonPosition *Position
	ReturnSpecifierPosition       *Position
}

type CstExprTable struct {
	Items []CstExprTableItem
}

func (CstExprTable) isCstNode() {}

type CstExprTableItem struct {
	Kind                 string
	EqualsPosition       *Position
	Separator            rune
	SeparatorPosition    Position
	IndexerOpenPosition  *Position
	IndexerClosePosition *Position
}

func (CstExprTableItem) isCstNode() {}

type CstExprIfElse struct {
	ThenPosition Position
	ElsePosition Position
	IsElseIf     bool
}

func (CstExprIfElse) isCstNode() {}

type CstExprInterpString struct {
	SourceStrings   []string
	StringPositions []Position
}

func (CstExprInterpString) isCstNode() {}

type CstExprConstantString struct {
	SourceString *string
	QuoteStyle   int
	BlockDepth   int
}

func (CstExprConstantString) isCstNode() {}

type CstTypeInstantiation struct {
	LeftArrow1Position Position
	LeftArrow2Position Position

	RightArrow1Position Position
	RightArrow2Position Position
	CommaPositions      []Position
}

func (CstTypeInstantiation) isCstNode() {}

type CstExprCall struct {
	OpenParens     *Position
	CloseParens    *Position
	CommaPositions []Position
	ExplicitTypes  *CstTypeInstantiation
}

func (CstExprCall) isCstNode() {}

type CstExprIndexExpr struct {
	OpenBracketPosition  Position
	CloseBracketPosition Position
}

func (CstExprIndexExpr) isCstNode() {}

type CstExprTypeAssertion struct {
	OpPosition Position
}

func (CstExprTypeAssertion) isCstNode() {}

type CstExprExplicitTypeInstantiation struct {
	Instantiation CstTypeInstantiation
}

func (CstExprExplicitTypeInstantiation) isCstNode() {}
