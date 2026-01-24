package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	Ext            = ".luau"
	AstDir         = "../test/ast"
	BenchmarkDir   = "../test/benchmark"
	ConformanceDir = "../test/conformance"
)

func indentStart(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", n) + line
	}
	return strings.Join(lines, "\n")
}

type Node interface {
	String() string
}

func IndentSize(indent int) string {
	return strings.Repeat("\t", indent)
}

type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func (p Position) After(op Position) bool {
	return p.Line > op.Line || (p.Line == op.Line && p.Column > op.Column)
}

func (p Position) String() string {
	return fmt.Sprintf("%d,%d", p.Line, p.Column)
}

// Location represents a source location range
type Location struct {
	Begin Position `json:"start"`
	End   Position `json:"end"`
}

func (l Location) Contains(ol Location) bool {
	return ol.Begin.After(l.Begin) && l.End.After(ol.End)
}

type AstName struct {
	Value string `json:"value"`
}

// infinite gfs (or, well, 9223372036854775807)
// var gfsCount int

// we've pretty much eliminated calls to this function lel, it'll always be needed for comments though
func (l Location) GetFromSource(source string) (string, error) {
	lines := strings.Split(source, "\n")
	if l.Begin.Line < 0 || l.End.Line >= len(lines) {
		return "", errors.New("location out of bounds")
	}

	var b strings.Builder
	for i := l.Begin.Line; i <= l.End.Line; i++ {
		line := lines[i]
		if i == l.Begin.Line && i == l.End.Line {
			line = line[l.Begin.Column:l.End.Column]
		} else if i == l.Begin.Line {
			line = line[l.Begin.Column:]
		} else if i == l.End.Line {
			line = line[:min(l.End.Column, len(line))]
		}
		b.WriteString(line)
		if i < l.End.Line {
			b.WriteString("\n")
		}
	}

	// gfsCount++
	// fmt.Println("gotFromSource", gfsCount)
	return b.String(), nil
}

func (l Location) String() string {
	return fmt.Sprintf("%s - %s", l.Begin, l.End)
}

// UnmarshalJSON custom unmarshaler for location strings like "0,0 - 2,0"
func (l *Location) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	_, err := fmt.Sscanf(s, "%d,%d - %d,%d", &l.Begin.Line, &l.Begin.Column, &l.End.Line, &l.End.Column)
	return err
}

// base for every node

type NodeLoc struct {
	Location Location `json:"location"`
}

// ast groops

type AstNode interface {
	isAstNode()
}

// --------------------------------------------------------------------------------
// -- EXPRESSION UNION TYPES
// --------------------------------------------------------------------------------

// --- Union type representing all possible expression nodes in the AST.
// --- Expressions are constructs that evaluate to a value.
type AstExpr interface {
	AstNode
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

// --- Union type representing all possible statement nodes in the AST.
// --- Statements are constructs that perform actions but don't produce values.
type AstStat interface {
	AstNode
	isAstNode()
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

// --------------------------------------------------------------------------------
// -- TYPE PACK UNION TYPES
// --------------------------------------------------------------------------------

// --- Union type representing all possible type pack nodes.
// --- Type packs represent multiple types, used for function return types and variadic arguments.

type AstTypePack interface {
	AstNode
	isAstNodePack()
	isAstTypePack()
}

var (
	_ AstTypePack = AstTypePackExplicit{}
	_ AstTypePack = AstTypePackGeneric{}
	_ AstTypePack = AstTypePackVariadic{}
)

// --------------------------------------------------------------------------------
// -- TYPE ANNOTATION UNION TYPES
// --------------------------------------------------------------------------------

// --- Union type representing all possible type annotation nodes.
// --- Type annotations specify the expected types of values in Luau's type system.

type AstType interface {
	AstNode
	isAstNode()
	isAstType()
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
	Type int `json:"type"`
	NodeLoc
}

// node types (ok, real ast now)

type AstArgumentName struct {
	Name     string   `json:"name"`
	Location Location `json:"location"`
}

type AstAttr struct {
	NodeLoc
	Name string `json:"name"`
}

type AstExprBinary struct {
	NodeLoc
	Op    int     `json:"op"`
	Left  AstExpr `json:"left"`
	Right AstExpr `json:"right"`
}

func (n AstExprBinary) isAstNode() {}
func (n AstExprBinary) isAstExpr() {}

type AstExprCall struct {
	NodeLoc
	Func          AstExpr          `json:"func"`
	Args          []AstExpr        `json:"args"`
	Self          bool             `json:"self"`
	ArgLocation   Location         `json:"argLocation"`
	TypeArguments *[]AstTypeOrPack `json:"typeArguments"`
}

func (n AstExprCall) isAstNode() {}
func (n AstExprCall) isAstExpr() {}

type AstExprConstantBool struct {
	NodeLoc
	Value bool `json:"value"`
}

func (n AstExprConstantBool) isAstNode() {}
func (n AstExprConstantBool) isAstExpr() {}

type AstExprConstantNil struct {
	NodeLoc
}

func (n AstExprConstantNil) isAstNode() {}
func (n AstExprConstantNil) isAstExpr() {}

type AstExprConstantNumber struct {
	NodeLoc
	Value float64 `json:"value"`
}

func (n AstExprConstantNumber) isAstNode() {}
func (n AstExprConstantNumber) isAstExpr() {}

type AstExprConstantString struct {
	NodeLoc
	Value string `json:"value"`
}

func (n AstExprConstantString) isAstNode() {}
func (n AstExprConstantString) isAstExpr() {}

type AstExprError struct {
	NodeLoc
	Expressions  []AstExpr `json:"expressions"`
	MessageIndex int       `json:"messageIndex"`
}

func (n AstExprError) isAstNode() {}
func (n AstExprError) isAstExpr() {}

type AstExprFunction struct {
	NodeLoc
	Attributes       []AstAttr            `json:"attributes"`
	Generics         []AstGenericType     `json:"generics"`
	GenericPacks     []AstGenericTypePack `json:"genericPacks"`
	Self             *AstLocal            `json:"self"`
	Args             []AstLocal           `json:"args"`
	ReturnAnnotation *AstTypePack         `json:"returnAnnotation"`
	Vararg           bool                 `json:"vararg"`
	VarargLocation   Location             `json:"varargLocation"`
	VarargAnnotation *AstTypePack         `json:"varargAnnotation"`
	Body             AstStatBlock         `json:"body"`
	FunctionDepth    int                  `json:"functionDepth"`
	Debugname        string               `json:"debugname"`
	ArgLocation      *Location            `json:"argLocation"`
}

func (n AstExprFunction) isAstNode() {}
func (n AstExprFunction) isAstExpr() {}

type AstExprGlobal struct {
	NodeLoc
	Global string `json:"global"`
}

func (n AstExprGlobal) isAstNode() {}
func (n AstExprGlobal) isAstExpr() {}

type AstExprGroup struct {
	NodeLoc
	Expr AstExpr `json:"expr"` // only contains one expression right? strange when you first think about it
}

func (n AstExprGroup) isAstNode() {}
func (n AstExprGroup) isAstExpr() {}

type AstExprIfElse struct {
	NodeLoc
	Condition AstExpr `json:"condition"`
	HasThen   bool    `json:"hasThen"`
	TrueExpr  AstExpr `json:"trueExpr"`
	HasElse   bool    `json:"hasElse"`
	FalseExpr AstExpr `json:"falseExpr"`
}

func (n AstExprIfElse) isAstNode() {}
func (n AstExprIfElse) isAstExpr() {}

type AstExprIndexExpr struct {
	NodeLoc
	Expr  AstExpr `json:"expr"`
	Index AstExpr `json:"index"`
}

func (n AstExprIndexExpr) isAstNode() {}
func (n AstExprIndexExpr) isAstExpr() {}

type AstExprIndexName struct {
	NodeLoc
	Expr          AstExpr  `json:"expr"`
	Index         string   `json:"index"`
	IndexLocation Location `json:"indexLocation"`
	OpPosition    Position `json:"opPosition"`
	Op            rune     `json:"op"`
}

func (n AstExprIndexName) isAstNode() {}
func (n AstExprIndexName) isAstExpr() {}

type AstExprInterpString struct {
	NodeLoc
	Strings     []string  `json:"strings"`
	Expressions []AstExpr `json:"expressions"`
}

func (n AstExprInterpString) isAstNode() {}
func (n AstExprInterpString) isAstExpr() {}

type AstExprInstantiate struct {
	NodeLoc
	Expr          AstExpr         `json:"expr"`
	TypeArguments []AstTypeOrPack `json:"typeArguments"`
}

func (n AstExprInstantiate) isAstNode() {}
func (n AstExprInstantiate) isAstExpr() {}

type AstExprLocal struct {
	NodeLoc
	Local   AstLocal `json:"local"`
	Upvalue bool     `json:"upvalue"`
}

func (n AstExprLocal) isAstNode() {}
func (n AstExprLocal) isAstExpr() {}

type AstExprTable struct {
	NodeLoc
	Items []AstExprTableItem `json:"items"`
}

func (n AstExprTable) isAstNode() {}
func (n AstExprTable) isAstExpr() {}

type AstExprTableItem struct {
	NodeLoc
	Kind  string   `json:"kind"` // "List" | "Record" | "General"
	Key   *AstExpr `json:"key"`
	Value AstExpr  `json:"value"`
}

type AstExprTypeAssertion struct {
	NodeLoc
	Expr       AstExpr `json:"expr"`
	Annotation AstType `json:"annotation"`
}

func (n AstExprTypeAssertion) isAstNode() {}
func (n AstExprTypeAssertion) isAstExpr() {}

type AstExprVarargs struct {
	NodeLoc
}

func (n AstExprVarargs) isAstNode() {}
func (n AstExprVarargs) isAstExpr() {}

type AstExprUnary struct {
	NodeLoc
	Op   string  `json:"op"`
	Expr AstExpr `json:"expr"`
}

func (n AstExprUnary) isAstNode() {}
func (n AstExprUnary) isAstExpr() {}

type AstGenericType struct {
	NodeLoc
	Name         string   `json:"name"`
	DefaultValue *AstType `json:"defaultValue"`
}

type AstGenericTypePack struct {
	NodeLoc
	Name string `json:"name"`
}

type AstLocal struct {
	NodeLoc
	Name          string    `json:"name"`
	Shadow        *AstLocal `json:"shadow"`
	FunctionDepth int       `json:"functionDepth"`
	LoopDepth     int       `json:"loopDepth"`
	Annotation    *AstType  `json:"annotation"`
}

type AstStatAssign struct {
	NodeLoc
	Vars         []AstExpr `json:"vars"`
	Values       []AstExpr `json:"values"`
	HasSemicolon *bool     `json:"hasSemicolon"`
}

func (n AstStatAssign) isAstNode() {}
func (n AstStatAssign) isAstStat() {}

type AstStatBlock struct {
	NodeLoc
	HasEnd            bool       `json:"hasEnd"`
	Body              []AstStat  `json:"body"`
	CommentsContained *[]Comment // not in json
}

func (n AstStatBlock) isAstNode() {}
func (n AstStatBlock) isAstStat() {}

type AstStatBreak struct {
	NodeLoc
}

func (n AstStatBreak) isAstNode() {}
func (n AstStatBreak) isAstStat() {}

type AstStatCompoundAssign struct {
	NodeLoc
	Op           int     `json:"op"`
	Var          AstExpr `json:"var"`
	Value        AstExpr `json:"value"`
	HasSemicolon *bool   `json:"hasSemicolon"`
}

func (n AstStatCompoundAssign) isAstNode() {}
func (n AstStatCompoundAssign) isAstStat() {}

type AstStatContinue struct {
	NodeLoc
}

func (n AstStatContinue) isAstNode() {}
func (n AstStatContinue) isAstStat() {}

type AstStatDeclareFunction struct {
	NodeLoc
	Attributes     []AstAttr            `json:"attributes"`
	Name           string               `json:"name"`
	NameLocation   Location             `json:"nameLocation"`
	Generics       []AstGenericType     `json:"generics"`
	GenericPacks   []AstGenericTypePack `json:"genericPacks"`
	Params         AstTypeList          `json:"params"`
	ParamNames     []AstArgumentName    `json:"paramNames"`
	Vararg         bool                 `json:"vararg"`
	VarargLocation Location             `json:"varargLocation"`
	RetTypes       AstTypePack          `json:"retTypes"`
	HasSemicolon   *bool                `json:"hasSemicolon"`
}

func (n AstStatDeclareFunction) isAstNode() {}
func (n AstStatDeclareFunction) isAstStat() {}

type AstStatDeclareGlobal struct {
	NodeLoc
	Name         string   `json:"name"`
	NameLocation Location `json:"nameLocation"`
	Type         AstType  `json:"type"`
	HasSemicolon *bool    `json:"hasSemicolon"`
}

func (n AstStatDeclareGlobal) isAstNode() {}
func (n AstStatDeclareGlobal) isAstStat() {}

type AstStatDeclareExternType struct {
	NodeLoc
	Name         string                          `json:"name"`
	SuperName    *string                         `json:"superName"`
	Props        []AstDeclaredExternTypeProperty `json:"props"`
	Indexer      *AstTableIndexer                `json:"indexer"`
	HasSemicolon *bool                           `json:"hasSemicolon"`
}

func (n AstStatDeclareExternType) isAstNode() {}
func (n AstStatDeclareExternType) isAstStat() {}

type AstDeclaredExternTypeProperty struct {
	Location     Location `json:"location"`
	Name         AstName  `json:"name"`
	NameLocation Location `json:"nameLocation"`
	Ty           AstType  `json:"type"`
	IsMethod     bool     `json:"isMethod"`
}

type AstStatError struct {
	NodeLoc
	Messages     []string  `json:"messages"`
	Expressions  []AstExpr `json:"expressions"`
	Statements   []AstStat `json:"statements"`
	MessageIndex int       `json:"messageIndex"`
	HasSemicolon *bool     `json:"hasSemicolon"`
}

func (n AstStatError) isAstNode() {}
func (n AstStatError) isAstStat() {}

type AstStatExpr struct {
	NodeLoc
	Expr AstExpr `json:"expr"`
}

func (n AstStatExpr) isAstNode() {}
func (n AstStatExpr) isAstStat() {}

type AstStatFor struct {
	NodeLoc
	Var          AstLocal     `json:"var"`
	From         AstExpr      `json:"from"`
	To           AstExpr      `json:"to"`
	Step         *AstExpr     `json:"step"`
	Body         AstStatBlock `json:"body"`
	HasDo        bool         `json:"hasDo"`
	DoLocation   Location     `json:"doLocation"`
	HasSemicolon *bool        `json:"hasSemicolon"`
}

func (n AstStatFor) isAstNode() {}
func (n AstStatFor) isAstStat() {}

type AstStatForIn struct {
	NodeLoc
	Vars         []AstLocal   `json:"vars"`
	Values       []AstExpr    `json:"values"`
	Body         AstStatBlock `json:"body"`
	HasIn        bool         `json:"hasIn"`
	InLocation   Location     `json:"inLocation"`
	HasDo        bool         `json:"hasDo"`
	DoLocation   Location     `json:"doLocation"`
	HasSemicolon *bool        `json:"hasSemicolon"`
}

func (n AstStatForIn) isAstNode() {}
func (n AstStatForIn) isAstStat() {}

type AstStatFunction struct {
	NodeLoc
	Name         AstExpr         `json:"name"`
	Func         AstExprFunction `json:"func"`
	HasSemicolon *bool           `json:"hasSemicolon"`
}

func (n AstStatFunction) isAstNode() {}
func (n AstStatFunction) isAstStat() {}

type AstStatIf struct {
	NodeLoc
	Condition    AstExpr      `json:"condition"`
	ThenBody     AstStatBlock `json:"thenbody"`
	ElseBody     *AstStat     `json:"elsebody"` // StatBlock | StatIf
	ThenLocation *Location    `json:"thenLocation"`
	ElseLocation *Location    `json:"elseLocation"`
	HasSemicolon *bool        `json:"hasSemicolon"`
}

func (n AstStatIf) isAstNode() {}
func (n AstStatIf) isAstStat() {}

type AstStatLocal struct {
	NodeLoc
	Vars               []AstLocal `json:"vars"`
	Values             []AstExpr  `json:"values"`
	EqualsSignLocation *Location  `json:"equalsSignLocation"`
	HasSemicolon       *bool      `json:"hasSemicolon"`
}

func (n AstStatLocal) isAstNode() {}
func (n AstStatLocal) isAstStat() {}

type AstStatLocalFunction struct {
	NodeLoc
	Name         AstLocal        `json:"name"`
	Func         AstExprFunction `json:"func"`
	HasSemicolon *bool           `json:"hasSemicolon"`
}

func (n AstStatLocalFunction) isAstNode() {}
func (n AstStatLocalFunction) isAstStat() {}

type AstStatRepeat struct {
	NodeLoc
	Condition    AstExpr      `json:"condition"`
	Body         AstStatBlock `json:"body"`
	HasUntil     bool         `json:"hasUntil"`
	HasSemicolon *bool        `json:"hasSemicolon"`
}

func (n AstStatRepeat) isAstNode() {}
func (n AstStatRepeat) isAstStat() {}

type AstStatReturn struct {
	NodeLoc
	List         []AstExpr `json:"list"`
	HasSemicolon *bool     `json:"hasSemicolon"`
}

func (n AstStatReturn) isAstNode() {}
func (n AstStatReturn) isAstStat() {}

type AstStatTypeAlias struct {
	NodeLoc
	Name         string               `json:"name"`
	NameLocation Location             `json:"nameLocation"`
	Generics     []AstGenericType     `json:"generics"`
	GenericPacks []AstGenericTypePack `json:"genericPacks"` // genericPacks always come after the generics
	Type         AstType              `json:"type"`
	Exported     bool                 `json:"exported"`
	HasSemicolon *bool                `json:"hasSemicolon"`
}

func (n AstStatTypeAlias) isAstNode() {}
func (n AstStatTypeAlias) isAstStat() {}

type AstStatTypeFunction struct {
	NodeLoc
	Name         string          `json:"name"`
	NameLocation Location        `json:"nameLocation"`
	Body         AstExprFunction `json:"body"`
	Exported     bool            `json:"exported"`
	HasErrors    bool            `json:"hasErrors"`
	HasSemicolon *bool           `json:"hasSemicolon"`
}

func (n AstStatTypeFunction) isAstNode() {}
func (n AstStatTypeFunction) isAstStat() {}

type AstStatWhile struct {
	NodeLoc
	Condition    AstExpr      `json:"condition"`
	Body         AstStatBlock `json:"body"`
	HasDo        bool         `json:"hasDo"`
	DoLocation   Location     `json:"doLocation"`
	HasSemicolon *bool        `json:"hasSemicolon"`
}

func (n AstStatWhile) isAstNode() {}
func (n AstStatWhile) isAstStat() {}

type AstTableIndexer struct {
	Location       Location  `json:"location"`
	IndexType      AstType   `json:"indexType"`
	ResultType     AstType   `json:"resultType"`
	Access         string    `json:"access"`
	AccessLocation *Location `json:"accessLocation"`
}

type AstTableProp struct {
	Name AstName `json:"name"`
	NodeLoc
	Type           AstType   `json:"type"`
	Access         string    `json:"access"`
	AccessLocation *Location `json:"accessLocation"`
}

type AstTypeError struct {
	NodeLoc
	Types        []AstType `json:"types"`
	IsMissing    bool      `json:"isMissing"`
	MessageIndex int       `json:"messageIndex"`
}

func (n AstTypeError) isAstNode() {}
func (n AstTypeError) isAstType() {}

type AstTypeFunction struct {
	NodeLoc
	Attributes   []AstAttr            `json:"attributes"`
	Generics     []AstGenericType     `json:"generics"`
	GenericPacks []AstGenericTypePack `json:"genericPacks"`
	ArgTypes     AstTypeList          `json:"argTypes"`
	ArgNames     []*AstArgumentName   `json:"argNames"`
	ReturnTypes  AstTypePackExplicit  `json:"returnTypes"`
}

func (n AstTypeFunction) isAstNode() {}
func (n AstTypeFunction) isAstType() {}

type AstTypeGroup struct {
	NodeLoc
	Type AstType `json:"type"`
}

func (n AstTypeGroup) isAstNode() {}
func (n AstTypeGroup) isAstType() {}

type AstTypeIntersection struct {
	NodeLoc
	Types []AstType `json:"types"`
}

func (n AstTypeIntersection) isAstNode() {}
func (n AstTypeIntersection) isAstType() {}

type AstTypeList struct {
	Types    []AstType    `json:"types"`
	TailType *AstTypePack `json:"tailType"`
}

func (n AstTypeList) isAstNode() {}
func (n AstTypeList) isAstType() {}

type AstTypeOptional struct {
	NodeLoc
}

func (n AstTypeOptional) isAstNode() {}
func (n AstTypeOptional) isAstType() {}

type AstTypeOrPack struct {
	Type *AstType     `json:"type"`
	Pack *AstTypePack `json:"pack"`
}

type AstTypePackExplicit struct {
	NodeLoc
	Types    AstTypeList  `json:"types"`
	TailType *AstTypePack `json:"tailType"`
}

func (n AstTypePackExplicit) isAstNode()     {}
func (n AstTypePackExplicit) isAstNodePack() {}
func (n AstTypePackExplicit) isAstTypePack() {}

type AstTypePackGeneric struct {
	NodeLoc
	GenericName string `json:"genericName"`
}

func (n AstTypePackGeneric) isAstNode()     {}
func (n AstTypePackGeneric) isAstNodePack() {}
func (n AstTypePackGeneric) isAstTypePack() {}

type AstTypePackVariadic struct {
	NodeLoc
	VariadicType AstType `json:"variadicType"`
}

func (n AstTypePackVariadic) isAstNode()     {}
func (n AstTypePackVariadic) isAstNodePack() {}
func (n AstTypePackVariadic) isAstTypePack() {}

type AstTypeReference struct {
	NodeLoc
	HasParameterList bool            `json:"hasParameterList"`
	Prefix           *string         `json:"prefix"`
	PrefixLocation   *Location       `json:"prefixLocation"`
	Name             string          `json:"name"`
	NameLocation     Location        `json:"nameLocation"`
	Parameters       []AstTypeOrPack `json:"parameters"`
}

func (n AstTypeReference) isAstNode() {}
func (n AstTypeReference) isAstType() {}

type AstTypeSingletonBool struct {
	NodeLoc
	Value bool `json:"value"`
}

func (n AstTypeSingletonBool) isAstNode() {}
func (n AstTypeSingletonBool) isAstType() {}

type AstTypeSingletonString struct {
	NodeLoc
	Value string `json:"value"`
}

func (n AstTypeSingletonString) isAstNode() {}
func (n AstTypeSingletonString) isAstType() {}

type AstTypeTable struct {
	NodeLoc
	Props   []AstTableProp   `json:"props"`
	Indexer *AstTableIndexer `json:"indexer"`
}

func (n AstTypeTable) isAstNode() {}
func (n AstTypeTable) isAstType() {}

// lol
type AstTypeTypeof struct {
	NodeLoc
	Expr AstExpr `json:"expr"`
}

func (n AstTypeTypeof) isAstNode() {}
func (n AstTypeTypeof) isAstType() {}

type AstTypeUnion struct {
	NodeLoc
	Types []AstType `json:"types"`
}

func (n AstTypeUnion) isAstNode() {}
func (n AstTypeUnion) isAstType() {}
