package main

import (
	"fmt"
	"strings"

	"github.com/Heliodex/coputer/ast/lex"
)

type (
	AstArgumentName struct {
		Name string
		Location lex.Location
	}

	// --------------------------------------------------------------------------------
	// -- AST NODE UNION TYPE
	// --------------------------------------------------------------------------------

	// Union type representing all possible AST nodes that can be stored in the CST map.
	// This includes statements, expressions, types, locals, and various helper nodes.

	AstNode interface {
		isAstNode()
		String() string
		// Source renders the node as formatted Luau code.
		Source() string
		// Minify renders the node as compact Luau code.
		Minify() string
	}

	// --------------------------------------------------------------------------------
	// -- EXPRESSION UNION TYPES
	// --------------------------------------------------------------------------------

	// Union type representing all possible expression nodes in the AST.
	// Expressions are constructs that evaluate to a value.
	AstExpr interface {
		AstNode
		isAstExpr()
		GetLocation() lex.Location
	}

	AstExprConstantNumberOrError interface {
		AstExpr
		isAstExprConstantNumberOrError()
	}

	AstExprConstantStringOrError interface {
		AstExpr
		isAstExprConstantStringOrError()
	}

	AstExprInterpStringOrError interface {
		AstExpr
		isAstExprInterpStringOrError()
	}

	AstExprLocalOrGlobalOrError interface {
		AstExpr
		isAstExprLocalOrGlobalOrError()
	}

	// --------------------------------------------------------------------------------
	// -- STATEMENT UNION TYPES
	// --------------------------------------------------------------------------------

	// Union type representing all possible statement nodes in the AST.
	// Statements are constructs that perform actions but don't produce values.
	AstStat interface {
		AstNode
		isAstStat()
		GetLocation() lex.Location
		SetLocation(loc lex.Location)
		SetHasSemicolon()
	}

	AstStatBreakOrError interface {
		AstStat
		isAstStatBreakOrError()
	}

	AstStatContinueOrError interface {
		AstStat
		isAstStatContinueOrError()
	}

	// extra bonus

	AstStatForOrForIn interface {
		AstStat
		isAstStatForOrForIn()
	}

	AstStatTypeAliasOrTypeFunction interface {
		AstStat
		isAstStatTypeAliasOrTypeFunction()
	}

	// --------------------------------------------------------------------------------
	// -- TYPE ANNOTATION UNION TYPES
	// --------------------------------------------------------------------------------

	// Union type representing all possible type annotation nodes.
	// Type annotations specify the expected types of values in Luau's type system.

	AstType interface {
		AstNode
		isAstType()
		// isAstNodePack() 😭
		GetLocation() lex.Location
	}

	AstDeclaredExternTypeProperty struct {
		Location lex.Location
		Name lex.AstName
		NameLocation lex.Location
		Ty AstType
		IsMethod bool
	}

	AstTableIndexer struct {
		Location lex.Location
		IndexType, ResultType AstType
		Access string
		AccessLocation *lex.Location
	}

	// --------------------------------------------------------------------------------
	// -- TYPE PACK UNION TYPES
	// --------------------------------------------------------------------------------

	// Union type representing all possible type pack nodes.
	// Type packs represent multiple types, used for function return types and variadic arguments.

	AstTypePack interface {
		AstNode
		GetLocation() lex.Location
	}

	AstTypeList struct {
		Types []AstType
		TailType *AstTypePack
	}

	AstTypeOrPack struct {
		Type *AstType
		Pack *AstTypePack
	}

	// bonus round

	AstTypePackVariadicOrGeneric interface {
		AstTypePack
		isAstTypePackVariadicOrGeneric()
	}

	ExprTableItemKind uint8

	// base for every node

	NodeLoc struct {
		Location lex.Location
	}

	// node types (ok, real ast now)
	AstAttr struct {
		*NodeLoc

		Type string
		Args []AstExpr
		Name *string
	}

	AstExprBinary struct {
		*NodeLoc

		Op int
		Left, Right AstExpr
	}

	AstExprCall struct {
		*NodeLoc

		Func AstExpr
		Args []AstExpr
		Self bool
		ArgLocation lex.Location
		TypeArguments *[]AstTypeOrPack
	}

	AstExprConstantBool struct {
		*NodeLoc

		Value bool
	}

	AstExprConstantInteger struct {
		*NodeLoc

		Value int64
		ParseResult NumberParseResult
	}

	AstExprConstantNil struct {
		*NodeLoc
	}

	AstExprConstantNumber struct {
		*NodeLoc

		Value float64
	}

	AstExprConstantString struct {
		*NodeLoc

		Value string
		QuoteStyle QuoteStyle
	}

	AstExprError struct {
		*NodeLoc

		Expressions []AstExpr
		MessageIndex int
	}

	AstExprGlobal struct {
		*NodeLoc

		Name string
	}

	AstExprGroup struct {
		*NodeLoc

		Expr AstExpr
	}

	AstExprIndexExpr struct {
		*NodeLoc

		Expr, Index AstExpr
	}

	AstExprIndexName struct {
		*NodeLoc

		Expr AstExpr
		Index string
		IndexLocation lex.Location
		OpPosition lex.Position
		Op rune
	}

	AstExprInstantiate struct {
		*NodeLoc

		Expr AstExpr
		TypeArguments []AstTypeOrPack
	}

	AstExprInterpString struct {
		*NodeLoc

		Strings []string
		Expressions []AstExpr
	}

	AstExprTableItem struct {
		*NodeLoc

		Kind ExprTableItemKind
		Key *AstExpr
		Value AstExpr
		// Separator is the separator written after the item. The last item of a table may have none.
		Separator ExprTableSeparator
	}

	AstExprTypeAssertion struct {
		*NodeLoc

		Expr AstExpr
		Annotation AstType
	}

	AstExprUnary struct {
		*NodeLoc

		Op UnaryOp
		Expr AstExpr
	}

	AstExprVarargs struct {
		*NodeLoc
	}

	AstGenericType struct {
		*NodeLoc

		Name string
		DefaultValue *AstType
	}

	AstGenericTypePack struct {
		*NodeLoc

		Name string
		DefaultValue *AstTypePack
	}

	AstLocal struct {
		*NodeLoc

		Name string
		Shadow *AstLocal
		FunctionDepth, LoopDepth int
		Annotation AstType
		IsConst, IsExported bool
	}

	AstExprIfElse struct {
		*NodeLoc

		Condition AstExpr
		HasThen bool
		TrueExpr AstExpr
		HasElse bool
		FalseExpr AstExpr
		// Active for `if local` / `if const` expressions; ConditionLocal is in scope for TrueExpr only
		ConditionLocal *AstLocal
		ConditionIsConst bool
		ConditionKeywordLocation, ConditionEqualsLocation *lex.Location
	}

	AstExprLocal struct {
		*NodeLoc

		Local AstLocal
		Upvalue bool
	}

	AstStatAssign struct {
		*NodeLoc

		Vars, Values []AstExpr
		HasSemicolon *bool
	}

	AstStatBreak struct {
		*NodeLoc

		HasSemicolon *bool
	}

	AstStatCompoundAssign struct {
		*NodeLoc

		Op BinaryOp
		Var, Value AstExpr
		HasSemicolon *bool
	}

	AstStatContinue struct {
		*NodeLoc

		HasSemicolon *bool
	}

	AstStatDeclareExternType struct {
		*NodeLoc

		Name string
		SuperName *string
		Props []AstDeclaredExternTypeProperty
		Indexer *AstTableIndexer
		HasSemicolon *bool
	}

	AstStatDeclareFunction struct {
		*NodeLoc

		Attributes []AstAttr
		Name string
		NameLocation lex.Location
		Generics []AstGenericType
		GenericPacks []AstGenericTypePack
		Params AstTypeList
		ParamNames []AstArgumentName
		Vararg bool
		VarargLocation lex.Location
		RetTypes AstTypePack
		HasSemicolon *bool
	}

	AstStatDeclareGlobal struct {
		*NodeLoc

		Name string
		NameLocation lex.Location
		Type AstType
		HasSemicolon *bool
	}

	AstStatError struct {
		*NodeLoc

		Expressions []AstExpr
		Statements []AstStat
		MessageIndex int
		HasSemicolon *bool
	}

	AstStatExpr struct {
		*NodeLoc

		Expr AstExpr
		HasSemicolon *bool
	}

	AstStatLocal struct {
		*NodeLoc

		Vars []AstLocal
		Values []AstExpr
		EqualsSignLocation *lex.Location
		IsConst, IsExported bool
		KeywordLocation *lex.Location
		HasSemicolon *bool
	}

	AstStatReturn struct {
		*NodeLoc

		List []AstExpr
		HasSemicolon *bool
	}

	AstStatTypeAlias struct {
		*NodeLoc

		Name string
		NameLocation lex.Location
		Generics []AstGenericType
		GenericPacks []AstGenericTypePack
		Type AstType
		Exported bool
		HasSemicolon *bool
	}

	AstTableProp struct {
		Name lex.AstName
		*NodeLoc

		Type AstType
		Access string
		AccessLocation *lex.Location
	}

	AstTypeError struct {
		*NodeLoc

		Types []AstType
		IsMissing bool
		MessageIndex int
	}

	AstTypeFunction struct {
		*NodeLoc

		Attributes []AstAttr
		Generics []AstGenericType
		GenericPacks []AstGenericTypePack
		ArgTypes AstTypeList
		ArgNames []*AstArgumentName
		ReturnTypes AstTypePack
	}

	AstTypeGroup struct {
		*NodeLoc

		Type AstType
	}

	AstTypeIntersection struct {
		*NodeLoc

		Types []AstType
	}

	AstTypeOptional struct {
		*NodeLoc
	}

	AstTypePackExplicit struct {
		*NodeLoc

		Types AstTypeList
		TailType *AstTypePack
	}

	AstTypePackGeneric struct {
		*NodeLoc

		GenericName string
	}

	AstTypePackVariadic struct {
		*NodeLoc

		VariadicType AstType
	}

	AstTypeReference struct {
		*NodeLoc

		HasParameterList bool
		Prefix *string
		PrefixLocation *lex.Location
		Name string
		NameLocation lex.Location
		Parameters []AstTypeOrPack
	}

	AstTypeSingletonBool struct {
		*NodeLoc

		Value bool
	}

	AstTypeSingletonString struct {
		*NodeLoc

		Value string
	}

	// lol
	AstTypeTypeof struct {
		*NodeLoc

		Expr AstExpr
	}

	AstTypeUnion struct {
		*NodeLoc

		Types []AstType
	}

	// ast

	Comment struct {
		Type lex.LexemeType
		// Content is the comment body, without the `--` or block delimiters.
		Content string
		*NodeLoc
	}

	AstExprTable struct {
		*NodeLoc

		Items []AstExprTableItem
		// Comments contains the comments lexed within this table, in source order.
		Comments []Comment
	}

	AstStatBlock struct {
		*NodeLoc

		Body []AstStat
		HasEnd bool
		HasSemicolon *bool
		// Comments contains the comments lexed within this block, in source order.
		Comments []Comment
	}

	AstExprFunction struct {
		*NodeLoc

		Attributes []AstAttr
		Generics []AstGenericType
		GenericPacks []AstGenericTypePack
		Self *AstLocal
		Args []AstLocal
		ReturnAnnotation *AstTypePack
		Vararg bool
		VarargLocation lex.Location
		VarargAnnotation *AstTypePack
		Body AstStatBlock
		FunctionDepth int
		Debugname string
		ArgLocation *lex.Location
	}

	AstStatFor struct {
		*NodeLoc

		Var *AstLocal
		From, To, Step AstExpr
		Body *AstStatBlock
		HasDo bool
		DoLocation lex.Location
		HasSemicolon *bool
	}

	AstStatForIn struct {
		*NodeLoc

		Vars []*AstLocal
		Values []AstExpr
		Body *AstStatBlock
		HasIn bool
		InLocation lex.Location
		HasDo bool
		DoLocation lex.Location
		HasSemicolon *bool
	}

	AstStatFunction struct {
		*NodeLoc

		Name AstExpr
		Func AstExprFunction
		HasSemicolon *bool
	}

	AstStatIf struct {
		*NodeLoc

		Condition AstExpr
		ThenBody AstStatBlock
		ElseBody AstStat
		ThenLocation, ElseLocation *lex.Location
		// Active for `if local` and `if const` statements
		ConditionLocal *AstLocal
		ConditionIsConst bool
		ConditionKeywordLocation, ConditionEqualsLocation *lex.Location
		HasSemicolon *bool
	}

	AstStatLocalFunction struct {
		*NodeLoc

		Name AstLocal
		Func AstExprFunction
		IsConst bool
		HasSemicolon *bool
	}

	AstStatRepeat struct {
		*NodeLoc

		Condition AstExpr
		Body *AstStatBlock
		HasUntil bool
		HasSemicolon *bool
	}

	AstStatTypeFunction struct {
		*NodeLoc

		Name string
		NameLocation lex.Location
		Body AstExprFunction
		Exported, HasErrors bool
		HasSemicolon *bool
	}

	AstStatWhile struct {
		*NodeLoc

		Condition AstExpr
		Body *AstStatBlock
		HasDo bool
		DoLocation lex.Location
		HasSemicolon *bool
	}

	AstTypeTable struct {
		*NodeLoc

		Props []AstTableProp
		Indexer *AstTableIndexer
		// Comments contains the comments lexed within this table, in source order.
		Comments []Comment
	}
)

const (
	Ext = ".luau"
	AstDir = "../test/ast"
	BenchmarkDir = "../test/benchmark"
	ConformanceDir = "../test/conformance"
	FormattingDir = "../test/formatting"
	// FormattingOutSuffix is appended to a formatting test's name to get the file containing its expected formatted output.
	FormattingOutSuffix = "_out"
)

const (
	General ExprTableItemKind = iota
	Record
	List
)

var (
	_ AstExpr = AstExprGroup{}
	_ AstExpr = AstExprConstantNil{}
	_ AstExpr = AstExprConstantBool{}
	_ AstExpr = AstExprConstantNumber{}
	_ AstExpr = AstExprConstantString{}
	_ AstExpr = AstExprLocal{}
	_ AstExpr = AstExprGlobal{}
	_ AstExpr = AstExprVarargs{}
	_ AstExpr = AstExprCall{}
	_ AstExpr = AstExprIndexName{}
	_ AstExpr = AstExprIndexExpr{}
	_ AstExpr = AstExprFunction{}
	_ AstExpr = AstExprTable{}
	_ AstExpr = AstExprUnary{}
	_ AstExpr = AstExprBinary{}
	_ AstExpr = AstExprTypeAssertion{}
	_ AstExpr = AstExprIfElse{}
	_ AstExpr = AstExprInterpString{}
	_ AstExpr = AstExprInstantiate{}
	_ AstExpr = AstExprError{}
	_ AstStat = &AstStatBlock{}
	_ AstStat = &AstStatIf{}
	_ AstStat = &AstStatWhile{}
	_ AstStat = &AstStatRepeat{}
	_ AstStat = &AstStatBreak{}
	_ AstStat = &AstStatContinue{}
	_ AstStat = &AstStatReturn{}
	_ AstStat = &AstStatExpr{}
	_ AstStat = &AstStatLocal{}
	_ AstStat = &AstStatFor{}
	_ AstStat = &AstStatForIn{}
	_ AstStat = &AstStatAssign{}
	_ AstStat = &AstStatCompoundAssign{}
	_ AstStat = &AstStatFunction{}
	_ AstStat = &AstStatLocalFunction{}
	_ AstStat = &AstStatTypeAlias{}
	_ AstStat = &AstStatTypeFunction{}
	_ AstStat = &AstStatDeclareGlobal{}
	_ AstStat = &AstStatDeclareFunction{}
	_ AstStat = &AstStatDeclareExternType{}
	_ AstStat = &AstStatError{}
	_ AstStatForOrForIn = &AstStatFor{}
	_ AstStatForOrForIn = &AstStatForIn{}
	_ AstStatBreakOrError = &AstStatBreak{}
	_ AstStatBreakOrError = &AstStatError{}
	_ AstStatContinueOrError = &AstStatContinue{}
	_ AstStatContinueOrError = &AstStatError{}
	_ AstStatTypeAliasOrTypeFunction = &AstStatTypeAlias{}
	_ AstStatTypeAliasOrTypeFunction = &AstStatTypeFunction{}
	_ AstExprLocalOrGlobalOrError = AstExprLocal{}
	_ AstExprLocalOrGlobalOrError = AstExprGlobal{}
	_ AstExprLocalOrGlobalOrError = AstExprError{}
	_ AstExprInterpStringOrError = AstExprInterpString{}
	_ AstExprInterpStringOrError = AstExprError{}
	_ AstExprConstantStringOrError = AstExprConstantString{}
	_ AstExprConstantStringOrError = AstExprError{}
	_ AstExprConstantNumberOrError = AstExprConstantNumber{}
	_ AstExprConstantNumberOrError = AstExprConstantInteger{}
	_ AstExprConstantNumberOrError = AstExprError{}
	_ AstTypePack = AstTypePackExplicit{}
	_ AstTypePack = AstTypePackGeneric{}
	_ AstTypePack = AstTypePackVariadic{}
	_ AstTypePackVariadicOrGeneric = AstTypePackGeneric{}
	_ AstTypePackVariadicOrGeneric = AstTypePackVariadic{}
	_ AstType = AstTypeReference{}
	_ AstType = AstTypeTable{}
	_ AstType = AstTypeFunction{}
	_ AstType = AstTypeTypeof{}
	_ AstType = AstTypeUnion{}
	_ AstType = AstTypeIntersection{}
	_ AstType = AstTypeSingletonBool{}
	_ AstType = AstTypeSingletonString{}
	_ AstType = AstTypeGroup{}
	_ AstType = AstTypeError{}
	_ AstType = AstTypeOptional{}

	// tru is an immutable sentinel used for HasSemicolon flags.
	tru = true
)

func (AstExprBinary) isAstExpr() {}
func (AstExprBinary) isAstNode() {}
func (AstExprCall) isAstExpr() {}
func (AstExprCall) isAstNode() {}
func (AstExprConstantBool) isAstExpr() {}
func (AstExprConstantBool) isAstNode() {}
func (AstExprConstantInteger) isAstExpr() {}
func (AstExprConstantInteger) isAstExprConstantNumberOrError() {}
func (AstExprConstantInteger) isAstNode() {}
func (AstExprConstantNil) isAstExpr() {}
func (AstExprConstantNil) isAstNode() {}
func (AstExprConstantNumber) isAstExpr() {}
func (AstExprConstantNumber) isAstExprConstantNumberOrError() {}
func (AstExprConstantNumber) isAstNode() {}
func (AstExprConstantString) isAstExpr() {}
func (AstExprConstantString) isAstExprConstantStringOrError() {}
func (AstExprConstantString) isAstNode() {}
func (AstExprError) isAstExpr() {}
func (AstExprError) isAstExprConstantNumberOrError() {}
func (AstExprError) isAstExprConstantStringOrError() {}
func (AstExprError) isAstExprInterpStringOrError() {}
func (AstExprError) isAstExprLocalOrGlobalOrError() {}
func (AstExprError) isAstNode() {}
func (AstExprFunction) isAstExpr() {}
func (AstExprFunction) isAstNode() {}
func (AstExprGlobal) isAstExpr() {}
func (AstExprGlobal) isAstExprLocalOrGlobalOrError() {}
func (AstExprGlobal) isAstNode() {}
func (AstExprGroup) isAstExpr() {}
func (AstExprGroup) isAstNode() {}
func (AstExprIfElse) isAstExpr() {}
func (AstExprIfElse) isAstNode() {}
func (AstExprIndexExpr) isAstExpr() {}
func (AstExprIndexExpr) isAstNode() {}
func (AstExprIndexName) isAstExpr() {}
func (AstExprIndexName) isAstNode() {}
func (AstExprInstantiate) isAstExpr() {}
func (AstExprInstantiate) isAstNode() {}
func (AstExprInterpString) isAstExpr() {}
func (AstExprInterpString) isAstExprInterpStringOrError() {}
func (AstExprInterpString) isAstNode() {}
func (AstExprLocal) isAstExpr() {}
func (AstExprLocal) isAstExprLocalOrGlobalOrError() {}
func (AstExprLocal) isAstNode() {}
func (AstExprTable) isAstExpr() {}
func (AstExprTable) isAstNode() {}
func (AstExprTypeAssertion) isAstExpr() {}
func (AstExprTypeAssertion) isAstNode() {}
func (AstExprUnary) isAstExpr() {}
func (AstExprUnary) isAstNode() {}
func (AstExprVarargs) isAstExpr() {}
func (AstExprVarargs) isAstNode() {}

func (n *AstStatAssign) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatAssign) isAstNode() {}
func (AstStatAssign) isAstStat() {}

func (n *AstStatBlock) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatBlock) isAstNode() {}
func (AstStatBlock) isAstStat() {}

func (n *AstStatBreak) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatBreak) isAstNode() {}
func (AstStatBreak) isAstStat() {}
func (AstStatBreak) isAstStatBreakOrError() {}

func (n *AstStatCompoundAssign) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatCompoundAssign) isAstNode() {}
func (AstStatCompoundAssign) isAstStat() {}

func (n *AstStatContinue) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatContinue) isAstNode() {}
func (AstStatContinue) isAstStat() {}
func (AstStatContinue) isAstStatContinueOrError() {}

func (n *AstStatDeclareExternType) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatDeclareExternType) isAstNode() {}
func (AstStatDeclareExternType) isAstStat() {}

func (n *AstStatDeclareFunction) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatDeclareFunction) isAstNode() {}
func (AstStatDeclareFunction) isAstStat() {}

func (n *AstStatDeclareGlobal) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatDeclareGlobal) isAstNode() {}
func (AstStatDeclareGlobal) isAstStat() {}

func (n *AstStatError) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatError) isAstNode() {}
func (AstStatError) isAstStat() {}
func (AstStatError) isAstStatBreakOrError() {}
func (AstStatError) isAstStatContinueOrError() {}

func (n *AstStatExpr) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatExpr) isAstNode() {}
func (AstStatExpr) isAstStat() {}

func (n *AstStatFor) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatFor) isAstNode() {}
func (AstStatFor) isAstStat() {}
func (AstStatFor) isAstStatForOrForIn() {}

func (n *AstStatForIn) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatForIn) isAstNode() {}
func (AstStatForIn) isAstStat() {}
func (AstStatForIn) isAstStatForOrForIn() {}

func (n *AstStatFunction) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatFunction) isAstNode() {}
func (AstStatFunction) isAstStat() {}

func (n *AstStatIf) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatIf) isAstNode() {}
func (AstStatIf) isAstStat() {}

func (n *AstStatLocal) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatLocal) isAstNode() {}
func (AstStatLocal) isAstStat() {}

func (n *AstStatLocalFunction) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatLocalFunction) isAstNode() {}
func (AstStatLocalFunction) isAstStat() {}

func (n *AstStatRepeat) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatRepeat) isAstNode() {}
func (AstStatRepeat) isAstStat() {}

func (n *AstStatReturn) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatReturn) isAstNode() {}
func (AstStatReturn) isAstStat() {}

func (n *AstStatTypeAlias) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatTypeAlias) isAstNode() {}
func (AstStatTypeAlias) isAstStat() {}
func (AstStatTypeAlias) isAstStatTypeAliasOrTypeFunction() {}

func (n *AstStatTypeFunction) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatTypeFunction) isAstNode() {}
func (AstStatTypeFunction) isAstStat() {}
func (AstStatTypeFunction) isAstStatTypeAliasOrTypeFunction() {}

func (n *AstStatWhile) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (AstStatWhile) isAstNode() {}
func (AstStatWhile) isAstStat() {}
func (AstTypeError) isAstNode() {}
func (AstTypeError) isAstType() {}
func (AstTypeFunction) isAstNode() {}
func (AstTypeFunction) isAstType() {}
func (AstTypeGroup) isAstNode() {}
func (AstTypeGroup) isAstType() {}
func (AstTypeIntersection) isAstNode() {}
func (AstTypeIntersection) isAstType() {}
func (AstTypeList) isAstNode() {}
func (AstTypeOptional) isAstNode() {}
func (AstTypeOptional) isAstType() {}
func (AstTypePackExplicit) isAstNode() {}
func (AstTypePackGeneric) isAstNode() {}
func (AstTypePackGeneric) isAstTypePackVariadicOrGeneric() {}
func (AstTypePackVariadic) isAstNode() {}
func (AstTypePackVariadic) isAstTypePackVariadicOrGeneric() {}
func (AstTypeReference) isAstNode() {}
func (AstTypeReference) isAstType() {}
func (AstTypeSingletonBool) isAstNode() {}
func (AstTypeSingletonBool) isAstType() {}
func (AstTypeSingletonString) isAstNode() {}
func (AstTypeSingletonString) isAstType() {}
func (AstTypeTable) isAstNode() {}
func (AstTypeTable) isAstType() {}
func (AstTypeTypeof) isAstNode() {}
func (AstTypeTypeof) isAstType() {}
func (AstTypeUnion) isAstNode() {}
func (AstTypeUnion) isAstType() {}

// ExprTableSeparator is the separator written after a table item: nothing, a comma, or a semicolon.
type ExprTableSeparator uint8

const (
	NoSeparator ExprTableSeparator = iota
	CommaSeparator
	SemicolonSeparator
)

func (e ExprTableItemKind) String() string {
	switch e {
	case General:
		return "General"
	case Record:
		return "Record"
	case List:
		return "List"
	}

	return "Unknown"
}

func (l NodeLoc) GetLocation() lex.Location {
	return l.Location
}

func (l *NodeLoc) SetLocation(loc lex.Location) {
	l.Location = loc
}

// ast groops

func indentStart(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")

	for i, line := range lines {
		lines[i] = strings.Repeat("  ", n) + line
	}

	return strings.Join(lines, "\n")
}

func (l *NodeLoc) String() string {
	return fmt.Sprintf("Location: %s\n", l.Location.String())
}

func (n Comment) String() string {
	var b strings.Builder

	b.WriteString("Comment\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Type: %v\n", n.Type)

	return b.String()
}

func (n AstAttr) String() string {
	var b strings.Builder

	b.WriteString("Attr\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Type: %q\n", n.Type)

	if len(n.Args) > 0 {
		b.WriteString("Args:\n")

		for _, arg := range n.Args {
			b.WriteString(indentStart(arg.String(), 2))
			b.WriteByte('\n')
		}
	}

	if n.Name != nil {
		fmt.Fprintf(&b, "Name: %q\n", *n.Name)
	}

	return b.String()
}

func (n AstArgumentName) String() string {
	var b strings.Builder

	b.WriteString("ArgumentName\n")
	fmt.Fprintf(&b, "Name: %q\n", n.Name)
	fmt.Fprintf(&b, "Location: %s\n", n.Location.String())

	return b.String()
}

func (n AstExprBinary) String() string {
	var b strings.Builder

	b.WriteString("ExprBinary\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Op: %d\n", n.Op)
	b.WriteString("Left:\n")
	b.WriteString(indentStart(n.Left.String(), 2))
	b.WriteByte('\n')
	b.WriteString("Right:\n")
	b.WriteString(indentStart(n.Right.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstExprCall) String() string {
	var b strings.Builder

	b.WriteString("ExprCall\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Func:\n")
	b.WriteString(indentStart(n.Func.String(), 2))
	b.WriteByte('\n')

	if len(n.Args) > 0 {
		b.WriteString("Args:\n")

		for _, arg := range n.Args {
			b.WriteString(indentStart(arg.String(), 2))
			b.WriteByte('\n')
		}
	}

	fmt.Fprintf(&b, "Self: %t\n", n.Self)
	fmt.Fprintf(&b, "ArgLocation: %s\n", n.ArgLocation.String())

	if n.TypeArguments != nil {
		b.WriteString("TypeArguments:\n")

		for _, typeArg := range *n.TypeArguments {
			b.WriteString(indentStart(typeArg.String(), 2))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func (n AstExprConstantBool) String() string {
	var b strings.Builder

	b.WriteString("ExprConstantBool\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Value: %t\n", n.Value)

	return b.String()
}

func (n AstExprConstantNil) String() string {
	var b strings.Builder

	b.WriteString("ExprConstantNil\n")
	b.WriteString(n.NodeLoc.String())

	return b.String()
}

func (n AstExprConstantNumber) String() string {
	var b strings.Builder

	b.WriteString("ExprConstantNumber\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Value: %f\n", n.Value)

	return b.String()
}

func (n AstExprConstantInteger) String() string {
	var b strings.Builder

	b.WriteString("ExprConstantInteger\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Value: %d\n", n.Value)

	return b.String()
}

func (n AstExprConstantString) String() string {
	var b strings.Builder

	b.WriteString("ExprConstantString\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Value: %q\n", n.Value)
	fmt.Fprintf(&b, "QuoteStyle: %v\n", n.QuoteStyle)

	return b.String()
}

func (n AstExprError) String() string {
	var b strings.Builder

	b.WriteString("ExprError\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Expressions:\n")

	for _, expr := range n.Expressions {
		b.WriteString(indentStart(expr.String(), 2))
		b.WriteByte('\n')
	}

	fmt.Fprintf(&b, "MessageIndex: %d\n", n.MessageIndex)

	return b.String()
}

func (n AstExprFunction) String() string {
	var b strings.Builder

	b.WriteString("ExprFunction\n")
	b.WriteString(n.NodeLoc.String())

	if len(n.Attributes) > 0 {
		b.WriteString("Attributes:\n")

		for _, attr := range n.Attributes {
			b.WriteString(indentStart(attr.String(), 2))
			b.WriteByte('\n')
		}
	}

	if len(n.Generics) > 0 {
		b.WriteString("Generics:\n")

		for _, generic := range n.Generics {
			b.WriteString(indentStart(generic.String(), 2))
			b.WriteByte('\n')
		}
	}

	if len(n.GenericPacks) > 0 {
		b.WriteString("GenericPacks:\n")

		for _, genericPack := range n.GenericPacks {
			b.WriteString(indentStart(genericPack.String(), 2))
			b.WriteByte('\n')
		}
	}

	if n.Self != nil {
		b.WriteString("Self:\n")
		b.WriteString(indentStart(n.Self.String(), 2))
		b.WriteByte('\n')
	}

	if len(n.Args) > 0 {
		b.WriteString("Args:\n")

		for _, arg := range n.Args {
			b.WriteString(indentStart(arg.String(), 2))
			b.WriteByte('\n')
		}
	}

	if n.ReturnAnnotation != nil {
		b.WriteString("ReturnAnnotation:\n")
		b.WriteString(indentStart((*n.ReturnAnnotation).String(), 2))
		b.WriteByte('\n')
	}

	fmt.Fprintf(&b, "Vararg: %t\n", n.Vararg)
	fmt.Fprintf(&b, "VarargLocation: %s\n", n.VarargLocation.String())

	if n.VarargAnnotation != nil {
		b.WriteString("VarargAnnotation:\n")
		b.WriteString(indentStart((*n.VarargAnnotation).String(), 2))
		b.WriteByte('\n')
	}

	b.WriteString("Body:\n")
	b.WriteString(indentStart(n.Body.String(), 2))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "FunctionDepth: %d\n", n.FunctionDepth)
	fmt.Fprintf(&b, "Debugname: %q\n", n.Debugname)

	if n.ArgLocation != nil {
		fmt.Fprintf(&b, "ArgLocation: %s\n", n.ArgLocation.String())
	}

	return b.String()
}

func (n AstExprGlobal) String() string {
	var b strings.Builder

	b.WriteString("ExprGlobal\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Name: %q\n", n.Name)

	return b.String()
}

func (n AstExprGroup) String() string {
	var b strings.Builder

	b.WriteString("ExprGroup\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Expr:\n")
	b.WriteString(indentStart(n.Expr.String(), 2))

	return b.String()
}

func (n AstExprIfElse) String() string {
	var b strings.Builder

	b.WriteString("ExprIfElse\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Condition:\n")
	b.WriteString(indentStart(n.Condition.String(), 2))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "HasThen: %t\n", n.HasThen)
	b.WriteString("TrueExpr:\n")
	b.WriteString(indentStart(n.TrueExpr.String(), 2))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "HasElse: %t\n", n.HasElse)
	b.WriteString("FalseExpr:\n")
	b.WriteString(indentStart(n.FalseExpr.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstExprIndexExpr) String() string {
	var b strings.Builder

	b.WriteString("ExprIndexExpr\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Expr:\n")
	b.WriteString(indentStart(n.Expr.String(), 2))
	b.WriteByte('\n')
	b.WriteString("Index:\n")
	b.WriteString(indentStart(n.Index.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstExprIndexName) String() string {
	var b strings.Builder

	b.WriteString("ExprIndexName\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Expr:\n")
	b.WriteString(indentStart(n.Expr.String(), 2))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "Index: %q\n", n.Index)
	fmt.Fprintf(&b, "IndexLocation: %s\n", n.IndexLocation.String())
	fmt.Fprintf(&b, "Op: %c\n", n.Op)

	return b.String()
}

func (n AstExprInterpString) String() string {
	var b strings.Builder

	b.WriteString("ExprInterpString\n")
	b.WriteString(n.NodeLoc.String())

	if len(n.Strings) > 0 {
		b.WriteString("Strings:\n")

		for _, s := range n.Strings {
			b.WriteString(indentStart(fmt.Sprintf("%q", s), 2))
			b.WriteByte('\n')
		}
	}

	if len(n.Expressions) > 0 {
		b.WriteString("Expressions:\n")

		for _, expr := range n.Expressions {
			b.WriteString(indentStart(expr.String(), 2))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func (n AstExprInstantiate) String() string {
	var b strings.Builder

	b.WriteString("ExprInstantiate\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Expr:\n")
	b.WriteString(indentStart(n.Expr.String(), 2))
	b.WriteByte('\n')

	if len(n.TypeArguments) > 0 {
		b.WriteString("TypeArguments:\n")

		for _, typeArg := range n.TypeArguments {
			b.WriteString(indentStart(typeArg.String(), 2))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func (n AstExprLocal) String() string {
	var b strings.Builder

	b.WriteString("ExprLocal\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Local:\n")
	b.WriteString(indentStart(n.Local.String(), 2))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "Upvalue: %t\n", n.Upvalue)

	return b.String()
}

func (n AstExprTable) String() string {
	var b strings.Builder

	b.WriteString("ExprTable\n")
	b.WriteString(n.NodeLoc.String())

	if len(n.Items) > 0 {
		b.WriteString("Items:\n")

		for _, item := range n.Items {
			b.WriteString(indentStart(item.String(), 2))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func (n AstExprTableItem) String() string {
	var b strings.Builder

	b.WriteString("ExprTableItem\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Kind: %q\n", n.Kind)

	if n.Key != nil {
		b.WriteString("Key:\n")
		b.WriteString(indentStart((*n.Key).String(), 2))
		b.WriteByte('\n')
	}

	b.WriteString("Value:\n")
	b.WriteString(indentStart(n.Value.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstExprTypeAssertion) String() string {
	var b strings.Builder

	b.WriteString("ExprTypeAssertion\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Expr:\n")
	b.WriteString(indentStart(n.Expr.String(), 2))
	b.WriteByte('\n')
	b.WriteString("Annotation:\n")
	b.WriteString(indentStart(n.Annotation.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstExprVarargs) String() string {
	var b strings.Builder

	b.WriteString("ExprVarargs\n")
	b.WriteString(n.NodeLoc.String())

	return b.String()
}

func (n AstExprUnary) String() string {
	var b strings.Builder

	b.WriteString("ExprUnary\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Op: %v\n", n.Op)
	b.WriteString("Expr:\n")
	b.WriteString(indentStart(n.Expr.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstGenericType) String() string {
	var b strings.Builder

	b.WriteString("GenericType\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Name: %q\n", n.Name)

	if n.DefaultValue != nil {
		b.WriteString("DefaultValue:\n")
		b.WriteString(indentStart((*n.DefaultValue).String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

func (n AstGenericTypePack) String() string {
	var b strings.Builder

	b.WriteString("GenericTypePack\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Name: %q\n", n.Name)

	if n.DefaultValue != nil {
		b.WriteString("DefaultValue:\n")
		b.WriteString(indentStart((*n.DefaultValue).String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

func (n AstLocal) String() string {
	var b strings.Builder

	b.WriteString("Local\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Name: %q\n", n.Name)

	if n.Shadow != nil {
		fmt.Fprintf(&b, "Shadow: %q\n", n.Shadow.Name)
	}

	if n.IsConst {
		b.WriteString("IsConst: true\n")
	}

	if n.IsExported {
		b.WriteString("IsExported: true\n")
	}

	fmt.Fprintf(&b, "FunctionDepth: %d\n", n.FunctionDepth)
	fmt.Fprintf(&b, "LoopDepth: %d\n", n.LoopDepth)

	if n.Annotation != nil {
		b.WriteString("Annotation:\n")
		b.WriteString(indentStart(n.Annotation.String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

func (n AstStatAssign) String() string {
	var b strings.Builder

	b.WriteString("StatAssign\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Vars:\n")

	for _, v := range n.Vars {
		b.WriteString(indentStart(v.String(), 2))
		b.WriteByte('\n')
	}

	if len(n.Values) > 0 {
		b.WriteString("Values:\n")

		for _, val := range n.Values {
			b.WriteString(indentStart(val.String(), 2))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func (n AstStatBlock) String() string {
	var b strings.Builder

	b.WriteString("StatBlock\n")
	b.WriteString(n.NodeLoc.String())

	if len(n.Body) > 0 {
		b.WriteString("Body:\n")

		for _, stat := range n.Body {
			b.WriteString(indentStart(stat.String(), 2))
			b.WriteByte('\n')
		}
	}

	fmt.Fprintf(&b, "HasEnd: %t\n", n.HasEnd)

	return b.String()
}

func (n AstStatBreak) String() string {
	var b strings.Builder

	b.WriteString("StatBreak\n")
	b.WriteString(n.NodeLoc.String())

	return b.String()
}

func (n AstStatCompoundAssign) String() string {
	var b strings.Builder

	b.WriteString("StatCompoundAssign\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Op: %d\n", n.Op)
	b.WriteString("Var:\n")
	b.WriteString(indentStart(n.Var.String(), 2))
	b.WriteByte('\n')
	b.WriteString("Value:\n")
	b.WriteString(indentStart(n.Value.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstStatContinue) String() string {
	var b strings.Builder

	b.WriteString("StatContinue\n")
	b.WriteString(n.NodeLoc.String())

	return b.String()
}

func (n AstStatDeclareFunction) String() string {
	var b strings.Builder

	b.WriteString("StatDeclareFunction\n")
	b.WriteString(n.NodeLoc.String())

	if len(n.Attributes) > 0 {
		b.WriteString("Attributes:\n")

		for _, attr := range n.Attributes {
			b.WriteString(indentStart(attr.String(), 2))
			b.WriteByte('\n')
		}
	}

	fmt.Fprintf(&b, "Name: %q\n", n.Name)
	fmt.Fprintf(&b, "NameLocation: %s\n", n.NameLocation.String())

	if len(n.Generics) > 0 {
		b.WriteString("Generics:\n")

		for _, generic := range n.Generics {
			b.WriteString(indentStart(generic.String(), 2))
			b.WriteByte('\n')
		}
	}

	if len(n.GenericPacks) > 0 {
		b.WriteString("GenericPacks:\n")

		for _, genericPack := range n.GenericPacks {
			b.WriteString(indentStart(genericPack.String(), 2))
			b.WriteByte('\n')
		}
	}

	b.WriteString("Params:\n")
	b.WriteString(indentStart(n.Params.String(), 2))
	b.WriteByte('\n')

	if len(n.ParamNames) > 0 {
		b.WriteString("ParamNames:\n")

		for _, name := range n.ParamNames {
			b.WriteString(indentStart(name.String(), 2))
			b.WriteByte('\n')
		}
	}

	fmt.Fprintf(&b, "Vararg: %t\n", n.Vararg)
	fmt.Fprintf(&b, "VarargLocation: %s\n", n.VarargLocation.String())

	if n.RetTypes != nil {
		b.WriteString("RetTypes:\n")
		b.WriteString(indentStart(n.RetTypes.String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

func (n AstStatDeclareGlobal) String() string {
	var b strings.Builder

	b.WriteString("StatDeclareGlobal\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Name: %q\n", n.Name)
	fmt.Fprintf(&b, "NameLocation: %s\n", n.NameLocation.String())
	b.WriteString("Type:\n")
	b.WriteString(indentStart(n.Type.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstStatDeclareExternType) String() string {
	var b strings.Builder

	b.WriteString("StatDeclareExternType\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Name: %q\n", n.Name)

	if n.SuperName != nil {
		fmt.Fprintf(&b, "SuperName: %q\n", *n.SuperName)
	}

	if len(n.Props) > 0 {
		b.WriteString("Props:\n")

		for _, prop := range n.Props {
			b.WriteString(indentStart(prop.String(), 2))
			b.WriteByte('\n')
		}
	}

	if n.Indexer != nil {
		b.WriteString("Indexer:\n")
		b.WriteString(indentStart(n.Indexer.String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

func (n AstDeclaredExternTypeProperty) String() string {
	var b strings.Builder

	b.WriteString("DeclaredExternTypeProperty\n")
	fmt.Fprintf(&b, "Location: %s\n", n.Location.String())
	fmt.Fprintf(&b, "Name: %q\n", n.Name.Value)
	fmt.Fprintf(&b, "NameLocation: %s\n", n.NameLocation.String())
	b.WriteString("Ty:\n")
	b.WriteString(indentStart(n.Ty.String(), 2))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "IsMethod: %t\n", n.IsMethod)

	return b.String()
}

func (n AstStatError) String() string {
	var b strings.Builder

	b.WriteString("StatError\n")
	b.WriteString(n.NodeLoc.String())

	if len(n.Expressions) > 0 {
		b.WriteString("Expressions:\n")

		for _, expr := range n.Expressions {
			b.WriteString(indentStart(expr.String(), 2))
			b.WriteByte('\n')
		}
	}

	if len(n.Statements) > 0 {
		b.WriteString("Statements:\n")

		for _, stat := range n.Statements {
			b.WriteString(indentStart(stat.String(), 2))
			b.WriteByte('\n')
		}
	}

	fmt.Fprintf(&b, "MessageIndex: %d\n", n.MessageIndex)

	return b.String()
}

func (n AstStatExpr) String() string {
	var b strings.Builder

	b.WriteString("StatExpr\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Expr:\n")
	b.WriteString(indentStart(n.Expr.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstStatFor) String() string {
	var b strings.Builder

	b.WriteString("StatFor\n")
	b.WriteString(n.NodeLoc.String())

	if n.Var != nil {
		b.WriteString("Var:\n")
		b.WriteString(indentStart(n.Var.String(), 2))
		b.WriteByte('\n')
	}

	b.WriteString("From:\n")
	b.WriteString(indentStart(n.From.String(), 2))
	b.WriteByte('\n')
	b.WriteString("To:\n")
	b.WriteString(indentStart(n.To.String(), 2))
	b.WriteByte('\n')

	if n.Step != nil {
		b.WriteString("Step:\n")
		b.WriteString(indentStart(n.Step.String(), 2))
		b.WriteByte('\n')
	}

	if n.Body != nil {
		b.WriteString("Body:\n")
		b.WriteString(indentStart(n.Body.String(), 2))
		b.WriteByte('\n')
	}

	fmt.Fprintf(&b, "HasDo: %t\n", n.HasDo)
	fmt.Fprintf(&b, "DoLocation: %s\n", n.DoLocation.String())

	return b.String()
}

func (n AstStatForIn) String() string {
	var b strings.Builder

	b.WriteString("StatForIn\n")
	b.WriteString(n.NodeLoc.String())

	if len(n.Vars) > 0 {
		b.WriteString("Vars:\n")

		for _, v := range n.Vars {
			b.WriteString(indentStart(v.String(), 2))
			b.WriteByte('\n')
		}
	}

	if len(n.Values) > 0 {
		b.WriteString("Values:\n")

		for _, val := range n.Values {
			b.WriteString(indentStart(val.String(), 2))
			b.WriteByte('\n')
		}
	}

	if n.Body != nil {
		b.WriteString("Body:\n")
		b.WriteString(indentStart(n.Body.String(), 2))
		b.WriteByte('\n')
	}

	fmt.Fprintf(&b, "HasIn: %t\n", n.HasIn)
	fmt.Fprintf(&b, "InLocation: %s\n", n.InLocation.String())
	fmt.Fprintf(&b, "HasDo: %t\n", n.HasDo)
	fmt.Fprintf(&b, "DoLocation: %s\n", n.DoLocation.String())

	return b.String()
}

func (n AstStatFunction) String() string {
	var b strings.Builder

	b.WriteString("StatFunction\n")
	b.WriteString(n.NodeLoc.String())

	if n.Name != nil {
		b.WriteString("Name:\n")
		b.WriteString(indentStart(n.Name.String(), 2))
	}

	b.WriteByte('\n')
	b.WriteString("Func:\n")
	b.WriteString(indentStart(n.Func.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstStatIf) String() string {
	var b strings.Builder

	b.WriteString("StatIf\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Condition:\n")
	b.WriteString(indentStart(n.Condition.String(), 2))
	b.WriteByte('\n')
	b.WriteString("ThenBody:\n")
	b.WriteString(indentStart(n.ThenBody.String(), 2))
	b.WriteByte('\n')

	if n.ElseBody != nil {
		b.WriteString("ElseBody:\n")
		b.WriteString(indentStart(n.ElseBody.String(), 2))
		b.WriteByte('\n')
	}

	if n.ThenLocation != nil {
		fmt.Fprintf(&b, "ThenLocation: %s\n", n.ThenLocation.String())
	}

	if n.ElseLocation != nil {
		fmt.Fprintf(&b, "ElseLocation: %s\n", n.ElseLocation.String())
	}

	return b.String()
}

func (n AstStatLocal) String() string {
	var b strings.Builder

	b.WriteString("StatLocal\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Vars:\n")

	for _, v := range n.Vars {
		b.WriteString(indentStart(v.String(), 2))
		b.WriteByte('\n')
	}

	if len(n.Values) > 0 {
		b.WriteString("Values:\n")

		for _, val := range n.Values {
			b.WriteString(indentStart(val.String(), 2))
			b.WriteByte('\n')
		}
	}

	if n.EqualsSignLocation != nil {
		fmt.Fprintf(&b, "EqualsSignLocation: %s\n", n.EqualsSignLocation.String())
	}

	return b.String()
}

func (n AstStatLocalFunction) String() string {
	var b strings.Builder

	b.WriteString("StatLocalFunction\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Name:\n")
	b.WriteString(indentStart(n.Name.String(), 2))
	b.WriteByte('\n')
	b.WriteString("Func:\n")
	b.WriteString(indentStart(n.Func.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstStatRepeat) String() string {
	var b strings.Builder

	b.WriteString("StatRepeat\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Condition:\n")
	b.WriteString(indentStart(n.Condition.String(), 2))
	b.WriteByte('\n')

	if n.Body != nil {
		b.WriteString("Body:\n")
		b.WriteString(indentStart(n.Body.String(), 2))
		b.WriteByte('\n')
	}

	fmt.Fprintf(&b, "HasUntil: %t\n", n.HasUntil)

	return b.String()
}

func (n AstStatReturn) String() string {
	var b strings.Builder

	b.WriteString("StatReturn\n")
	b.WriteString(n.NodeLoc.String())

	if len(n.List) > 0 {
		b.WriteString("List:\n")

		for _, expr := range n.List {
			b.WriteString(indentStart(expr.String(), 2))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func (n AstStatTypeAlias) String() string {
	var b strings.Builder

	b.WriteString("StatTypeAlias\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Name: %q\n", n.Name)
	fmt.Fprintf(&b, "NameLocation: %s\n", n.NameLocation.String())

	if len(n.Generics) > 0 {
		b.WriteString("Generics:\n")

		for _, generic := range n.Generics {
			b.WriteString(indentStart(generic.String(), 2))
			b.WriteByte('\n')
		}
	}

	if len(n.GenericPacks) > 0 {
		b.WriteString("GenericPacks:\n")

		for _, genericPack := range n.GenericPacks {
			b.WriteString(indentStart(genericPack.String(), 2))
			b.WriteByte('\n')
		}
	}

	b.WriteString("Type:\n")
	b.WriteString(indentStart(n.Type.String(), 2))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "Exported: %t\n", n.Exported)

	return b.String()
}

func (n AstStatTypeFunction) String() string {
	var b strings.Builder

	b.WriteString("StatTypeFunction\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Name: %q\n", n.Name)
	fmt.Fprintf(&b, "NameLocation: %s\n", n.NameLocation.String())
	b.WriteString("Body:\n")
	b.WriteString(indentStart(n.Body.String(), 2))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "Exported: %t\n", n.Exported)
	fmt.Fprintf(&b, "HasErrors: %t\n", n.HasErrors)

	return b.String()
}

func (n AstStatWhile) String() string {
	var b strings.Builder

	b.WriteString("StatWhile\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Condition:\n")
	b.WriteString(indentStart(n.Condition.String(), 2))
	b.WriteByte('\n')

	if n.Body != nil {
		b.WriteString("Body:\n")
		b.WriteString(indentStart(n.Body.String(), 2))
		b.WriteByte('\n')
	}

	fmt.Fprintf(&b, "HasDo: %t\n", n.HasDo)
	fmt.Fprintf(&b, "DoLocation: %s\n", n.DoLocation.String())

	return b.String()
}

func (n AstTableIndexer) String() string {
	var b strings.Builder

	b.WriteString("TableIndexer\n")
	fmt.Fprintf(&b, "Location: %s\n", n.Location.String())
	b.WriteString("IndexType:\n")
	b.WriteString(indentStart(n.IndexType.String(), 2))
	b.WriteByte('\n')
	b.WriteString("ResultType:\n")
	b.WriteString(indentStart(n.ResultType.String(), 2))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "Access: %q\n", n.Access)

	if n.AccessLocation != nil {
		fmt.Fprintf(&b, "AccessLocation: %s\n", n.AccessLocation.String())
	}

	return b.String()
}

func (n AstTableProp) String() string {
	var b strings.Builder

	b.WriteString("TableProp\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Name: %q\n", n.Name.Value)
	b.WriteString("Type:\n")
	b.WriteString(indentStart(n.Type.String(), 2))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "Access: %q\n", n.Access)

	if n.AccessLocation != nil {
		fmt.Fprintf(&b, "AccessLocation: %s\n", n.AccessLocation.String())
	}

	return b.String()
}

func (n AstTypeError) String() string {
	var b strings.Builder

	b.WriteString("TypeError\n")
	b.WriteString(n.NodeLoc.String())

	if len(n.Types) > 0 {
		b.WriteString("Types:\n")

		for _, ty := range n.Types {
			b.WriteString(indentStart(ty.String(), 2))
			b.WriteByte('\n')
		}
	}

	fmt.Fprintf(&b, "IsMissing: %t\n", n.IsMissing)
	fmt.Fprintf(&b, "MessageIndex: %d\n", n.MessageIndex)

	return b.String()
}

func (n AstTypeFunction) String() string {
	var b strings.Builder

	b.WriteString("TypeFunction\n")
	b.WriteString(n.NodeLoc.String())

	if len(n.Attributes) > 0 {
		b.WriteString("Attributes:\n")

		for _, attr := range n.Attributes {
			b.WriteString(indentStart(attr.String(), 2))
			b.WriteByte('\n')
		}
	}

	if len(n.Generics) > 0 {
		b.WriteString("Generics:\n")

		for _, generic := range n.Generics {
			b.WriteString(indentStart(generic.String(), 2))
			b.WriteByte('\n')
		}
	}

	if len(n.GenericPacks) > 0 {
		b.WriteString("GenericPacks:\n")

		for _, genericPack := range n.GenericPacks {
			b.WriteString(indentStart(genericPack.String(), 2))
			b.WriteByte('\n')
		}
	}

	b.WriteString("ArgTypes:\n")
	b.WriteString(indentStart(n.ArgTypes.String(), 2))
	b.WriteByte('\n')

	if len(n.ArgNames) > 0 {
		b.WriteString("ArgNames:\n")

		for _, name := range n.ArgNames {
			if name != nil {
				b.WriteString(indentStart(name.String(), 2))
				b.WriteByte('\n')
			}
		}
	}

	b.WriteString("ReturnTypes:\n")
	b.WriteString(indentStart(n.ReturnTypes.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstTypeGroup) String() string {
	var b strings.Builder

	b.WriteString("TypeGroup\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Type:\n")
	b.WriteString(indentStart(n.Type.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

func (n AstTypeIntersection) String() string {
	var b strings.Builder

	b.WriteString("TypeIntersection\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Types:\n")

	for _, ty := range n.Types {
		b.WriteString(indentStart(ty.String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

func (n AstTypeList) String() string {
	var b strings.Builder

	b.WriteString("TypeList\n")

	if len(n.Types) > 0 {
		b.WriteString("Types:\n")

		for _, ty := range n.Types {
			b.WriteString(indentStart(ty.String(), 2))
			b.WriteByte('\n')
		}
	}

	if n.TailType != nil {
		b.WriteString("TailType:\n")
		b.WriteString(indentStart((*n.TailType).String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

func (n AstTypeOptional) String() string {
	var b strings.Builder

	b.WriteString("TypeOptional\n")
	b.WriteString(n.NodeLoc.String())

	return b.String()
}

func (n AstTypeOrPack) String() string {
	var b strings.Builder

	b.WriteString("TypeOrPack\n")

	if n.Type != nil {
		b.WriteString("Type:\n")
		b.WriteString(indentStart((*n.Type).String(), 2))
		b.WriteByte('\n')
	}

	if n.Pack != nil {
		b.WriteString("Pack:\n")
		b.WriteString(indentStart((*n.Pack).String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

func (n AstTypePackExplicit) String() string {
	var b strings.Builder

	b.WriteString("TypePackExplicit\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Types:\n")
	b.WriteString(indentStart(n.Types.String(), 2))
	b.WriteByte('\n')

	if n.TailType != nil {
		b.WriteString("TailType:\n")
		b.WriteString(indentStart((*n.TailType).String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

func (n AstTypePackGeneric) String() string {
	var b strings.Builder

	b.WriteString("TypePackGeneric\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "GenericName: %q\n", n.GenericName)

	return b.String()
}

func (n AstTypePackVariadic) String() string {
	var b strings.Builder

	b.WriteString("TypePackVariadic\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("VariadicType:\n")
	b.WriteString(indentStart(n.VariadicType.String(), 2))

	return b.String()
}

func (n AstTypeReference) String() string {
	var b strings.Builder

	b.WriteString("TypeReference\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "HasParameterList: %t\n", n.HasParameterList)

	if n.Prefix != nil {
		fmt.Fprintf(&b, "Prefix: %q\n", *n.Prefix)
	}

	if n.PrefixLocation != nil {
		fmt.Fprintf(&b, "PrefixLocation: %s\n", n.PrefixLocation.String())
	}

	fmt.Fprintf(&b, "Name: %q\n", n.Name)
	fmt.Fprintf(&b, "NameLocation: %s\n", n.NameLocation.String())

	if len(n.Parameters) > 0 {
		b.WriteString("Parameters:\n")

		for _, param := range n.Parameters {
			b.WriteString(indentStart(param.String(), 2))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

func (n AstTypeSingletonBool) String() string {
	var b strings.Builder

	b.WriteString("TypeSingletonBool\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Value: %t\n", n.Value)

	return b.String()
}

func (n AstTypeSingletonString) String() string {
	var b strings.Builder

	b.WriteString("TypeSingletonString\n")
	b.WriteString(n.NodeLoc.String())
	fmt.Fprintf(&b, "Value: %q\n", n.Value)

	return b.String()
}

func (n AstTypeTable) String() string {
	var b strings.Builder

	b.WriteString("TypeTable\n")
	b.WriteString(n.NodeLoc.String())

	if len(n.Props) > 0 {
		b.WriteString("Props:\n")

		for _, prop := range n.Props {
			b.WriteString(indentStart(prop.String(), 2))
			b.WriteByte('\n')
		}
	}

	if n.Indexer != nil {
		b.WriteString("Indexer:\n")
		b.WriteString(indentStart(n.Indexer.String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

func (n AstTypeTypeof) String() string {
	var b strings.Builder

	b.WriteString("TypeTypeof\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Expr:\n")
	b.WriteString(indentStart(n.Expr.String(), 2))

	return b.String()
}

func (n AstTypeUnion) String() string {
	var b strings.Builder

	b.WriteString("TypeUnion\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Types:\n")

	for _, ty := range n.Types {
		b.WriteString(indentStart(ty.String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}
