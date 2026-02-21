package main

import (
	"fmt"
	"strings"

	"github.com/Heliodex/coputer/ast/lex"
)

const (
	Ext            = ".luau"
	AstDir         = "../test/ast"
	BenchmarkDir   = "../test/benchmark"
	ConformanceDir = "../test/conformance"
)

// base for every node

type NodeLoc struct {
	Location lex.Location
}

func (l NodeLoc) GetLocation() lex.Location {
	return l.Location
}

func (l *NodeLoc) SetLocation(loc lex.Location) {
	l.Location = loc
}

func (l *NodeLoc) String() string {
	return fmt.Sprintf("Location: %s\n", l.Location.String())
}

// ast groops

func indentStart(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", n) + line
	}
	return strings.Join(lines, "\n")
}

// --------------------------------------------------------------------------------
// -- AST NODE UNION TYPE
// --------------------------------------------------------------------------------

// Union type representing all possible AST nodes that can be stored in the CST map.
// This includes statements, expressions, types, locals, and various helper nodes.

type AstNode interface {
	isAstNode()
	String() string
}

// --------------------------------------------------------------------------------
// -- EXPRESSION UNION TYPES
// --------------------------------------------------------------------------------

// Union type representing all possible expression nodes in the AST.
// Expressions are constructs that evaluate to a value.
type AstExpr interface {
	AstNode
	isAstExpr()
	GetLocation() lex.Location
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
	GetLocation() lex.Location
	SetLocation(loc lex.Location)
	SetHasSemicolon()
}

var tru = true

var (
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
)

// extra bonus

type AstStatForOrForIn interface {
	AstStat
	isAstStatForOrForIn()
}

var (
	_ AstStatForOrForIn = &AstStatFor{}
	_ AstStatForOrForIn = &AstStatForIn{}
)

type AstStatBreakOrError interface {
	AstStat
	isAstStatBreakOrError()
}

var (
	_ AstStatBreakOrError = &AstStatBreak{}
	_ AstStatBreakOrError = &AstStatError{}
)

type AstStatContinueOrError interface {
	AstStat
	isAstStatContinueOrError()
}

var (
	_ AstStatContinueOrError = &AstStatContinue{}
	_ AstStatContinueOrError = &AstStatError{}
)

type AstStatTypeAliasOrTypeFunction interface {
	AstStat
	isAstStatTypeAliasOrTypeFunction()
}

var (
	_ AstStatTypeAliasOrTypeFunction = &AstStatTypeAlias{}
	_ AstStatTypeAliasOrTypeFunction = &AstStatTypeFunction{}
)

type AstExprLocalOrGlobalOrError interface {
	AstExpr
	isAstExprLocalOrGlobalOrError()
}

var (
	_ AstExprLocalOrGlobalOrError = AstExprLocal{}
	_ AstExprLocalOrGlobalOrError = AstExprGlobal{}
	_ AstExprLocalOrGlobalOrError = AstExprError{}
)

type AstExprInterpStringOrError interface {
	AstExpr
	isAstExprInterpStringOrError()
}

var (
	_ AstExprInterpStringOrError = AstExprInterpString{}
	_ AstExprInterpStringOrError = AstExprError{}
)

type AstExprConstantStringOrError interface {
	AstExpr
	isAstExprConstantStringOrError()
}

var (
	_ AstExprConstantStringOrError = AstExprConstantString{}
	_ AstExprConstantStringOrError = AstExprError{}
)

type AstExprConstantNumberOrError interface {
	AstExpr
	isAstExprConstantNumberOrError()
}

var (
	_ AstExprConstantNumberOrError = AstExprConstantNumber{}
	_ AstExprConstantNumberOrError = AstExprError{}
)

// --------------------------------------------------------------------------------
// -- TYPE PACK UNION TYPES
// --------------------------------------------------------------------------------

// Union type representing all possible type pack nodes.
// Type packs represent multiple types, used for function return types and variadic arguments.

type AstTypePack interface {
	AstNode
	GetLocation() lex.Location
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
	GetLocation() lex.Location
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
	*NodeLoc
}

func (n Comment) String() string {
	var b strings.Builder

	b.WriteString("Comment\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Type: %v\n", n.Type))

	return b.String()
}

// node types (ok, real ast now)
type AstAttr struct {
	*NodeLoc
	Type string
	Args []AstExpr
	Name *string
}

func (n AstAttr) String() string {
	var b strings.Builder

	b.WriteString("Attr\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Type: %q\n", n.Type))
	if len(n.Args) > 0 {
		b.WriteString("Args:\n")
		for _, arg := range n.Args {
			b.WriteString(indentStart(arg.String(), 2))
			b.WriteByte('\n')
		}
	}
	if n.Name != nil {
		b.WriteString(fmt.Sprintf("Name: %q\n", *n.Name))
	}

	return b.String()
}

type AstArgumentName struct {
	Name     string
	Location lex.Location
}

func (n AstArgumentName) String() string {
	var b strings.Builder

	b.WriteString("ArgumentName\n")
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name))
	b.WriteString(fmt.Sprintf("Location: %s\n", n.Location.String()))

	return b.String()
}

type AstExprBinary struct {
	*NodeLoc
	Op    int
	Left  AstExpr
	Right AstExpr
}

func (AstExprBinary) isAstNode() {}
func (AstExprBinary) isAstExpr() {}
func (n AstExprBinary) String() string {
	var b strings.Builder

	b.WriteString("ExprBinary\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Op: %d\n", n.Op))
	b.WriteString("Left:\n")
	b.WriteString(indentStart(n.Left.String(), 2))
	b.WriteByte('\n')
	b.WriteString("Right:\n")
	b.WriteString(indentStart(n.Right.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

type AstExprCall struct {
	*NodeLoc
	Func          AstExpr
	Args          []AstExpr
	Self          bool
	ArgLocation   lex.Location
	TypeArguments *[]AstTypeOrPack
}

func (AstExprCall) isAstNode() {}
func (AstExprCall) isAstExpr() {}
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
	b.WriteString(fmt.Sprintf("Self: %t\n", n.Self))
	b.WriteString(fmt.Sprintf("ArgLocation: %s\n", n.ArgLocation.String()))
	if n.TypeArguments != nil {
		b.WriteString("TypeArguments:\n")
		for _, typeArg := range *n.TypeArguments {
			b.WriteString(indentStart(typeArg.String(), 2))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

type AstExprConstantBool struct {
	*NodeLoc
	Value bool
}

func (AstExprConstantBool) isAstNode() {}
func (AstExprConstantBool) isAstExpr() {}
func (n AstExprConstantBool) String() string {
	var b strings.Builder

	b.WriteString("ExprConstantBool\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Value: %t\n", n.Value))

	return b.String()
}

type AstExprConstantNil struct {
	*NodeLoc
}

func (AstExprConstantNil) isAstNode() {}
func (AstExprConstantNil) isAstExpr() {}
func (n AstExprConstantNil) String() string {
	var b strings.Builder

	b.WriteString("ExprConstantNil\n")
	b.WriteString(n.NodeLoc.String())

	return b.String()
}

type AstExprConstantNumber struct {
	*NodeLoc
	Value float64
}

func (AstExprConstantNumber) isAstNode()                      {}
func (AstExprConstantNumber) isAstExpr()                      {}
func (AstExprConstantNumber) isAstExprConstantNumberOrError() {}
func (n AstExprConstantNumber) String() string {
	var b strings.Builder

	b.WriteString("ExprConstantNumber\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Value: %f\n", n.Value))

	return b.String()
}

type AstExprConstantString struct {
	*NodeLoc
	Value string
}

func (AstExprConstantString) isAstNode()                      {}
func (AstExprConstantString) isAstExpr()                      {}
func (AstExprConstantString) isAstExprConstantStringOrError() {}
func (n AstExprConstantString) String() string {
	var b strings.Builder

	b.WriteString("ExprConstantString\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Value: %q\n", n.Value))

	return b.String()
}

type AstExprError struct {
	*NodeLoc
	Expressions  []AstExpr
	MessageIndex int
}

func (AstExprError) isAstNode()                      {}
func (AstExprError) isAstExpr()                      {}
func (AstExprError) isAstExprLocalOrGlobalOrError()  {}
func (AstExprError) isAstExprInterpStringOrError()   {}
func (AstExprError) isAstExprConstantStringOrError() {}
func (AstExprError) isAstExprConstantNumberOrError() {}
func (n AstExprError) String() string {
	var b strings.Builder

	b.WriteString("ExprError\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Expressions:\n")
	for _, expr := range n.Expressions {
		b.WriteString(indentStart(expr.String(), 2))
		b.WriteByte('\n')
	}
	b.WriteString(fmt.Sprintf("MessageIndex: %d\n", n.MessageIndex))

	return b.String()
}

type AstExprFunction struct {
	*NodeLoc
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
	b.WriteString(fmt.Sprintf("Vararg: %t\n", n.Vararg))
	b.WriteString(fmt.Sprintf("VarargLocation: %s\n", n.VarargLocation.String()))
	if n.VarargAnnotation != nil {
		b.WriteString("VarargAnnotation:\n")
		b.WriteString(indentStart((*n.VarargAnnotation).String(), 2))
		b.WriteByte('\n')
	}
	b.WriteString("Body:\n")
	b.WriteString(indentStart(n.Body.String(), 2))
	b.WriteByte('\n')
	b.WriteString(fmt.Sprintf("FunctionDepth: %d\n", n.FunctionDepth))
	b.WriteString(fmt.Sprintf("Debugname: %q\n", n.Debugname))
	if n.ArgLocation != nil {
		b.WriteString(fmt.Sprintf("ArgLocation: %s\n", n.ArgLocation.String()))
	}

	return b.String()
}

type AstExprGlobal struct {
	*NodeLoc
	Name string
}

func (AstExprGlobal) isAstNode()                     {}
func (AstExprGlobal) isAstExpr()                     {}
func (AstExprGlobal) isAstExprLocalOrGlobalOrError() {}
func (n AstExprGlobal) String() string {
	var b strings.Builder

	b.WriteString("ExprGlobal\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name))

	return b.String()
}

type AstExprGroup struct {
	*NodeLoc
	Expr AstExpr
}

func (AstExprGroup) isAstNode() {}
func (AstExprGroup) isAstExpr() {}
func (n AstExprGroup) String() string {
	var b strings.Builder

	b.WriteString("ExprGroup\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Expr:\n")
	b.WriteString(indentStart(n.Expr.String(), 2))

	return b.String()
}

type AstExprIfElse struct {
	*NodeLoc
	Condition AstExpr
	HasThen   bool
	TrueExpr  AstExpr
	HasElse   bool
	FalseExpr AstExpr
}

func (AstExprIfElse) isAstNode() {}
func (AstExprIfElse) isAstExpr() {}
func (n AstExprIfElse) String() string {
	var b strings.Builder

	b.WriteString("ExprIfElse\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Condition:\n")
	b.WriteString(indentStart(n.Condition.String(), 2))
	b.WriteByte('\n')
	b.WriteString(fmt.Sprintf("HasThen: %t\n", n.HasThen))
	b.WriteString("TrueExpr:\n")
	b.WriteString(indentStart(n.TrueExpr.String(), 2))
	b.WriteByte('\n')
	b.WriteString(fmt.Sprintf("HasElse: %t\n", n.HasElse))
	b.WriteString("FalseExpr:\n")
	b.WriteString(indentStart(n.FalseExpr.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

type AstExprIndexExpr struct {
	*NodeLoc
	Expr  AstExpr
	Index AstExpr
}

func (AstExprIndexExpr) isAstNode() {}
func (AstExprIndexExpr) isAstExpr() {}
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

type AstExprIndexName struct {
	*NodeLoc
	Expr          AstExpr
	Index         string
	IndexLocation lex.Location
	OpPosition    lex.Position
	Op            rune
}

func (AstExprIndexName) isAstNode() {}
func (AstExprIndexName) isAstExpr() {}
func (n AstExprIndexName) String() string {
	var b strings.Builder

	b.WriteString("ExprIndexName\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Expr:\n")
	b.WriteString(indentStart(n.Expr.String(), 2))
	b.WriteByte('\n')
	b.WriteString(fmt.Sprintf("Index: %q\n", n.Index))
	b.WriteString(fmt.Sprintf("IndexLocation: %s\n", n.IndexLocation.String()))
	b.WriteString(fmt.Sprintf("Op: %c\n", n.Op))

	return b.String()
}

type AstExprInterpString struct {
	*NodeLoc
	Strings     []string
	Expressions []AstExpr
}

func (AstExprInterpString) isAstNode()                    {}
func (AstExprInterpString) isAstExpr()                    {}
func (AstExprInterpString) isAstExprInterpStringOrError() {}
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

type AstExprInstantiate struct {
	*NodeLoc
	Expr          AstExpr
	TypeArguments []AstTypeOrPack
}

func (AstExprInstantiate) isAstNode() {}
func (AstExprInstantiate) isAstExpr() {}
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

type AstExprLocal struct {
	*NodeLoc
	Local   AstLocal
	Upvalue bool
}

func (AstExprLocal) isAstNode()                     {}
func (AstExprLocal) isAstExpr()                     {}
func (AstExprLocal) isAstExprLocalOrGlobalOrError() {}
func (n AstExprLocal) String() string {
	var b strings.Builder

	b.WriteString("ExprLocal\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Local:\n")
	b.WriteString(indentStart(n.Local.String(), 2))
	b.WriteByte('\n')
	b.WriteString(fmt.Sprintf("Upvalue: %t\n", n.Upvalue))

	return b.String()
}

type AstExprTable struct {
	*NodeLoc
	Items []AstExprTableItem
}

func (AstExprTable) isAstNode() {}
func (AstExprTable) isAstExpr() {}
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

type AstExprTableItem struct {
	*NodeLoc
	Kind  string
	Key   *AstExpr
	Value AstExpr
}

func (n AstExprTableItem) String() string {
	var b strings.Builder

	b.WriteString("ExprTableItem\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Kind: %q\n", n.Kind))
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

type AstExprTypeAssertion struct {
	*NodeLoc
	Expr       AstExpr
	Annotation AstType
}

func (AstExprTypeAssertion) isAstNode() {}
func (AstExprTypeAssertion) isAstExpr() {}
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

type AstExprVarargs struct {
	*NodeLoc
}

func (AstExprVarargs) isAstNode() {}
func (AstExprVarargs) isAstExpr() {}
func (n AstExprVarargs) String() string {
	var b strings.Builder

	b.WriteString("ExprVarargs\n")
	b.WriteString(n.NodeLoc.String())

	return b.String()
}

type AstExprUnary struct {
	*NodeLoc
	Op   UnaryOp
	Expr AstExpr
}

func (AstExprUnary) isAstNode() {}
func (AstExprUnary) isAstExpr() {}
func (n AstExprUnary) String() string {
	var b strings.Builder

	b.WriteString("ExprUnary\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Op: %v\n", n.Op))
	b.WriteString("Expr:\n")
	b.WriteString(indentStart(n.Expr.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

type AstGenericType struct {
	*NodeLoc
	Name         string
	DefaultValue *AstType
}

func (n AstGenericType) String() string {
	var b strings.Builder

	b.WriteString("GenericType\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name))
	if n.DefaultValue != nil {
		b.WriteString("DefaultValue:\n")
		b.WriteString(indentStart((*n.DefaultValue).String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

type AstGenericTypePack struct {
	*NodeLoc
	Name         string
	DefaultValue *AstTypePack
}

func (n AstGenericTypePack) String() string {
	var b strings.Builder

	b.WriteString("GenericTypePack\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name))
	if n.DefaultValue != nil {
		b.WriteString("DefaultValue:\n")
		b.WriteString(indentStart((*n.DefaultValue).String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

type AstLocal struct {
	*NodeLoc
	Name          string
	Shadow        *AstLocal
	FunctionDepth int
	LoopDepth     int
	Annotation    AstType
}

func (n AstLocal) String() string {
	var b strings.Builder

	b.WriteString("Local\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name))
	if n.Shadow != nil {
		b.WriteString(fmt.Sprintf("Shadow: %q\n", n.Shadow.Name))
	}
	b.WriteString(fmt.Sprintf("FunctionDepth: %d\n", n.FunctionDepth))
	b.WriteString(fmt.Sprintf("LoopDepth: %d\n", n.LoopDepth))
	if n.Annotation != nil {
		b.WriteString("Annotation:\n")
		b.WriteString(indentStart(n.Annotation.String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

type AstStatAssign struct {
	*NodeLoc
	Vars         []AstExpr
	Values       []AstExpr
	HasSemicolon *bool
}

func (AstStatAssign) isAstNode() {}
func (AstStatAssign) isAstStat() {}
func (n *AstStatAssign) SetHasSemicolon() {
	n.HasSemicolon = &tru
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

type AstStatBlock struct {
	*NodeLoc
	Body         []AstStat
	HasEnd       bool
	HasSemicolon *bool
}

func (AstStatBlock) isAstNode() {}
func (AstStatBlock) isAstStat() {}
func (n *AstStatBlock) SetHasSemicolon() {
	n.HasSemicolon = &tru
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
	b.WriteString(fmt.Sprintf("HasEnd: %t\n", n.HasEnd))

	return b.String()
}

type AstStatBreak struct {
	*NodeLoc
	HasSemicolon *bool
}

func (AstStatBreak) isAstNode()             {}
func (AstStatBreak) isAstStat()             {}
func (AstStatBreak) isAstStatBreakOrError() {}
func (n *AstStatBreak) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (n AstStatBreak) String() string {
	var b strings.Builder

	b.WriteString("StatBreak\n")
	b.WriteString(n.NodeLoc.String())

	return b.String()
}

type AstStatCompoundAssign struct {
	*NodeLoc
	Op           int
	Var          AstExpr
	Value        AstExpr
	HasSemicolon *bool
}

func (AstStatCompoundAssign) isAstNode() {}
func (AstStatCompoundAssign) isAstStat() {}
func (n *AstStatCompoundAssign) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (n AstStatCompoundAssign) String() string {
	var b strings.Builder

	b.WriteString("StatCompoundAssign\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Op: %d\n", n.Op))
	b.WriteString("Var:\n")
	b.WriteString(indentStart(n.Var.String(), 2))
	b.WriteByte('\n')
	b.WriteString("Value:\n")
	b.WriteString(indentStart(n.Value.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

type AstStatContinue struct {
	*NodeLoc
	HasSemicolon *bool
}

func (AstStatContinue) isAstNode()                {}
func (AstStatContinue) isAstStat()                {}
func (AstStatContinue) isAstStatContinueOrError() {}
func (n *AstStatContinue) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (n AstStatContinue) String() string {
	var b strings.Builder

	b.WriteString("StatContinue\n")
	b.WriteString(n.NodeLoc.String())

	return b.String()
}

type AstStatDeclareFunction struct {
	*NodeLoc
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
func (n *AstStatDeclareFunction) SetHasSemicolon() {
	n.HasSemicolon = &tru
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
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name))
	b.WriteString(fmt.Sprintf("NameLocation: %s\n", n.NameLocation.String()))
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
	b.WriteString(fmt.Sprintf("Vararg: %t\n", n.Vararg))
	b.WriteString(fmt.Sprintf("VarargLocation: %s\n", n.VarargLocation.String()))
	if n.RetTypes != nil {
		b.WriteString("RetTypes:\n")
		b.WriteString(indentStart(n.RetTypes.String(), 2))
		b.WriteByte('\n')
	}

	return b.String()
}

type AstStatDeclareGlobal struct {
	*NodeLoc
	Name         string
	NameLocation lex.Location
	Type         AstType
	HasSemicolon *bool
}

func (AstStatDeclareGlobal) isAstNode() {}
func (AstStatDeclareGlobal) isAstStat() {}
func (n *AstStatDeclareGlobal) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (n AstStatDeclareGlobal) String() string {
	var b strings.Builder

	b.WriteString("StatDeclareGlobal\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name))
	b.WriteString(fmt.Sprintf("NameLocation: %s\n", n.NameLocation.String()))
	b.WriteString("Type:\n")
	b.WriteString(indentStart(n.Type.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

type AstStatDeclareExternType struct {
	*NodeLoc
	Name         string
	SuperName    *string
	Props        []AstDeclaredExternTypeProperty
	Indexer      *AstTableIndexer
	HasSemicolon *bool
}

func (AstStatDeclareExternType) isAstNode() {}
func (AstStatDeclareExternType) isAstStat() {}
func (n *AstStatDeclareExternType) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (n AstStatDeclareExternType) String() string {
	var b strings.Builder

	b.WriteString("StatDeclareExternType\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name))
	if n.SuperName != nil {
		b.WriteString(fmt.Sprintf("SuperName: %q\n", *n.SuperName))
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

type AstDeclaredExternTypeProperty struct {
	Location     lex.Location
	Name         lex.AstName
	NameLocation lex.Location
	Ty           AstType
	IsMethod     bool
}

func (n AstDeclaredExternTypeProperty) String() string {
	var b strings.Builder

	b.WriteString("DeclaredExternTypeProperty\n")
	b.WriteString(fmt.Sprintf("Location: %s\n", n.Location.String()))
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name.Value))
	b.WriteString(fmt.Sprintf("NameLocation: %s\n", n.NameLocation.String()))
	b.WriteString("Ty:\n")
	b.WriteString(indentStart(n.Ty.String(), 2))
	b.WriteByte('\n')
	b.WriteString(fmt.Sprintf("IsMethod: %t\n", n.IsMethod))

	return b.String()
}

type AstStatError struct {
	*NodeLoc
	Expressions  []AstExpr
	Statements   []AstStat
	MessageIndex int
	HasSemicolon *bool
}

func (AstStatError) isAstNode()                {}
func (AstStatError) isAstStat()                {}
func (AstStatError) isAstStatBreakOrError()    {}
func (AstStatError) isAstStatContinueOrError() {}
func (n *AstStatError) SetHasSemicolon() {
	n.HasSemicolon = &tru
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
	b.WriteString(fmt.Sprintf("MessageIndex: %d\n", n.MessageIndex))

	return b.String()
}

type AstStatExpr struct {
	*NodeLoc
	Expr         AstExpr
	HasSemicolon *bool
}

func (AstStatExpr) isAstNode() {}
func (AstStatExpr) isAstStat() {}
func (n *AstStatExpr) SetHasSemicolon() {
	n.HasSemicolon = &tru
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

type AstStatFor struct {
	*NodeLoc
	Var          *AstLocal
	From         AstExpr
	To           AstExpr
	Step         AstExpr
	Body         *AstStatBlock
	HasDo        bool
	DoLocation   lex.Location
	HasSemicolon *bool
}

func (AstStatFor) isAstNode()           {}
func (AstStatFor) isAstStat()           {}
func (AstStatFor) isAstStatForOrForIn() {}
func (n *AstStatFor) SetHasSemicolon() {
	n.HasSemicolon = &tru
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
	b.WriteString(fmt.Sprintf("HasDo: %t\n", n.HasDo))
	b.WriteString(fmt.Sprintf("DoLocation: %s\n", n.DoLocation.String()))

	return b.String()
}

type AstStatForIn struct {
	*NodeLoc
	Vars         []*AstLocal
	Values       []AstExpr
	Body         *AstStatBlock
	HasIn        bool
	InLocation   lex.Location
	HasDo        bool
	DoLocation   lex.Location
	HasSemicolon *bool
}

func (AstStatForIn) isAstNode()           {}
func (AstStatForIn) isAstStat()           {}
func (AstStatForIn) isAstStatForOrForIn() {}
func (n *AstStatForIn) SetHasSemicolon() {
	n.HasSemicolon = &tru
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
	b.WriteString(fmt.Sprintf("HasIn: %t\n", n.HasIn))
	b.WriteString(fmt.Sprintf("InLocation: %s\n", n.InLocation.String()))
	b.WriteString(fmt.Sprintf("HasDo: %t\n", n.HasDo))
	b.WriteString(fmt.Sprintf("DoLocation: %s\n", n.DoLocation.String()))

	return b.String()
}

type AstStatFunction struct {
	*NodeLoc
	Name         AstExpr
	Func         AstExprFunction
	HasSemicolon *bool
}

func (AstStatFunction) isAstNode() {}
func (AstStatFunction) isAstStat() {}
func (n *AstStatFunction) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (n AstStatFunction) String() string {
	var b strings.Builder

	b.WriteString("StatFunction\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Name:\n")
	b.WriteString(indentStart(n.Name.String(), 2))
	b.WriteByte('\n')
	b.WriteString("Func:\n")
	b.WriteString(indentStart(n.Func.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

type AstStatIf struct {
	*NodeLoc
	Condition    AstExpr
	ThenBody     AstStatBlock
	ElseBody     AstStat
	ThenLocation *lex.Location
	ElseLocation *lex.Location
	HasSemicolon *bool
}

func (AstStatIf) isAstNode() {}
func (AstStatIf) isAstStat() {}
func (n *AstStatIf) SetHasSemicolon() {
	n.HasSemicolon = &tru
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
		b.WriteString(fmt.Sprintf("ThenLocation: %s\n", n.ThenLocation.String()))
	}
	if n.ElseLocation != nil {
		b.WriteString(fmt.Sprintf("ElseLocation: %s\n", n.ElseLocation.String()))
	}

	return b.String()
}

type AstStatLocal struct {
	*NodeLoc
	Vars               []AstLocal
	Values             []AstExpr
	EqualsSignLocation *lex.Location
	HasSemicolon       *bool
}

func (AstStatLocal) isAstNode() {}
func (AstStatLocal) isAstStat() {}
func (n *AstStatLocal) SetHasSemicolon() {
	n.HasSemicolon = &tru
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
		b.WriteString(fmt.Sprintf("EqualsSignLocation: %s\n", n.EqualsSignLocation.String()))
	}

	return b.String()
}

type AstStatLocalFunction struct {
	*NodeLoc
	Name         AstLocal
	Func         AstExprFunction
	HasSemicolon *bool
}

func (AstStatLocalFunction) isAstNode() {}
func (AstStatLocalFunction) isAstStat() {}
func (n *AstStatLocalFunction) SetHasSemicolon() {
	n.HasSemicolon = &tru
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

type AstStatRepeat struct {
	*NodeLoc
	Condition    AstExpr
	Body         *AstStatBlock
	HasUntil     bool
	HasSemicolon *bool
}

func (AstStatRepeat) isAstNode() {}
func (AstStatRepeat) isAstStat() {}
func (n *AstStatRepeat) SetHasSemicolon() {
	n.HasSemicolon = &tru
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
	b.WriteString(fmt.Sprintf("HasUntil: %t\n", n.HasUntil))

	return b.String()
}

type AstStatReturn struct {
	*NodeLoc
	List         []AstExpr
	HasSemicolon *bool
}

func (AstStatReturn) isAstNode() {}
func (AstStatReturn) isAstStat() {}
func (n *AstStatReturn) SetHasSemicolon() {
	n.HasSemicolon = &tru
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

type AstStatTypeAlias struct {
	*NodeLoc
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
func (n *AstStatTypeAlias) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (n AstStatTypeAlias) String() string {
	var b strings.Builder

	b.WriteString("StatTypeAlias\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name))
	b.WriteString(fmt.Sprintf("NameLocation: %s\n", n.NameLocation.String()))
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
	b.WriteString(fmt.Sprintf("Exported: %t\n", n.Exported))

	return b.String()
}

type AstStatTypeFunction struct {
	*NodeLoc
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
func (n *AstStatTypeFunction) SetHasSemicolon() {
	n.HasSemicolon = &tru
}

func (n AstStatTypeFunction) String() string {
	var b strings.Builder

	b.WriteString("StatTypeFunction\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name))
	b.WriteString(fmt.Sprintf("NameLocation: %s\n", n.NameLocation.String()))
	b.WriteString("Body:\n")
	b.WriteString(indentStart(n.Body.String(), 2))
	b.WriteByte('\n')
	b.WriteString(fmt.Sprintf("Exported: %t\n", n.Exported))
	b.WriteString(fmt.Sprintf("HasErrors: %t\n", n.HasErrors))

	return b.String()
}

type AstStatWhile struct {
	*NodeLoc
	Condition    AstExpr
	Body         *AstStatBlock
	HasDo        bool
	DoLocation   lex.Location
	HasSemicolon *bool
}

func (AstStatWhile) isAstNode() {}
func (AstStatWhile) isAstStat() {}
func (n *AstStatWhile) SetHasSemicolon() {
	n.HasSemicolon = &tru
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
	b.WriteString(fmt.Sprintf("HasDo: %t\n", n.HasDo))
	b.WriteString(fmt.Sprintf("DoLocation: %s\n", n.DoLocation.String()))

	return b.String()
}

type AstTableIndexer struct {
	Location       lex.Location
	IndexType      AstType
	ResultType     AstType
	Access         string
	AccessLocation *lex.Location
}

func (n AstTableIndexer) String() string {
	var b strings.Builder

	b.WriteString("TableIndexer\n")
	b.WriteString(fmt.Sprintf("Location: %s\n", n.Location.String()))
	b.WriteString("IndexType:\n")
	b.WriteString(indentStart(n.IndexType.String(), 2))
	b.WriteByte('\n')
	b.WriteString("ResultType:\n")
	b.WriteString(indentStart(n.ResultType.String(), 2))
	b.WriteByte('\n')
	b.WriteString(fmt.Sprintf("Access: %q\n", n.Access))
	if n.AccessLocation != nil {
		b.WriteString(fmt.Sprintf("AccessLocation: %s\n", n.AccessLocation.String()))
	}

	return b.String()
}

type AstTableProp struct {
	Name lex.AstName
	*NodeLoc
	Type           AstType
	Access         string
	AccessLocation *lex.Location
}

func (n AstTableProp) String() string {
	var b strings.Builder

	b.WriteString("TableProp\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name.Value))
	b.WriteString("Type:\n")
	b.WriteString(indentStart(n.Type.String(), 2))
	b.WriteByte('\n')
	b.WriteString(fmt.Sprintf("Access: %q\n", n.Access))
	if n.AccessLocation != nil {
		b.WriteString(fmt.Sprintf("AccessLocation: %s\n", n.AccessLocation.String()))
	}

	return b.String()
}

type AstTypeError struct {
	*NodeLoc
	Types        []AstType
	IsMissing    bool
	MessageIndex int
}

func (AstTypeError) isAstNode() {}
func (AstTypeError) isAstType() {}
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
	b.WriteString(fmt.Sprintf("IsMissing: %t\n", n.IsMissing))
	b.WriteString(fmt.Sprintf("MessageIndex: %d\n", n.MessageIndex))

	return b.String()
}

type AstTypeFunction struct {
	*NodeLoc
	Attributes   []AstAttr
	Generics     []AstGenericType
	GenericPacks []AstGenericTypePack
	ArgTypes     AstTypeList
	ArgNames     []*AstArgumentName
	ReturnTypes  AstTypePackExplicit
}

func (AstTypeFunction) isAstNode() {}
func (AstTypeFunction) isAstType() {}
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

type AstTypeGroup struct {
	*NodeLoc
	Type AstType
}

func (AstTypeGroup) isAstNode() {}
func (AstTypeGroup) isAstType() {}
func (n AstTypeGroup) String() string {
	var b strings.Builder

	b.WriteString("TypeGroup\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Type:\n")
	b.WriteString(indentStart(n.Type.String(), 2))
	b.WriteByte('\n')

	return b.String()
}

type AstTypeIntersection struct {
	*NodeLoc
	Types []AstType
}

func (AstTypeIntersection) isAstNode() {}
func (AstTypeIntersection) isAstType() {}
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

type AstTypeList struct {
	Types    []AstType
	TailType *AstTypePack
}

func (AstTypeList) isAstNode() {}
func (AstTypeList) isAstType() {}
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

type AstTypeOptional struct {
	*NodeLoc
}

func (AstTypeOptional) isAstNode() {}
func (AstTypeOptional) isAstType() {}
func (n AstTypeOptional) String() string {
	var b strings.Builder

	b.WriteString("TypeOptional\n")
	b.WriteString(n.NodeLoc.String())

	return b.String()
}

type AstTypeOrPack struct {
	Type *AstType
	Pack *AstTypePack
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

type AstTypePackExplicit struct {
	*NodeLoc
	Types    AstTypeList
	TailType *AstTypePack
}

func (AstTypePackExplicit) isAstNode()     {}
func (AstTypePackExplicit) isAstTypePack() {}
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

type AstTypePackGeneric struct {
	*NodeLoc
	GenericName string
}

func (AstTypePackGeneric) isAstNode()                      {}
func (AstTypePackGeneric) isAstTypePack()                  {}
func (AstTypePackGeneric) isAstTypePackVariadicOrGeneric() {}
func (n AstTypePackGeneric) String() string {
	var b strings.Builder

	b.WriteString("TypePackGeneric\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("GenericName: %q\n", n.GenericName))

	return b.String()
}

type AstTypePackVariadic struct {
	*NodeLoc
	VariadicType AstType
}

func (AstTypePackVariadic) isAstNode()                      {}
func (AstTypePackVariadic) isAstTypePack()                  {}
func (AstTypePackVariadic) isAstTypePackVariadicOrGeneric() {}
func (n AstTypePackVariadic) String() string {
	var b strings.Builder

	b.WriteString("TypePackVariadic\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("VariadicType:\n")
	b.WriteString(indentStart(n.VariadicType.String(), 2))

	return b.String()
}

type AstTypeReference struct {
	*NodeLoc
	HasParameterList bool
	Prefix           *string
	PrefixLocation   *lex.Location
	Name             string
	NameLocation     lex.Location
	Parameters       []AstTypeOrPack
}

func (AstTypeReference) isAstNode() {}
func (AstTypeReference) isAstType() {}
func (n AstTypeReference) String() string {
	var b strings.Builder

	b.WriteString("TypeReference\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("HasParameterList: %t\n", n.HasParameterList))
	if n.Prefix != nil {
		b.WriteString(fmt.Sprintf("Prefix: %q\n", *n.Prefix))
	}
	if n.PrefixLocation != nil {
		b.WriteString(fmt.Sprintf("PrefixLocation: %s\n", n.PrefixLocation.String()))
	}
	b.WriteString(fmt.Sprintf("Name: %q\n", n.Name))
	b.WriteString(fmt.Sprintf("NameLocation: %s\n", n.NameLocation.String()))
	if len(n.Parameters) > 0 {
		b.WriteString("Parameters:\n")
		for _, param := range n.Parameters {
			b.WriteString(indentStart(param.String(), 2))
			b.WriteByte('\n')
		}
	}

	return b.String()
}

type AstTypeSingletonBool struct {
	*NodeLoc
	Value bool
}

func (AstTypeSingletonBool) isAstNode() {}
func (AstTypeSingletonBool) isAstType() {}
func (n AstTypeSingletonBool) String() string {
	var b strings.Builder

	b.WriteString("TypeSingletonBool\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Value: %t\n", n.Value))

	return b.String()
}

type AstTypeSingletonString struct {
	*NodeLoc
	Value string
}

func (AstTypeSingletonString) isAstNode() {}
func (AstTypeSingletonString) isAstType() {}
func (n AstTypeSingletonString) String() string {
	var b strings.Builder

	b.WriteString("TypeSingletonString\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString(fmt.Sprintf("Value: %q\n", n.Value))

	return b.String()
}

type AstTypeTable struct {
	*NodeLoc
	Props   []AstTableProp
	Indexer *AstTableIndexer
}

func (AstTypeTable) isAstNode() {}
func (AstTypeTable) isAstType() {}
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

// lol
type AstTypeTypeof struct {
	*NodeLoc
	Expr AstExpr
}

func (AstTypeTypeof) isAstNode() {}
func (AstTypeTypeof) isAstType() {}
func (n AstTypeTypeof) String() string {
	var b strings.Builder

	b.WriteString("TypeTypeof\n")
	b.WriteString(n.NodeLoc.String())
	b.WriteString("Expr:\n")
	b.WriteString(indentStart(n.Expr.String(), 2))

	return b.String()
}

type AstTypeUnion struct {
	*NodeLoc
	Types []AstType
}

func (AstTypeUnion) isAstNode() {}
func (AstTypeUnion) isAstType() {}
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
