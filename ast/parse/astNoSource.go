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
	GetLocation() Location
	String() string
	Kind() string
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
type ASTNode struct {
	Type string `json:"type"`
}

type NodeLoc struct {
	ASTNode
	Location Location `json:"location"`
}

func (nl NodeLoc) GetLocation() Location {
	return nl.Location
}

// ast

type Comment struct {
	NodeLoc
}

// statblocks are the most important node for comments
type StatBlockDepth struct {
	AstStatBlock
	Depth int
}

type AST struct {
	Root             AstStatBlock `json:"root"`
	CommentLocations []Comment    `json:"commentLocations"`
}

type AddStatBlock func(AstStatBlock, int)

// node type groops

// node types (ok, real ast now)

type AstArgumentName struct {
	ASTNode
	Name     string   `json:"name"`
	Location Location `json:"location"`
}

func (a AstArgumentName) GetLocation() Location {
	return a.Location
}

func (a AstArgumentName) Kind() string {
	return "AstArgumentName"
}

type AstAttr struct {
	NodeLoc
	Name string `json:"name"`
}

func (a AstAttr) Kind() string {
	return "AstAttr"
}

type AstExprBinary struct {
	NodeLoc
	Op    string  `json:"op"`
	Left  AstExpr `json:"left"`
	Right AstExpr `json:"right"`
}

func (n AstExprBinary) Kind() string {
	return "AstExprBinary"
}

type AstExprCall struct {
	NodeLoc
	Func          AstExpr          `json:"func"`
	Args          []AstExpr        `json:"args"`
	Self          bool             `json:"self"`
	ArgLocation   Location         `json:"argLocation"`
	TypeArguments *[]AstTypeOrPack `json:"typeArguments"`
}

func (n AstExprCall) Kind() string {
	return "AstExprCall"
}

type AstExprConstantBool struct {
	NodeLoc
	Value bool `json:"value"`
}

func (n AstExprConstantBool) Kind() string {
	return "AstExprConstantBool"
}

type AstExprConstantNil struct {
	NodeLoc
}

func (n AstExprConstantNil) Kind() string {
	return "AstExprConstantNil"
}

type AstExprConstantNumber struct {
	NodeLoc
	Value float64 `json:"value"`
}

func (n AstExprConstantNumber) Kind() string {
	return "AstExprConstantNumber"
}

type AstExprConstantString struct {
	NodeLoc
	Value string `json:"value"`
}

func (n AstExprConstantString) Kind() string {
	return "AstExprConstantString"
}

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

func (n AstExprFunction) Kind() string {
	return "AstExprFunction"
}

type AstExprGlobal struct {
	NodeLoc
	Global string `json:"global"`
}

func (n AstExprGlobal) Kind() string {
	return "AstExprGlobal"
}

type AstExprGroup struct {
	NodeLoc
	Expr AstExpr `json:"expr"` // only contains one expression right? strange when you first think about it
}

func (n AstExprGroup) Kind() string {
	return "AstExprGroup"
}

type AstExprIfElse struct {
	NodeLoc
	Condition AstExpr `json:"condition"`
	HasThen   bool    `json:"hasThen"`
	TrueExpr  AstExpr `json:"trueExpr"`
	HasElse   bool    `json:"hasElse"`
	FalseExpr AstExpr `json:"falseExpr"`
}

func (n AstExprIfElse) Kind() string {
	return "AstExprIfElse"
}

type AstExprIndexExpr struct {
	NodeLoc
	Expr  AstExpr `json:"expr"`
	Index AstExpr `json:"index"`
}

func (n AstExprIndexExpr) Kind() string {
	return "AstExprIndexExpr"
}

type AstExprIndexName struct {
	NodeLoc
	Expr          AstExpr  `json:"expr"`
	Index         string   `json:"index"`
	IndexLocation Location `json:"indexLocation"`
	OpPosition    Position `json:"opPosition"`
	Op            rune     `json:"op"`
}

func (n AstExprIndexName) Kind() string {
	return "AstExprIndexName"
}

type AstExprInterpString struct {
	NodeLoc
	Strings     []string  `json:"strings"`
	Expressions []AstExpr `json:"expressions"`
}

func (n AstExprInterpString) Kind() string {
	return "AstExprInterpString"
}

type AstExprLocal struct {
	NodeLoc
	Local   AstLocal `json:"local"`
	Upvalue bool     `json:"upvalue"`
}

func (n AstExprLocal) Kind() string {
	return "AstExprLocal"
}

type AstExprTable struct {
	NodeLoc
	Items []AstExprTableItem `json:"items"`
}

func (n AstExprTable) Kind() string {
	return "AstExprTable"
}

type AstExprTableItem struct {
	ASTNode
	kind  string   `json:"kind"` // "List" | "Record" | "General"
	Key   *AstExpr `json:"key"`
	Value AstExpr  `json:"value"`
}

func (n AstExprTableItem) GetLocation() Location {
	return Location{}
}

func (n AstExprTableItem) Kind() string {
	return n.kind // Weerd.m4a
}

type AstExprTypeAssertion struct {
	NodeLoc
	Expr       AstExpr `json:"expr"`
	Annotation AstType `json:"annotation"`
}

func (n AstExprTypeAssertion) Kind() string {
	return "AstExprTypeAssertion"
}

type AstExprVarargs struct {
	NodeLoc
}

func (n AstExprVarargs) Kind() string {
	return "AstExprVarargs"
}

type AstExprUnary struct {
	NodeLoc
	Op   string  `json:"op"`
	Expr AstExpr `json:"expr"`
}

func (n AstExprUnary) Kind() string {
	return "AstExprUnary"
}

type AstGenericType struct {
	ASTNode
	Name string `json:"name"`
}

func (g AstGenericType) GetLocation() Location {
	return Location{}
}

func (g AstGenericType) Kind() string {
	return "AstGenericType"
}

type AstGenericTypePack struct {
	ASTNode
	Name string `json:"name"`
}

func (g AstGenericTypePack) GetLocation() Location {
	return Location{}
}

func (g AstGenericTypePack) Kind() string {
	return "AstGenericTypePack"
}

type AstLocal struct {
	Name string `json:"name"`
	NodeLoc
	Shadow        *AstLocal `json:"shadow"`
	FunctionDepth int       `json:"functionDepth"`
	LoopDepth     int       `json:"loopDepth"`
	Annotation    *AstType  `json:"annotation"`
}

func (n AstLocal) Kind() string {
	return "AstLocal"
}

type AstStatAssign struct {
	NodeLoc
	Vars         []AstExpr `json:"vars"`
	Values       []AstExpr `json:"values"`
	HasSemicolon *bool     `json:"hasSemicolon"`
}

func (n AstStatAssign) Kind() string {
	return "AstStatAssign"
}

type AstStatBlock struct {
	NodeLoc
	HasEnd            bool       `json:"hasEnd"`
	Body              []AstStat  `json:"body"`
	CommentsContained *[]Comment // not in json
}

func (n AstStatBlock) Kind() string {
	return "AstStatBlock"
}

type AstStatBreak struct {
	NodeLoc
}

func (n AstStatBreak) Kind() string {
	return "AstStatBreak"
}

type AstStatCompoundAssign struct {
	NodeLoc
	Op           int     `json:"op"`
	Var          AstExpr `json:"var"`
	Value        AstExpr `json:"value"`
	HasSemicolon *bool   `json:"hasSemicolon"`
}

func (n AstStatCompoundAssign) Kind() string {
	return "AstStatCompoundAssign"
}

type AstStatContinue struct {
	NodeLoc
}

func (n AstStatContinue) Kind() string {
	return "AstStatContinue"
}

type AstStatExpr struct {
	NodeLoc
	Expr AstExpr `json:"expr"`
}

func (n AstStatExpr) Kind() string {
	return "AstStatExpr"
}

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

func (n AstStatFor) Kind() string {
	return "AstStatFor"
}

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

func (n AstStatForIn) Kind() string {
	return "AstStatForIn"
}

type AstStatFunction struct {
	NodeLoc
	Name         AstExpr         `json:"name"`
	Func         AstExprFunction `json:"func"`
	HasSemicolon *bool           `json:"hasSemicolon"`
}

func (n AstStatFunction) Kind() string {
	return "AstStatFunction"
}

type AstStatIf struct {
	NodeLoc
	Condition    AstExpr      `json:"condition"`
	ThenBody     AstStatBlock `json:"thenbody"`
	ElseBody     *AstStat     `json:"elsebody"` // StatBlock | StatIf
	ThenLocation *Location    `json:"thenLocation"`
	ElseLocation *Location    `json:"elseLocation"`
	HasSemicolon *bool        `json:"hasSemicolon"`
}

func (n AstStatIf) Kind() string {
	return "AstStatIf"
}

type AstStatLocal struct {
	NodeLoc
	Vars               []AstLocal `json:"vars"`
	Values             []AstExpr  `json:"values"`
	EqualsSignLocation *Location  `json:"equalsSignLocation"`
	HasSemicolon       *bool      `json:"hasSemicolon"`
}

func (n AstStatLocal) Kind() string {
	return "AstStatLocal"
}

type AstStatLocalFunction struct {
	NodeLoc
	Name         AstLocal        `json:"name"`
	Func         AstExprFunction `json:"func"`
	HasSemicolon *bool           `json:"hasSemicolon"`
}

func (n AstStatLocalFunction) Kind() string {
	return "AstStatLocalFunction"
}

type AstStatRepeat struct {
	NodeLoc
	Condition    AstExpr      `json:"condition"`
	Body         AstStatBlock `json:"body"`
	HasUntil     bool         `json:"hasUntil"`
	HasSemicolon *bool        `json:"hasSemicolon"`
}

func (n AstStatRepeat) Kind() string {
	return "AstStatRepeat"
}

type AstStatReturn struct {
	NodeLoc
	List         []AstExpr `json:"list"`
	HasSemicolon *bool     `json:"hasSemicolon"`
}

func (n AstStatReturn) Kind() string {
	return "AstStatReturn"
}

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

func (n AstStatTypeAlias) Kind() string {
	return "AstStatTypeAlias"
}

type AstStatWhile struct {
	NodeLoc
	Condition    AstExpr      `json:"condition"`
	Body         AstStatBlock `json:"body"`
	HasDo        bool         `json:"hasDo"`
	DoLocation   Location     `json:"doLocation"`
	HasSemicolon *bool        `json:"hasSemicolon"`
}

func (n AstStatWhile) Kind() string {
	return "AstStatWhile"
}

type AstTableProp struct {
	Name AstName `json:"name"`
	NodeLoc
	Type           AstType   `json:"type"`
	Access         string    `json:"access"`
	AccessLocation *Location `json:"accessLocation"`
}

func (n AstTableProp) Kind() string {
	return "AstTableProp"
}

type AstTypeFunction struct {
	NodeLoc
	Attributes   []AstAttr            `json:"attributes"`
	Generics     []AstGenericType     `json:"generics"`
	GenericPacks []AstGenericTypePack `json:"genericPacks"`
	ArgTypes     AstTypeList          `json:"argTypes"`
	ArgNames     []*AstArgumentName   `json:"argNames"`
	ReturnTypes  AstTypePackExplicit  `json:"returnTypes"`
}

func (n AstTypeFunction) Kind() string {
	return "AstTypeFunction"
}

type AstTypeGroup struct {
	NodeLoc
	Type AstType `json:"type"`
}

func (n AstTypeGroup) Kind() string {
	return "AstTypeGroup"
}

type AstTypeList struct {
	ASTNode
	Types    []AstType    `json:"types"`
	TailType *AstTypePack `json:"tailType"`
}

func (n AstTypeList) GetLocation() Location {
	return Location{}
}

func (n AstTypeList) Kind() string {
	return "AstTypeList"
}

type AstTypeOptional struct {
	NodeLoc
}

func (AstTypeOptional) Kind() string {
	return "AstTypeOptional"
}

type AstTypePackExplicit struct {
	NodeLoc
	TypeList AstTypeList `json:"typeList"`
}

func (n AstTypePackExplicit) Kind() string {
	return "AstTypePackExplicit"
}

type AstTypePackGeneric struct {
	NodeLoc
	GenericName string `json:"genericName"`
}

func (n AstTypePackGeneric) Kind() string {
	return "AstTypePackGeneric"
}

type AstTypePackVariadic struct {
	NodeLoc
	VariadicType AstType `json:"variadicType"`
}

func (n AstTypePackVariadic) Kind() string {
	return "AstTypePackVariadic"
}

type AstTypeReference struct {
	NodeLoc
	HasParameterList bool            `json:"hasParameterList"`
	Prefix           *string         `json:"prefix"`
	PrefixLocation   *Location       `json:"prefixLocation"`
	Name             string          `json:"name"`
	NameLocation     Location        `json:"nameLocation"`
	Parameters       []AstTypeOrPack `json:"parameters"`
}

func (n AstTypeReference) Kind() string {
	return "AstTypeReference"
}

type AstTypeSingletonBool struct {
	NodeLoc
	Value bool `json:"value"`
}

func (n AstTypeSingletonBool) Kind() string {
	return "AstTypeSingletonBool"
}

type AstTypeSingletonString struct {
	NodeLoc
	Value string `json:"value"`
}

func (n AstTypeSingletonString) Kind() string {
	return "AstTypeSingletonString"
}

type AstTableIndexer struct {
	Location       Location  `json:"location"`
	IndexType      AstType   `json:"indexType"`
	ResultType     AstType   `json:"resultType"`
	Access         string    `json:"access"`
	AccessLocation *Location `json:"accessLocation"`
}

type AstTypeTable struct {
	NodeLoc
	Props   []AstTableProp   `json:"props"`
	Indexer *AstTableIndexer `json:"indexer"`
}

func (n AstTypeTable) Kind() string {
	return "AstTypeTable"
}

// lol
type AstTypeTypeof struct {
	NodeLoc
	Expr AstExpr `json:"expr"`
}

func (n AstTypeTypeof) Kind() string {
	return "AstTypeTypeof"
}

type AstTypeUnion struct {
	NodeLoc
	Types []AstType `json:"types"`
}

func (n AstTypeUnion) Kind() string {
	return "AstTypeUnion"
}
