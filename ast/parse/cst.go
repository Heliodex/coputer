package main

import "github.com/Heliodex/coputer/ast/lex"

type (
	CstExprConstantInteger struct {
		Value string
	}

	CstExprConstantNumber struct {
		Value string
	}

	CstExprConstantString struct {
		SourceString *string
		QuoteStyle, BlockDepth int
	}

	CstExprFunction struct {
		FunctionKeywordPosition lex.Position
		OpenGenericsPosition *lex.Position
		GenericsCommaPositions []lex.Position
		CloseGenericsPosition *lex.Position
		ArgsAnnotationColonPositions []*lex.Position
		ArgsCommaPositions []lex.Position
		VarargAnnotationColonPosition, ReturnSpecifierPosition *lex.Position
	}

	CstExprIfElse struct {
		ThenPosition, ElsePosition lex.Position
		IsElseIf bool
	}

	CstExprIndexExpr struct {
		OpenBracketPosition, CloseBracketPosition lex.Position
	}

	CstExprInterpString struct {
		SourceStrings []string
		StringPositions []lex.Position
	}

	CstExprOp struct {
		OpPosition lex.Position
	}

	CstExprTableItem struct {
		Kind ExprTableItemKind
		EqualsPosition *lex.Position
		Separator *int
		SeparatorPosition lex.Position
		IndexerOpenPosition, IndexerClosePosition *lex.Position
	}

	CstExprTable struct {
		Items []CstExprTableItem
	}

	CstExprTypeAssertion struct {
		OpPosition lex.Position
	}

	CstGenericType struct {
		DefaultEqualsPosition *lex.Position
	}

	CstGenericTypePack struct {
		EllipsisPosition lex.Position
		DefaultEqualsPosition *lex.Position
	}

	// --------------------------------------------------------------------------------
	// -- CST NODE UNION TYPES
	// --------------------------------------------------------------------------------

	// Union type representing all possible Concrete Syntax Tree nodes.
	// CST nodes store positional data for syntactic elements (commas, keywords, etc.)
	// that are not semantically significant but needed for source-accurate transformations.

	CstNode interface {
		isCstNode()
	}

	CstStatAssign struct {
		VarsCommaPositions []lex.Position
		EqualsPosition lex.Position
		ValuesCommaPositions []lex.Position
	}

	CstStatBlock struct {
		BodyCommaPositions []lex.Position
	}

	CstStatCompoundAssign struct {
		OpPosition lex.Position
	}

	CstStatDo struct {
		StatsStart, EndPosition lex.Position
	}

	// type CstStatDo_DEPRECATED struct {
	// 	EndPosition lex.Position
	// }

	// func (CstStatDo_DEPRECATED) isCstNode() {}

	CstStatFor struct {
		AnnotationColonPosition *lex.Position
		EqualsPosition, EndCommaPosition lex.Position
		StepCommaPosition *lex.Position
	}

	CstStatForIn struct {
		VarsAnnotationColonPositions []*lex.Position
		VarsCommaPositions, ValuesCommaPositions []lex.Position
	}

	CstStatFunction struct {
		FunctionKeywordPosition lex.Position
	}

	CstStatLocal struct {
		VarsAnnotationColonPositions []*lex.Position
		VarsCommaPositions, ValuesCommaPositions []lex.Position
	}

	CstStatLocalFunction struct {
		LocalKeywordPosition, FunctionKeywordPosition lex.Position
	}

	CstStatRepeat struct {
		UntilPosition lex.Position
	}

	CstStatReturn struct {
		CommaPositions []lex.Position
	}

	CstStatTypeAlias struct {
		TypeKeywordPosition lex.Position
		GenericsOpenPosition *lex.Position
		GenericsCommaPositions []lex.Position
		GenericsClosePosition *lex.Position
		EqualsPosition lex.Position
	}

	CstStatTypeFunction struct {
		TypeKeywordPosition, FunctionKeywordPosition lex.Position
	}

	CstTypeFunction struct {
		OpenGenericsPosition *lex.Position
		GenericsCommaPositions []lex.Position
		CloseGenericsPosition *lex.Position
		OpenArgsPosition lex.Position
		ArgumentNameColonPositions []*lex.Position
		ArgumentsCommaPositions []lex.Position
		CloseArgsPosition, ReturnArrowPosition lex.Position
	}

	CstTypeInstantiation struct {
		LeftArrow1Position, LeftArrow2Position, RightArrow1Position, RightArrow2Position lex.Position
		CommaPositions []lex.Position
	}

	CstExprCall struct {
		OpenParens, CloseParens *lex.Position
		CommaPositions []lex.Position
		ExplicitTypes *CstTypeInstantiation
	}

	CstExprExplicitTypeInstantiation struct {
		Instantiation CstTypeInstantiation
	}

	CstTypeIntersection struct {
		LeadingPosition *lex.Position
		SeparatorPositions []lex.Position
	}

	CstTypePackExplicit struct {
		OpenParenthesesPosition, CloseParenthesesPosition *lex.Position
		CommaPositions *[]lex.Position
	}

	CstTypePackGeneric struct {
		EllipsisPosition lex.Position
	}

	CstTypeReference struct {
		PrefixPointPosition, OpenParametersPosition *lex.Position
		ParametersCommaPositions []lex.Position
		CloseParametersPosition *lex.Position
	}

	CstTypeSingletonString struct {
		SourceString []string
		QuoteStyle, BlockDepth int
	}

	CstTypeTableItem struct {
		Kind string // "Property" | "StringProperty" | "Indexer"
		ColonPosition *lex.Position
		Separator *int
		SeparatorPosition, IndexerOpenPosition, IndexerClosePosition *lex.Position
		StringInfo *CstExprConstantString
		StringPosition *lex.Location
		EqualsPosition *lex.Position
	}

	CstTypeTable struct {
		Items []CstTypeTableItem
		IsArray bool
	}

	CstTypeTypeof struct {
		OpenPosition, ClosePosition lex.Position
	}

	CstTypeUnion struct {
		LeadingPosition *lex.Position
		SeparatorPositions []lex.Position
	}
)

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
	_ CstNode = CstExprConstantInteger{}
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

func (CstExprCall) isCstNode() {}
func (CstExprConstantInteger) isCstNode() {}
func (CstExprConstantNumber) isCstNode() {}
func (CstExprConstantString) isCstNode() {}
func (CstExprExplicitTypeInstantiation) isCstNode() {}
func (CstExprFunction) isCstNode() {}
func (CstExprIfElse) isCstNode() {}
func (CstExprIndexExpr) isCstNode() {}
func (CstExprInterpString) isCstNode() {}
func (CstExprOp) isCstNode() {}
func (CstExprTable) isCstNode() {}
func (CstExprTableItem) isCstNode() {}
func (CstExprTypeAssertion) isCstNode() {}
func (CstGenericType) isCstNode() {}
func (CstGenericTypePack) isCstNode() {}
func (CstStatAssign) isCstNode() {}
func (CstStatBlock) isCstNode() {}
func (CstStatCompoundAssign) isCstNode() {}
func (CstStatDo) isCstNode() {}
func (CstStatFor) isCstNode() {}
func (CstStatForIn) isCstNode() {}
func (CstStatFunction) isCstNode() {}
func (CstStatLocal) isCstNode() {}
func (CstStatLocalFunction) isCstNode() {}
func (CstStatRepeat) isCstNode() {}
func (CstStatReturn) isCstNode() {}
func (CstStatTypeAlias) isCstNode() {}
func (CstStatTypeFunction) isCstNode() {}
func (CstTypeFunction) isCstNode() {}
func (CstTypeInstantiation) isCstNode() {}
func (CstTypeIntersection) isCstNode() {}
func (CstTypePackExplicit) isCstNode() {}
func (CstTypePackGeneric) isCstNode() {}
func (CstTypeReference) isCstNode() {}
func (CstTypeSingletonString) isCstNode() {}
func (CstTypeTable) isCstNode() {}
func (CstTypeTableItem) isCstNode() {}
func (CstTypeTypeof) isCstNode() {}
func (CstTypeUnion) isCstNode() {}

// SeparatorIsComma reports whether the item is followed by a comma. An item with no separator, such as the last item of a table without a trailing separator, reports false.
func (item CstExprTableItem) SeparatorIsComma() bool {
	return item.Separator == &pzero
}

// SeparatorIsSemicolon reports whether the item is followed by a semicolon. An item with no separator reports false.
func (item CstExprTableItem) SeparatorIsSemicolon() bool {
	return item.Separator == &pone
}
