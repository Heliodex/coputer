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
	Type() string
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

func (a AstArgumentName) Type() string {
	return "AstArgumentName"
}

type AstAttr struct {
	NodeLoc
	Name string `json:"name"`
}

func (a AstAttr) Type() string {
	return "AstAttr"
}

type AstDeclaredClassProp struct {
	Name         string   `json:"name"`
	NameLocation Location `json:"nameLocation"`
	ASTNode
	LuauType T        `json:"luauType"`
	Location Location `json:"location"`
}

func (d AstDeclaredClassProp) GetLocation() Location {
	return d.Location
}

func (d AstDeclaredClassProp) Type() string {
	return "AstDeclaredClassProp"
}

type AstExprBinary struct {
	NodeLoc
	Op    string  `json:"op"`
	Left  AstExpr `json:"left"`
	Right AstExpr `json:"right"`
}

func (n AstExprBinary) Type() string {
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

func (n AstExprCall) Type() string {
	return "AstExprCall"
}

type AstExprConstantBool struct {
	NodeLoc
	Value bool `json:"value"`
}

func (n AstExprConstantBool) Type() string {
	return "AstExprConstantBool"
}

type AstExprConstantNil struct {
	NodeLoc
}

func (n AstExprConstantNil) Type() string {
	return "AstExprConstantNil"
}

type AstExprConstantNumber struct {
	NodeLoc
	Value float64 `json:"value"`
}

func (n AstExprConstantNumber) Type() string {
	return "AstExprConstantNumber"
}

type AstExprConstantString struct {
	NodeLoc
	Value string `json:"value"`
}

func (n AstExprConstantString) Type() string {
	return "AstExprConstantString"
}

type AstExprFunction struct {
	NodeLoc
	Attributes       []AstAttr            `json:"attributes"`
	Generics         []AstGenericType     `json:"generics"`
	GenericPacks     []AstGenericTypePack `json:"genericPacks"`
	Args             []T                  `json:"args"`
	ReturnAnnotation *AstTypePackExplicit `json:"returnAnnotation"`
	Vararg           bool                 `json:"vararg"`
	VarargLocation   Location             `json:"varargLocation"`
	Body             AstStatBlock         `json:"body"`
	FunctionDepth    int                  `json:"functionDepth"`
	Debugname        string               `json:"debugname"`
}

func (n AstExprFunction) Type() string {
	return "AstExprFunction"
}

type AstExprGlobal struct {
	NodeLoc
	Global string `json:"global"`
}

func (n AstExprGlobal) Type() string {
	return "AstExprGlobal"
}

type AstExprGroup struct {
	NodeLoc
	Expr T `json:"expr"` // only contains one expression right? strange when you first think about it
}

func (n AstExprGroup) Type() string {
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

func (n AstExprIfElse) Type() string {
	return "AstExprIfElse"
}

type AstExprIndexExpr struct {
	NodeLoc
	Expr  AstExpr `json:"expr"`
	Index AstExpr `json:"index"`
}

func (n AstExprIndexExpr) Type() string {
	return "AstExprIndexExpr"
}

type AstExprIndexName struct {
	NodeLoc
	Expr          T        `json:"expr"`
	Index         string   `json:"index"`
	IndexLocation Location `json:"indexLocation"`
	Op            string   `json:"op"`
}

func (n AstExprIndexName) Type() string {
	return "AstExprIndexName"
}

type AstExprInterpString struct {
	NodeLoc
	Strings     []string `json:"strings"`
	Expressions []T      `json:"expressions"`
}

func (n AstExprInterpString) Type() string {
	return "AstExprInterpString"
}

type AstExprLocal struct {
	NodeLoc
	Local T `json:"local"`
}

func (n AstExprLocal) Type() string {
	return "AstExprLocal"
}

type AstExprTable struct {
	NodeLoc
	Items []T `json:"items"`
}

func (n AstExprTable) Type() string {
	return "AstExprTable"
}

type AstExprTableItem struct {
	ASTNode
	Kind  string `json:"kind"`
	Key   *T     `json:"key"`
	Value T      `json:"value"`
}

func (n AstExprTableItem) GetLocation() Location {
	return Location{}
}

func (n AstExprTableItem) Type() string {
	return "AstExprTableItem"
}

type AstExprTypeAssertion struct {
	NodeLoc
	Expr       T `json:"expr"`
	Annotation T `json:"annotation"`
}

func (n AstExprTypeAssertion) Type() string {
	return "AstExprTypeAssertion"
}

type AstExprVarargs struct {
	NodeLoc
}

func (n AstExprVarargs) Type() string {
	return "AstExprVarargs"
}

type AstExprUnary struct {
	NodeLoc
	Op   string `json:"op"`
	Expr T      `json:"expr"`
}

func (n AstExprUnary) Type() string {
	return "AstExprUnary"
}

type AstGenericType struct {
	ASTNode
	Name string `json:"name"`
}

func (g AstGenericType) GetLocation() Location {
	return Location{}
}

func (g AstGenericType) Type() string {
	return "AstGenericType"
}

type AstGenericTypePack struct {
	ASTNode
	Name string `json:"name"`
}

func (g AstGenericTypePack) GetLocation() Location {
	return Location{}
}

func (g AstGenericTypePack) Type() string {
	return "AstGenericTypePack"
}

type AstLocal struct {
	LuauType *T     `json:"luauType"` // for now it's probably nil?
	Name     string `json:"name"`
	NodeLoc
}

func (n AstLocal) Type() string {
	return "AstLocal"
}

type AstStatAssign struct {
	NodeLoc
	Vars   []T `json:"vars"`
	Values []T `json:"values"`
}

func (n AstStatAssign) Type() string {
	return "AstStatAssign"
}

type AstStatBlock struct {
	NodeLoc
	HasEnd            bool       `json:"hasEnd"`
	Body              []T        `json:"body"`
	CommentsContained *[]Comment // not in json
}

func (n AstStatBlock) Type() string {
	return "AstStatBlock"
}

type AstStatBreak struct {
	NodeLoc
}

func (n AstStatBreak) Type() string {
	return "AstStatBreak"
}

type AstStatCompoundAssign struct {
	NodeLoc
	Op    string `json:"op"`
	Var   T      `json:"var"`
	Value T      `json:"value"`
}

func (n AstStatCompoundAssign) Type() string {
	return "AstStatCompoundAssign"
}

type AstStatContinue struct {
	NodeLoc
}

func (n AstStatContinue) Type() string {
	return "AstStatContinue"
}

type AstStatDeclareClass struct {
	NodeLoc
	Name      string  `json:"name"`
	SuperName *string `json:"superName"`
	Props     []T     `json:"props"`
	Indexer   *T      `json:"indexer"`
}

func (n AstStatDeclareClass) Type() string {
	return "AstStatDeclareClass"
}

type AstStatExpr struct {
	NodeLoc
	Expr T `json:"expr"`
}

func (n AstStatExpr) Type() string {
	return "AstStatExpr"
}

type AstStatFor struct {
	NodeLoc
	Var   T            `json:"var"`
	From  T            `json:"from"`
	To    T            `json:"to"`
	Step  *T           `json:"step"`
	Body  AstStatBlock `json:"body"`
	HasDo bool         `json:"hasDo"`
}

func (n AstStatFor) Type() string {
	return "AstStatFor"
}

type AstStatForIn struct {
	NodeLoc
	Vars   []AstLocal   `json:"vars"`
	Values []T          `json:"values"`
	Body   AstStatBlock `json:"body"`
	HasIn  bool         `json:"hasIn"`
	HasDo  bool         `json:"hasDo"`
}

func (n AstStatForIn) Type() string {
	return "AstStatForIn"
}

type AstStatFunction struct {
	NodeLoc
	Name         AstExpr         `json:"name"`
	Func         AstExprFunction `json:"func"`
	HasSemicolon *bool           `json:"hasSemicolon"`
}

func (n AstStatFunction) Type() string {
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

func (n AstStatIf) Type() string {
	return "AstStatIf"
}

type AstStatLocal struct {
	NodeLoc
	Vars   []AstLocal `json:"vars"`
	Values []AstExpr        `json:"values"`
	EqualsSignLocation *Location `json:"equalsSignLocation"`
	HasSemicolon *bool      `json:"hasSemicolon"`
}

func (n AstStatLocal) Type() string {
	return "AstStatLocal"
}

type AstStatLocalFunction struct {
	NodeLoc
	Name         AstLocal        `json:"name"`
	Func         AstExprFunction `json:"func"`
	HasSemicolon *bool           `json:"hasSemicolon"`
}

func (n AstStatLocalFunction) Type() string {
	return "AstStatLocalFunction"
}

type AstStatRepeat struct {
	NodeLoc
	Condition T            `json:"condition"`
	Body      AstStatBlock `json:"body"`
}

func (n AstStatRepeat) Type() string {
	return "AstStatRepeat"
}

type AstStatReturn struct {
	NodeLoc
	List []T `json:"list"`
}

func (n AstStatReturn) Type() string {
	return "AstStatReturn"
}

type AstStatTypeAlias struct {
	NodeLoc
	Name         string               `json:"name"`
	Generics     []AstGenericType     `json:"generics"`
	GenericPacks []AstGenericTypePack `json:"genericPacks"` // genericPacks always come after the generics
	Value        T                    `json:"value"`
	Exported     bool                 `json:"exported"`
}

func (n AstStatTypeAlias) Type() string {
	return "AstStatTypeAlias"
}

type AstStatWhile struct {
	NodeLoc
	Condition T            `json:"condition"`
	Body      AstStatBlock `json:"body"`
	HasDo     bool         `json:"hasDo"`
}

func (n AstStatWhile) Type() string {
	return "AstStatWhile"
}

type AstTableProp struct {
	Name string `json:"name"`
	NodeLoc
	PropType T `json:"propType"`
}

func (n AstTableProp) Type() string {
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

func (n AstTypeFunction) Type() string {
	return "AstTypeFunction"
}

type AstTypeGroup struct {
	NodeLoc
	Inner T `json:"inner"`
}

func (n AstTypeGroup) Type() string {
	return "AstTypeGroup"
}

type AstTypeList struct {
	ASTNode
	Types    []T `json:"types"`
	TailType *T  `json:"tailType"`
}

func (n AstTypeList) GetLocation() Location {
	return Location{}
}

func (n AstTypeList) Type() string {
	return "AstTypeList"
}

type AstTypeOptional struct {
	NodeLoc
}

func (AstTypeOptional) Type() string {
	return "AstTypeOptional"
}

type AstTypePackExplicit struct {
	NodeLoc
	TypeList AstTypeList `json:"typeList"`
}

func (n AstTypePackExplicit) Type() string {
	return "AstTypePackExplicit"
}

type AstTypePackGeneric struct {
	NodeLoc
	GenericName string `json:"genericName"`
}

func (n AstTypePackGeneric) Type() string {
	return "AstTypePackGeneric"
}

type AstTypePackVariadic struct {
	NodeLoc
	VariadicType AstType `json:"variadicType"`
}

func (n AstTypePackVariadic) Type() string {
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

func (n AstTypeReference) Type() string {
	return "AstTypeReference"
}

type AstTypeSingletonBool struct {
	NodeLoc
	Value bool `json:"value"`
}

func (n AstTypeSingletonBool) Type() string {
	return "AstTypeSingletonBool"
}

type AstTypeSingletonString struct {
	NodeLoc
	Value string `json:"value"`
}

func (n AstTypeSingletonString) Type() string {
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

func (n AstTypeTable) Type() string {
	return "AstTypeTable"
}

// lol
type AstTypeTypeof struct {
	NodeLoc
	Expr AstExpr `json:"expr"`
}

func (n AstTypeTypeof) Type() string {
	return "AstTypeTypeof"
}

type AstTypeUnion struct {
	NodeLoc
	Types []AstType `json:"types"`
}

func (n AstTypeUnion) Type() string {
	return "AstTypeUnion"
}
