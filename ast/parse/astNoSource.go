package main

import "github.com/Heliodex/coputer/ast/lex"

const (
	Ext            = ".luau"
	AstDir         = "../test/ast"
	BenchmarkDir   = "../test/benchmark"
	ConformanceDir = "../test/conformance"
)

// base for every node

type NodeLoc struct {
	lex.Location
}

func (l NodeLoc) GetLocation() lex.Location {
	return l.Location
}

// ast groops

// --------------------------------------------------------------------------------
// -- AST NODE UNION TYPE
// --------------------------------------------------------------------------------

// Union type representing all possible AST nodes that can be stored in the CST map.
// This includes statements, expressions, types, locals, and various helper nodes.

type AstNode interface {
	isAstNode()
}

// --------------------------------------------------------------------------------
// -- EXPRESSION UNION TYPES
// --------------------------------------------------------------------------------

// Union type representing all possible expression nodes in the AST.
// Expressions are constructs that evaluate to a value.
type AstExpr interface {
	AstNode
	GetLocation() lex.Location
	isAstExpr()
}

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
)

// --------------------------------------------------------------------------------
// -- STATEMENT UNION TYPES
// --------------------------------------------------------------------------------

// Union type representing all possible statement nodes in the AST.
// Statements are constructs that perform actions but don't produce values.
type AstStat interface {
	AstNode
	isAstStat()
}

var (
	_ AstStat = AstStatBlock{}
	_ AstStat = AstStatIf{}
	_ AstStat = AstStatWhile{}
	_ AstStat = AstStatRepeat{}
	_ AstStat = AstStatBreak{}
	_ AstStat = AstStatContinue{}
	_ AstStat = AstStatReturn{}
	_ AstStat = AstStatExpr{}
	_ AstStat = AstStatLocal{}
	_ AstStat = AstStatFor{}
	_ AstStat = AstStatForIn{}
	_ AstStat = AstStatAssign{}
	_ AstStat = AstStatCompoundAssign{}
	_ AstStat = AstStatFunction{}
	_ AstStat = AstStatLocalFunction{}
	_ AstStat = AstStatTypeAlias{}
	_ AstStat = AstStatTypeFunction{}
	_ AstStat = AstStatDeclareGlobal{}
	_ AstStat = AstStatDeclareFunction{}
	_ AstStat = AstStatDeclareExternType{}
	_ AstStat = AstStatError{}
)

// extra bonus

type AstStatForOrForIn interface {
	AstStat
	isAstStatForOrForIn()
}

var (
	_ AstStatForOrForIn = AstStatFor{}
	_ AstStatForOrForIn = AstStatForIn{}
)

type AstStatBreakOrError interface {
	AstStat
	isAstStatBreakOrError()
}

var (
	_ AstStatBreakOrError = AstStatBreak{}
	_ AstStatBreakOrError = AstStatError{}
)

type AstStatContinueOrError interface {
	AstStat
	isAstStatContinueOrError()
}

var (
	_ AstStatContinueOrError = AstStatContinue{}
	_ AstStatContinueOrError = AstStatError{}
)

type AstStatTypeAliasOrTypeFunction interface {
	AstStat
	isAstStatTypeAliasOrTypeFunction()
}

var (
	_ AstStatTypeAliasOrTypeFunction = AstStatTypeAlias{}
	_ AstStatTypeAliasOrTypeFunction = AstStatTypeFunction{}
)

// --------------------------------------------------------------------------------
// -- TYPE PACK UNION TYPES
// --------------------------------------------------------------------------------

// Union type representing all possible type pack nodes.
// Type packs represent multiple types, used for function return types and variadic arguments.

type AstTypePack interface {
	AstNode
	isAstTypePack()
}

var (
	_ AstTypePack = AstTypePackExplicit{}
	_ AstTypePack = AstTypePackGeneric{}
	_ AstTypePack = AstTypePackVariadic{}
)

// bonus round

type AstTypePackVariadicOrGeneric interface {
	AstTypePack
	isAstTypePackVariadicOrGeneric()
}

var (
	_ AstTypePackVariadicOrGeneric = AstTypePackGeneric{}
	_ AstTypePackVariadicOrGeneric = AstTypePackVariadic{}
)

// --------------------------------------------------------------------------------
// -- TYPE ANNOTATION UNION TYPES
// --------------------------------------------------------------------------------

// Union type representing all possible type annotation nodes.
// Type annotations specify the expected types of values in Luau's type system.

type AstType interface {
	AstNode
	isAstType()
	// isAstNodePack() 😭
}

var (
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
)

// ast

type Comment struct {
	Type lex.LexemeType
	NodeLoc
}

// node types (ok, real ast now)
type AstAttr struct {
	NodeLoc
	Type string
	Args []AstExpr
	Name *string
}

type AstArgumentName struct {
	Name     string
	Location lex.Location
}

type AstExprBinary struct {
	NodeLoc
	Op    int
	Left  AstExpr
	Right AstExpr
}

func (AstExprBinary) isAstNode() {}
func (AstExprBinary) isAstExpr() {}

type AstExprCall struct {
	NodeLoc
	Func          AstExpr
	Args          []AstExpr
	Self          bool
	ArgLocation   lex.Location
	TypeArguments *[]AstTypeOrPack
}

func (AstExprCall) isAstNode() {}
func (AstExprCall) isAstExpr() {}

type AstExprConstantBool struct {
	NodeLoc
	Value bool
}

func (AstExprConstantBool) isAstNode() {}
func (AstExprConstantBool) isAstExpr() {}

type AstExprConstantNil struct {
	NodeLoc
}

func (AstExprConstantNil) isAstNode() {}
func (AstExprConstantNil) isAstExpr() {}

type AstExprConstantNumber struct {
	NodeLoc
	Value float64
}

func (AstExprConstantNumber) isAstNode() {}
func (AstExprConstantNumber) isAstExpr() {}

type AstExprConstantString struct {
	NodeLoc
	Value string
}

func (AstExprConstantString) isAstNode() {}
func (AstExprConstantString) isAstExpr() {}

type AstExprError struct {
	NodeLoc
	Expressions  []AstExpr
	MessageIndex int
}

func (AstExprError) isAstNode() {}
func (AstExprError) isAstExpr() {}

type AstExprFunction struct {
	NodeLoc
	Attributes       []AstAttr
	Generics         []AstGenericType
	GenericPacks     []AstGenericTypePack
	Self             *AstLocal
	Args             []AstLocal
	ReturnAnnotation *AstTypePack
	Vararg           bool
	VarargLocation   lex.Location
	VarargAnnotation *AstTypePack
	Body             AstStatBlock
	FunctionDepth    int
	Debugname        string
	ArgLocation      *lex.Location
}

func (AstExprFunction) isAstNode() {}
func (AstExprFunction) isAstExpr() {}

type AstExprGlobal struct {
	NodeLoc
	Global string
}

func (AstExprGlobal) isAstNode() {}
func (AstExprGlobal) isAstExpr() {}

type AstExprGroup struct {
	NodeLoc
	Expr AstExpr
}

func (AstExprGroup) isAstNode() {}
func (AstExprGroup) isAstExpr() {}

type AstExprIfElse struct {
	NodeLoc
	Condition AstExpr
	HasThen   bool
	TrueExpr  AstExpr
	HasElse   bool
	FalseExpr AstExpr
}

func (AstExprIfElse) isAstNode() {}
func (AstExprIfElse) isAstExpr() {}

type AstExprIndexExpr struct {
	NodeLoc
	Expr  AstExpr
	Index AstExpr
}

func (AstExprIndexExpr) isAstNode() {}
func (AstExprIndexExpr) isAstExpr() {}

type AstExprIndexName struct {
	NodeLoc
	Expr          AstExpr
	Index         string
	IndexLocation lex.Location
	OpPosition    lex.Position
	Op            rune
}

func (AstExprIndexName) isAstNode() {}
func (AstExprIndexName) isAstExpr() {}

type AstExprInterpString struct {
	NodeLoc
	Strings     []string
	Expressions []AstExpr
}

func (AstExprInterpString) isAstNode() {}
func (AstExprInterpString) isAstExpr() {}

type AstExprInstantiate struct {
	NodeLoc
	Expr          AstExpr
	TypeArguments []AstTypeOrPack
}

func (AstExprInstantiate) isAstNode() {}
func (AstExprInstantiate) isAstExpr() {}

type AstExprLocal struct {
	NodeLoc
	Local   AstLocal
	Upvalue bool
}

func (AstExprLocal) isAstNode() {}
func (AstExprLocal) isAstExpr() {}

type AstExprTable struct {
	NodeLoc
	Items []AstExprTableItem
}

func (AstExprTable) isAstNode() {}
func (AstExprTable) isAstExpr() {}

type AstExprTableItem struct {
	NodeLoc
	Kind  string
	Key   *AstExpr
	Value AstExpr
}

type AstExprTypeAssertion struct {
	NodeLoc
	Expr       AstExpr
	Annotation AstType
}

func (AstExprTypeAssertion) isAstNode() {}
func (AstExprTypeAssertion) isAstExpr() {}

type AstExprVarargs struct {
	NodeLoc
}

func (AstExprVarargs) isAstNode() {}
func (AstExprVarargs) isAstExpr() {}

type AstExprUnary struct {
	NodeLoc
	Op   string
	Expr AstExpr
}

func (AstExprUnary) isAstNode() {}
func (AstExprUnary) isAstExpr() {}

type AstGenericType struct {
	NodeLoc
	Name         string
	DefaultValue *AstType
}

type AstGenericTypePack struct {
	NodeLoc
	Name         string
	DefaultValue *AstTypePack
}

type AstLocal struct {
	NodeLoc
	Name          string
	Shadow        *AstLocal
	FunctionDepth int
	LoopDepth     int
	Annotation    AstType
}

type AstStatAssign struct {
	NodeLoc
	Vars         []AstExpr
	Values       []AstExpr
	HasSemicolon *bool
}

func (AstStatAssign) isAstNode() {}
func (AstStatAssign) isAstStat() {}

type AstStatBlock struct {
	NodeLoc
	Body         []AstStat
	HasEnd       bool
	HasSemicolon *bool
}

func (AstStatBlock) isAstNode() {}
func (AstStatBlock) isAstStat() {}

type AstStatBreak struct {
	NodeLoc
}

func (AstStatBreak) isAstNode()             {}
func (AstStatBreak) isAstStat()             {}
func (AstStatBreak) isAstStatBreakOrError() {}

type AstStatCompoundAssign struct {
	NodeLoc
	Op           int
	Var          AstExpr
	Value        AstExpr
	HasSemicolon *bool
}

func (AstStatCompoundAssign) isAstNode() {}
func (AstStatCompoundAssign) isAstStat() {}

type AstStatContinue struct {
	NodeLoc
}

func (AstStatContinue) isAstNode()                {}
func (AstStatContinue) isAstStat()                {}
func (AstStatContinue) isAstStatContinueOrError() {}

type AstStatDeclareFunction struct {
	NodeLoc
	Attributes     []AstAttr
	Name           string
	NameLocation   lex.Location
	Generics       []AstGenericType
	GenericPacks   []AstGenericTypePack
	Params         AstTypeList
	ParamNames     []AstArgumentName
	Vararg         bool
	VarargLocation lex.Location
	RetTypes       AstTypePack
	HasSemicolon   *bool
}

func (AstStatDeclareFunction) isAstNode() {}
func (AstStatDeclareFunction) isAstStat() {}

type AstStatDeclareGlobal struct {
	NodeLoc
	Name         string
	NameLocation lex.Location
	Type         AstType
	HasSemicolon *bool
}

func (AstStatDeclareGlobal) isAstNode() {}
func (AstStatDeclareGlobal) isAstStat() {}

type AstStatDeclareExternType struct {
	NodeLoc
	Name         string
	SuperName    *string
	Props        []AstDeclaredExternTypeProperty
	Indexer      *AstTableIndexer
	HasSemicolon *bool
}

func (AstStatDeclareExternType) isAstNode() {}
func (AstStatDeclareExternType) isAstStat() {}

type AstDeclaredExternTypeProperty struct {
	Location     lex.Location
	Name         lex.AstName
	NameLocation lex.Location
	Ty           AstType
	IsMethod     bool
}

type AstStatError struct {
	NodeLoc
	Expressions  []AstExpr
	Statements   []AstStat
	MessageIndex int
	HasSemicolon *bool
}

func (AstStatError) isAstNode()                {}
func (AstStatError) isAstStat()                {}
func (AstStatError) isAstStatBreakOrError()    {}
func (AstStatError) isAstStatContinueOrError() {}

type AstStatExpr struct {
	NodeLoc
	Expr AstExpr
}

func (AstStatExpr) isAstNode() {}
func (AstStatExpr) isAstStat() {}

type AstStatFor struct {
	NodeLoc
	Var          AstLocal
	From         AstExpr
	To           AstExpr
	Step         *AstExpr
	Body         AstStatBlock
	HasDo        bool
	DoLocation   lex.Location
	HasSemicolon *bool
}

func (AstStatFor) isAstNode()           {}
func (AstStatFor) isAstStat()           {}
func (AstStatFor) isAstStatForOrForIn() {}

type AstStatForIn struct {
	NodeLoc
	Vars         []AstLocal
	Values       []AstExpr
	Body         AstStatBlock
	HasIn        bool
	InLocation   lex.Location
	HasDo        bool
	DoLocation   lex.Location
	HasSemicolon *bool
}

func (AstStatForIn) isAstNode()           {}
func (AstStatForIn) isAstStat()           {}
func (AstStatForIn) isAstStatForOrForIn() {}

type AstStatFunction struct {
	NodeLoc
	Name         AstExpr
	Func         AstExprFunction
	HasSemicolon *bool
}

func (AstStatFunction) isAstNode() {}
func (AstStatFunction) isAstStat() {}

type AstStatIf struct {
	NodeLoc
	Condition    AstExpr
	ThenBody     AstStatBlock
	ElseBody     *AstStat
	ThenLocation *lex.Location
	ElseLocation *lex.Location
	HasSemicolon *bool
}

func (AstStatIf) isAstNode() {}
func (AstStatIf) isAstStat() {}

type AstStatLocal struct {
	NodeLoc
	Vars               []AstLocal
	Values             []AstExpr
	EqualsSignLocation *lex.Location
	HasSemicolon       *bool
}

func (AstStatLocal) isAstNode() {}
func (AstStatLocal) isAstStat() {}

type AstStatLocalFunction struct {
	NodeLoc
	Name         AstLocal
	Func         AstExprFunction
	HasSemicolon *bool
}

func (AstStatLocalFunction) isAstNode() {}
func (AstStatLocalFunction) isAstStat() {}

type AstStatRepeat struct {
	NodeLoc
	Condition    AstExpr
	Body         AstStatBlock
	HasUntil     bool
	HasSemicolon *bool
}

func (AstStatRepeat) isAstNode() {}
func (AstStatRepeat) isAstStat() {}

type AstStatReturn struct {
	NodeLoc
	List         []AstExpr
	HasSemicolon *bool
}

func (AstStatReturn) isAstNode() {}
func (AstStatReturn) isAstStat() {}

type AstStatTypeAlias struct {
	NodeLoc
	Name         string
	NameLocation lex.Location
	Generics     []AstGenericType
	GenericPacks []AstGenericTypePack
	Type         AstType
	Exported     bool
	HasSemicolon *bool
}

func (AstStatTypeAlias) isAstNode()                        {}
func (AstStatTypeAlias) isAstStat()                        {}
func (AstStatTypeAlias) isAstStatTypeAliasOrTypeFunction() {}

type AstStatTypeFunction struct {
	NodeLoc
	Name         string
	NameLocation lex.Location
	Body         AstExprFunction
	Exported     bool
	HasErrors    bool
	HasSemicolon *bool
}

func (AstStatTypeFunction) isAstNode()                        {}
func (AstStatTypeFunction) isAstStat()                        {}
func (AstStatTypeFunction) isAstStatTypeAliasOrTypeFunction() {}

type AstStatWhile struct {
	NodeLoc
	Condition    AstExpr
	Body         AstStatBlock
	HasDo        bool
	DoLocation   lex.Location
	HasSemicolon *bool
}

func (AstStatWhile) isAstNode() {}
func (AstStatWhile) isAstStat() {}

type AstTableIndexer struct {
	Location       lex.Location
	IndexType      AstType
	ResultType     AstType
	Access         string
	AccessLocation *lex.Location
}

type AstTableProp struct {
	Name lex.AstName
	NodeLoc
	Type           AstType
	Access         string
	AccessLocation *lex.Location
}

type AstTypeError struct {
	NodeLoc
	Types        []AstType
	IsMissing    bool
	MessageIndex int
}

func (AstTypeError) isAstNode() {}
func (AstTypeError) isAstType() {}

type AstTypeFunction struct {
	NodeLoc
	Attributes   []AstAttr
	Generics     []AstGenericType
	GenericPacks []AstGenericTypePack
	ArgTypes     AstTypeList
	ArgNames     []*AstArgumentName
	ReturnTypes  AstTypePackExplicit
}

func (AstTypeFunction) isAstNode() {}
func (AstTypeFunction) isAstType() {}

type AstTypeGroup struct {
	NodeLoc
	Type AstType
}

func (AstTypeGroup) isAstNode() {}
func (AstTypeGroup) isAstType() {}

type AstTypeIntersection struct {
	NodeLoc
	Types []AstType
}

func (AstTypeIntersection) isAstNode() {}
func (AstTypeIntersection) isAstType() {}

type AstTypeList struct {
	Types    []AstType
	TailType *AstTypePack
}

func (AstTypeList) isAstNode() {}
func (AstTypeList) isAstType() {}

type AstTypeOptional struct {
	NodeLoc
}

func (AstTypeOptional) isAstNode() {}
func (AstTypeOptional) isAstType() {}

type AstTypeOrPack struct {
	Type *AstType
	Pack *AstTypePack
}

type AstTypePackExplicit struct {
	NodeLoc
	Types    AstTypeList
	TailType *AstTypePack
}

func (AstTypePackExplicit) isAstNode()     {}
func (AstTypePackExplicit) isAstTypePack() {}

type AstTypePackGeneric struct {
	NodeLoc
	GenericName string
}

func (AstTypePackGeneric) isAstNode()                      {}
func (AstTypePackGeneric) isAstTypePack()                  {}
func (AstTypePackGeneric) isAstTypePackVariadicOrGeneric() {}

type AstTypePackVariadic struct {
	NodeLoc
	VariadicType AstType
}

func (AstTypePackVariadic) isAstNode()                      {}
func (AstTypePackVariadic) isAstTypePack()                  {}
func (AstTypePackVariadic) isAstTypePackVariadicOrGeneric() {}

type AstTypeReference struct {
	NodeLoc
	HasParameterList bool
	Prefix           *string
	PrefixLocation   *lex.Location
	Name             string
	NameLocation     lex.Location
	Parameters       []AstTypeOrPack
}

func (AstTypeReference) isAstNode() {}
func (AstTypeReference) isAstType() {}

type AstTypeSingletonBool struct {
	NodeLoc
	Value bool
}

func (AstTypeSingletonBool) isAstNode() {}
func (AstTypeSingletonBool) isAstType() {}

type AstTypeSingletonString struct {
	NodeLoc
	Value string
}

func (AstTypeSingletonString) isAstNode() {}
func (AstTypeSingletonString) isAstType() {}

type AstTypeTable struct {
	NodeLoc
	Props   []AstTableProp
	Indexer *AstTableIndexer
}

func (AstTypeTable) isAstNode() {}
func (AstTypeTable) isAstType() {}

// lol
type AstTypeTypeof struct {
	NodeLoc
	Expr AstExpr
}

func (AstTypeTypeof) isAstNode() {}
func (AstTypeTypeof) isAstType() {}

type AstTypeUnion struct {
	NodeLoc
	Types []AstType
}

func (AstTypeUnion) isAstNode() {}
func (AstTypeUnion) isAstType() {}
