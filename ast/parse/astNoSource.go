package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
)

const (
	Ext            = ".luau"
	AstDir         = "../test/ast"
	BenchmarkDir   = "../test/benchmark"
	ConformanceDir = "../test/conformance"
)

func transformAst(in []byte) []byte {
	strin := string(in)
	out := strings.ReplaceAll(strin, `"value":Infinity`, `"value":"Infinity"`)
	return []byte(out)
}

func LuauAst(path string) (out []byte, err error) {
	cmd := exec.Command("luau-ast", path)
	out, err = cmd.Output()
	if err != nil {
		return
	}
	return transformAst(out), nil
}

func LuauAstInput(source []byte) (out []byte, err error) {
	tempfile, err := os.CreateTemp("", "luau-ast-*.luau")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tempfile.Name())

	if _, err = tempfile.Write(source); err != nil {
		return nil, fmt.Errorf("write to temp file: %w", err)
	}
	if err = tempfile.Close(); err != nil {
		return nil, fmt.Errorf("close temp file: %w", err)
	}
	return LuauAst(tempfile.Name())
}

func indentStart(s string, n int) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i, line := range lines {
		lines[i] = strings.Repeat(" ", n) + line
	}
	return strings.Join(lines, "\n")
}

type Number float64

func (n Number) MarshalJSON() ([]byte, error) {
	if float64(n) == math.Inf(1) {
		return json.Marshal("Infinity")
	}
	return json.Marshal(float64(n))
}

func (n *Number) UnmarshalJSON(data []byte) (err error) {
	// check if it's "Infinity"
	if string(data) == `"Infinity"` {
		*n = Number(math.Inf(1))
		return
	}

	var f float64
	if err = json.Unmarshal(data, &f); err != nil {
		return
	}
	*n = Number(f)
	return
}

type String string

func (s String) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// oh god
// trying to get the JSON strings from luau-ast's insane format
func (s *String) UnmarshalJSON(data []byte) (err error) {
	// dataNoQuotes := strings.Trim(string(data), `"`)
	dataNoQuotes := string(data[1 : len(data)-1])

	var data2 []byte
	lnq := len(dataNoQuotes)
	for i := 0; i < lnq; i++ {
		lenLeft := lnq - i

		next10 := dataNoQuotes[i:][:min(10, lenLeft)]
		if lenLeft < 10 || next10[:8] != `\uffffff` { // fix luau-ast invalid unicode escapes
			data2 = append(data2, dataNoQuotes[i])
			continue
		}

		char := next10[8:]
		decoded, err := hex.DecodeString(char)
		if err != nil {
			return err
		}

		data2 = append(data2, decoded[0])
		i += 9
	}

	// fmt.Println("UnmarshalJSON", string(data2))

	// another pass to check and replace \u escapes
	d2s := string(data2)
	d2s = strings.ReplaceAll(d2s, `\"`, `"`)
	d2s = strings.ReplaceAll(d2s, `\\`, `\`)

	var data3 []byte
	ld2 := len(d2s)
	for i := 0; i < ld2; i++ {
		lenLeft := ld2 - i

		next6 := d2s[i:][:min(6, lenLeft)]
		if lenLeft < 6 || next6[:2] != `\u` {
			data3 = append(data3, d2s[i])
			continue
		}

		char := next6[2:]
		decoded, err := hex.DecodeString(char)
		if err != nil {
			return err
		}

		if decoded[0] == 0 {
			decoded = decoded[1:]
		}

		// fmt.Println("decoded", decoded)
		data3 = append(data3, decoded...)
		i += 5
	}

	*s = String(string(data3))

	return
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

func (n ASTNode) String() string {
	return fmt.Sprintf("Type: %s\n", n.Type)
}

type NodeLoc struct {
	ASTNode
	Location Location `json:"location"`
}

func (nl NodeLoc) GetLocation() Location {
	return nl.Location
}

func StringMaybeEvaluated(val any) string {
	if v, ok := val.(json.RawMessage); ok {
		var node ASTNode
		if err := json.Unmarshal(v, &node); err != nil {
			return fmt.Sprintf("decode Node: %v", err)
		}
		return node.String()
	}
	return fmt.Sprintf("%v", val)
}

// ast

type Comment struct {
	NodeLoc
}

func (c Comment) String() string {
	var b strings.Builder

	b.WriteString(c.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s\n", c.Location))

	return b.String()
}

// statblocks are the most important node for comments
type StatBlockDepth struct {
	AstStatBlock[Node]
	Depth int
}

type AST[T any] struct {
	Root             AstStatBlock[T] `json:"root"`
	CommentLocations []Comment       `json:"commentLocations"`
}

func (ast AST[T]) String() string {
	var b strings.Builder

	b.WriteString("Root:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(ast.Root), 4))
	b.WriteString("\n\n")

	b.WriteString("Comment Locations:")
	for _, c := range ast.CommentLocations {
		b.WriteByte('\n')
		b.WriteString(indentStart(c.String(), 4))
	}
	b.WriteByte('\n')

	return b.String()
}

type AddStatBlock func(AstStatBlock[Node], int)

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

func (a AstArgumentName) String() string {
	var b strings.Builder

	b.WriteString(a.ASTNode.String())
	b.WriteString(fmt.Sprintf("Name: %s\n", a.Name))
	b.WriteString(fmt.Sprintf("Location: %s\n", a.Location))

	return b.String()
}

type AstAttr struct {
	NodeLoc
	Name string `json:"name"`
}

func (a AstAttr) Type() string {
	return "AstAttr"
}

func (a AstAttr) String() string {
	var b strings.Builder

	b.WriteString(a.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s\n", a.Location))
	b.WriteString(fmt.Sprintf("Name: %s\n", a.Name))

	return b.String()
}

type AstDeclaredClassProp[T any] struct {
	Name         string   `json:"name"`
	NameLocation Location `json:"nameLocation"`
	ASTNode
	LuauType T        `json:"luauType"`
	Location Location `json:"location"`
}

func (d AstDeclaredClassProp[T]) GetLocation() Location {
	return d.Location
}

func (d AstDeclaredClassProp[T]) Type() string {
	return "AstDeclaredClassProp"
}

func (d AstDeclaredClassProp[T]) String() string {
	var b strings.Builder

	b.WriteString(d.ASTNode.String())
	b.WriteString(fmt.Sprintf("Name: %s", d.Name))
	b.WriteString(fmt.Sprintf("\nNameLocation: %s", d.NameLocation))
	b.WriteString(fmt.Sprintf("\nLocation: %s", d.Location))
	b.WriteString("\nLuauType:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(d.LuauType), 4))

	return b.String()
}

type AstExprBinary[T any] struct {
	NodeLoc
	Op    string `json:"op"`
	Left  T      `json:"left"`
	Right T      `json:"right"`
}

func (n AstExprBinary[T]) Type() string {
	return "AstExprBinary"
}

func (n AstExprBinary[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nOp: %s", n.Op))
	b.WriteString("\nLeft:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Left), 4))
	b.WriteString("\nRight:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Right), 4))

	return b.String()
}

type AstExprCall[T any] struct {
	NodeLoc
	Func        T        `json:"func"`
	Args        []T      `json:"args"`
	Self        bool     `json:"self"`
	ArgLocation Location `json:"argLocation"`
}

func (n AstExprCall[T]) Type() string {
	return "AstExprCall"
}

func (n AstExprCall[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nFunc:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Func), 4))
	b.WriteString("\nArgs:")

	for _, arg := range n.Args {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(arg), 4))
	}

	b.WriteString(fmt.Sprintf("\nSelf: %t", n.Self))
	b.WriteString(fmt.Sprintf("\nArgLocation: %s", n.ArgLocation))

	return b.String()
}

type AstExprConstantBool struct {
	NodeLoc
	Value bool `json:"value"`
}

func (n AstExprConstantBool) Type() string {
	return "AstExprConstantBool"
}

func (n AstExprConstantBool) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nValue: %t", n.Value))

	return b.String()
}

type AstExprConstantNil struct {
	NodeLoc
}

func (n AstExprConstantNil) Type() string {
	return "AstExprConstantNil"
}

func (n AstExprConstantNil) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))

	return b.String()
}

type AstExprConstantNumber struct {
	NodeLoc
	Value Number `json:"value"`
}

func (n AstExprConstantNumber) Type() string {
	return "AstExprConstantNumber"
}

func (n AstExprConstantNumber) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nValue: %f", n.Value))

	return b.String()
}

type AstExprConstantString struct {
	NodeLoc
	Value String `json:"value"`
}

func (n AstExprConstantString) Type() string {
	return "AstExprConstantString"
}

func (n AstExprConstantString) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nValue: %s", n.Value))

	return b.String()
}

type AstExprFunction[T any] struct {
	NodeLoc
	Attributes       []AstAttr               `json:"attributes"`
	Generics         []GenericType           `json:"generics"`
	GenericPacks     []GenericTypePack       `json:"genericPacks"`
	Args             []T                     `json:"args"`
	ReturnAnnotation *AstTypePackExplicit[T] `json:"returnAnnotation"`
	Vararg           bool                    `json:"vararg"`
	VarargLocation   Location                `json:"varargLocation"`
	Body             AstStatBlock[T]         `json:"body"`
	FunctionDepth    int                     `json:"functionDepth"`
	Debugname        string                  `json:"debugname"`
}

func (n AstExprFunction[T]) Type() string {
	return "AstExprFunction"
}

func (n AstExprFunction[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nAttributes:")
	for _, attr := range n.Attributes {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(attr), 4))
	}
	b.WriteString("\nGenerics:")
	for _, gen := range n.Generics {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(gen), 4))
	}
	b.WriteString("\nGenericPacks:")
	for _, pack := range n.GenericPacks {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(pack), 4))
	}
	b.WriteString("\nArgs:")
	for _, arg := range n.Args {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(arg), 4))
	}
	b.WriteString("\nReturnAnnotation:")
	if n.ReturnAnnotation != nil {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(*n.ReturnAnnotation), 4))
	}

	b.WriteString(fmt.Sprintf("\nVararg: %t", n.Vararg))
	b.WriteString(fmt.Sprintf("\nVarargLocation: %s", n.VarargLocation))
	b.WriteString("\nBody:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Body), 4))
	b.WriteString(fmt.Sprintf("\nFunctionDepth: %d", n.FunctionDepth))
	b.WriteString(fmt.Sprintf("\nDebugname: %s", n.Debugname))

	return b.String()
}

type AstExprGlobal struct {
	NodeLoc
	Global string `json:"global"`
}

func (n AstExprGlobal) Type() string {
	return "AstExprGlobal"
}

func (n AstExprGlobal) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nGlobal: %s", n.Global))

	return b.String()
}

type AstExprGroup[T any] struct {
	NodeLoc
	Expr T `json:"expr"` // only contains one expression right? strange when you first think about it
}

func (n AstExprGroup[T]) Type() string {
	return "AstExprGroup"
}

func (n AstExprGroup[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))

	return b.String()
}

type AstExprIfElse[T any] struct {
	NodeLoc
	Condition T    `json:"condition"`
	HasThen   bool `json:"hasThen"`
	TrueExpr  T    `json:"trueExpr"`
	HasElse   bool `json:"hasElse"`
	FalseExpr T    `json:"falseExpr"`
}

func (n AstExprIfElse[T]) Type() string {
	return "AstExprIfElse"
}

func (n AstExprIfElse[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nCondition:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Condition), 4))
	b.WriteString(fmt.Sprintf("\nHasThen: %t", n.HasThen))
	b.WriteString("\nTrueExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.TrueExpr), 4))
	b.WriteString(fmt.Sprintf("\nHasElse: %t", n.HasElse))
	b.WriteString("\nFalseExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.FalseExpr), 4))

	return b.String()
}

type AstExprIndexExpr[T any] struct {
	NodeLoc
	Expr  T `json:"expr"`
	Index T `json:"index"`
}

func (n AstExprIndexExpr[T]) Type() string {
	return "AstExprIndexExpr"
}

func (n AstExprIndexExpr[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))
	b.WriteString("\nIndex:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Index), 4))

	return b.String()
}

type AstExprIndexName[T any] struct {
	NodeLoc
	Expr          T        `json:"expr"`
	Index         string   `json:"index"`
	IndexLocation Location `json:"indexLocation"`
	Op            string   `json:"op"`
}

func (n AstExprIndexName[T]) Type() string {
	return "AstExprIndexName"
}

func (n AstExprIndexName[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))
	b.WriteString(fmt.Sprintf("\nIndex: %s", n.Index))

	return b.String()
}

type AstExprInterpString[T any] struct {
	NodeLoc
	Strings     []string `json:"strings"`
	Expressions []T      `json:"expressions"`
}

func (n AstExprInterpString[T]) Type() string {
	return "AstExprInterpString"
}

func (n AstExprInterpString[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nStrings:")
	for _, str := range n.Strings {
		b.WriteByte('\n')
		b.WriteString(indentStart(str, 4))
	}
	b.WriteString("\nExpressions:")
	for _, expr := range n.Expressions {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(expr), 4))
	}

	return b.String()
}

type AstExprLocal[T any] struct {
	NodeLoc
	Local T `json:"local"`
}

func (n AstExprLocal[T]) Type() string {
	return "AstExprLocal"
}

func (n AstExprLocal[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nLocal:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Local), 4))

	return b.String()
}

type AstExprTable[T any] struct {
	NodeLoc
	Items []T `json:"items"`
}

func (n AstExprTable[T]) Type() string {
	return "AstExprTable"
}

func (n AstExprTable[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nItems:")

	for _, item := range n.Items {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(item), 4))
	}

	return b.String()
}

type AstExprTableItem[T any] struct {
	ASTNode
	Kind  string `json:"kind"`
	Key   *T     `json:"key"`
	Value T      `json:"value"`
}

func (n AstExprTableItem[T]) GetLocation() Location {
	return Location{}
}

func (n AstExprTableItem[T]) Type() string {
	return "AstExprTableItem"
}

func (n AstExprTableItem[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Kind: %s", n.Kind))
	b.WriteString("\nKey:")
	if n.Key != nil {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(*n.Key), 4))
	}
	b.WriteString("\nValue:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Value), 4))

	return b.String()
}

type AstExprTypeAssertion[T any] struct {
	NodeLoc
	Expr       T `json:"expr"`
	Annotation T `json:"annotation"`
}

func (n AstExprTypeAssertion[T]) Type() string {
	return "AstExprTypeAssertion"
}

func (n AstExprTypeAssertion[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))
	b.WriteString("\nAnnotation:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Annotation), 4))

	return b.String()
}

type AstExprVarargs struct {
	NodeLoc
}

func (n AstExprVarargs) Type() string {
	return "AstExprVarargs"
}

func (n AstExprVarargs) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))

	return b.String()
}

type AstExprUnary[T any] struct {
	NodeLoc
	Op   string `json:"op"`
	Expr T      `json:"expr"`
}

func (n AstExprUnary[T]) Type() string {
	return "AstExprUnary"
}

func (n AstExprUnary[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nOp: %s", n.Op))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))

	return b.String()
}

type GenericType struct {
	ASTNode
	Name string `json:"name"`
}

func (g GenericType) GetLocation() Location {
	return Location{}
}

func (g GenericType) Type() string {
	return "AstGenericType"
}

func (g GenericType) String() string {
	var b strings.Builder

	b.WriteString(g.ASTNode.String())
	b.WriteString(fmt.Sprintf("Name: %s", g.Name))

	return b.String()
}

type GenericTypePack struct {
	ASTNode
	Name string `json:"name"`
}

func (g GenericTypePack) GetLocation() Location {
	return Location{}
}

func (g GenericTypePack) Type() string {
	return "AstGenericTypePack"
}

func (g GenericTypePack) String() string {
	var b strings.Builder

	b.WriteString(g.ASTNode.String())
	b.WriteString(fmt.Sprintf("Name: %s", g.Name))

	return b.String()
}

type AstLocal[T any] struct {
	LuauType *T     `json:"luauType"` // for now it's probably nil?
	Name     string `json:"name"`
	NodeLoc
}

func (n AstLocal[T]) Type() string {
	return "AstLocal"
}

func (n AstLocal[T]) String() string {
	var b strings.Builder

	b.WriteString("LuauType:")
	if n.LuauType != nil {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(*n.LuauType), 4))
	}
	b.WriteString(fmt.Sprintf("\nName: %s\n", n.Name))
	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))

	return b.String()
}

type AstStatAssign[T any] struct {
	NodeLoc
	Vars   []T `json:"vars"`
	Values []T `json:"values"`
}

func (n AstStatAssign[T]) Type() string {
	return "AstStatAssign"
}

func (n AstStatAssign[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nVars:")
	for _, v := range n.Vars {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(v), 4))
	}
	b.WriteString("\nValues:")
	for _, v := range n.Values {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(v), 4))
	}

	return b.String()
}

type AstStatBlock[T any] struct {
	NodeLoc
	HasEnd            bool       `json:"hasEnd"`
	Body              []T        `json:"body"`
	CommentsContained *[]Comment // not in json
}

func (n AstStatBlock[T]) Type() string {
	return "AstStatBlock"
}

func (n AstStatBlock[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nHasEnd: %t", n.HasEnd))
	b.WriteString("\nBody:\n")

	for _, node := range n.Body {
		b.WriteString(indentStart(StringMaybeEvaluated(node), 4))
		b.WriteString("\n\n")
	}

	return b.String()
}

type AstStatBreak struct {
	NodeLoc
}

func (n AstStatBreak) Type() string {
	return "AstStatBreak"
}

func (n AstStatBreak) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))

	return b.String()
}

type AstStatCompoundAssign[T any] struct {
	NodeLoc
	Op    string `json:"op"`
	Var   T      `json:"var"`
	Value T      `json:"value"`
}

func (n AstStatCompoundAssign[T]) Type() string {
	return "AstStatCompoundAssign"
}

func (n AstStatCompoundAssign[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nOp: %s", n.Op))
	b.WriteString("\nVar:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Var), 4))
	b.WriteString("\nValue:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Value), 4))

	return b.String()
}

type AstStatContinue struct {
	NodeLoc
}

func (n AstStatContinue) Type() string {
	return "AstStatContinue"
}

func (n AstStatContinue) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))

	return b.String()
}

type AstStatDeclareClass[T any] struct {
	NodeLoc
	Name      string  `json:"name"`
	SuperName *string `json:"superName"`
	Props     []T     `json:"props"`
	Indexer   *T      `json:"indexer"`
}

func (n AstStatDeclareClass[T]) Type() string {
	return "AstStatDeclareClass"
}

func (n AstStatDeclareClass[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nName: %s", n.Name))
	if n.SuperName != nil {
		b.WriteString(fmt.Sprintf("\nSuperName: %s", *n.SuperName))
	}
	b.WriteString("\nProps:")
	for _, prop := range n.Props {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(prop), 4))
	}
	b.WriteString("\nIndexer:")
	if n.Indexer != nil {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(*n.Indexer), 4))
	}

	return b.String()
}

type AstStatExpr[T any] struct {
	NodeLoc
	Expr T `json:"expr"`
}

func (n AstStatExpr[T]) Type() string {
	return "AstStatExpr"
}

func (n AstStatExpr[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))
	b.WriteByte('\n')

	return b.String()
}

type AstStatFor[T any] struct {
	NodeLoc
	Var   T               `json:"var"`
	From  T               `json:"from"`
	To    T               `json:"to"`
	Step  *T              `json:"step"`
	Body  AstStatBlock[T] `json:"body"`
	HasDo bool            `json:"hasDo"`
}

func (n AstStatFor[T]) Type() string {
	return "AstStatFor"
}

func (n AstStatFor[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nVar:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Var), 4))
	b.WriteString("\nFrom:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.From), 4))
	b.WriteString("\nTo:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.To), 4))
	b.WriteString("\nStep:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Step), 4))
	b.WriteString("\nBody:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Body), 4))
	b.WriteString(fmt.Sprintf("\nHasDo: %t\n", n.HasDo))

	return b.String()
}

type AstStatForIn[T any] struct {
	NodeLoc
	Vars   []AstLocal[T]   `json:"vars"`
	Values []T             `json:"values"`
	Body   AstStatBlock[T] `json:"body"`
	HasIn  bool            `json:"hasIn"`
	HasDo  bool            `json:"hasDo"`
}

func (n AstStatForIn[T]) Type() string {
	return "AstStatForIn"
}

func (n AstStatForIn[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nVars:")
	for _, v := range n.Vars {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(v), 4))
	}
	b.WriteString("\nValues:")
	for _, v := range n.Values {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(v), 4))
	}
	b.WriteString("\nBody:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Body), 4))
	b.WriteString(fmt.Sprintf("\nHasIn: %t\n", n.HasIn))
	b.WriteString(fmt.Sprintf("HasDo: %t\n", n.HasDo))

	return b.String()
}

type AstStatFunction[T any] struct {
	NodeLoc
	Name T                  `json:"name"`
	Func AstExprFunction[T] `json:"func"`
}

func (n AstStatFunction[T]) Type() string {
	return "AstStatFunction"
}

func (n AstStatFunction[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nName:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Name), 4))
	b.WriteString("\nFunc:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Func), 4))

	return b.String()
}

type AstStatIf[T any] struct {
	NodeLoc
	Condition T               `json:"condition"`
	ThenBody  AstStatBlock[T] `json:"thenbody"`
	ElseBody  *T              `json:"elsebody"` // StatBlock[T] | StatIf[T]
	HasThen   bool            `json:"hasThen"`
}

func (n AstStatIf[T]) Type() string {
	return "AstStatIf"
}

func (n AstStatIf[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nCondition:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Condition), 4))
	b.WriteString("\nThenBody:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.ThenBody), 4))
	b.WriteString("\nElseBody:\n")
	if n.ElseBody != nil {
		b.WriteString(indentStart(StringMaybeEvaluated(*n.ElseBody), 4))
		b.WriteByte('\n')
	}
	b.WriteString(fmt.Sprintf("HasThen: %t\n", n.HasThen))

	return b.String()
}

type AstStatLocal[T any] struct {
	NodeLoc
	Vars   []AstLocal[T] `json:"vars"`
	Values []T           `json:"values"`
}

func (n AstStatLocal[T]) Type() string {
	return "AstStatLocal"
}

func (n AstStatLocal[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nVars:")
	for _, v := range n.Vars {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(v), 4))
	}
	b.WriteString("\nValues:")
	for _, v := range n.Values {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(v), 4))
	}

	return b.String()
}

type AstStatLocalFunction[T any] struct {
	NodeLoc
	Name AstLocal[T]        `json:"name"`
	Func AstExprFunction[T] `json:"func"`
}

func (n AstStatLocalFunction[T]) Type() string {
	return "AstStatLocalFunction"
}

func (n AstStatLocalFunction[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nName:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Name), 4))
	b.WriteString("\nFunc:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Func), 4))

	return b.String()
}

type AstStatRepeat[T any] struct {
	NodeLoc
	Condition T               `json:"condition"`
	Body      AstStatBlock[T] `json:"body"`
}

func (n AstStatRepeat[T]) Type() string {
	return "AstStatRepeat"
}

func (n AstStatRepeat[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nCondition:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Condition), 4))
	b.WriteString("\nBody:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Body), 4))

	return b.String()
}

type AstStatReturn[T any] struct {
	NodeLoc
	List []T `json:"list"`
}

func (n AstStatReturn[T]) Type() string {
	return "AstStatReturn"
}

func (n AstStatReturn[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nList:")

	for _, item := range n.List {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(item), 4))
	}

	return b.String()
}

type AstStatTypeAlias[T any] struct {
	NodeLoc
	Name         string            `json:"name"`
	Generics     []GenericType     `json:"generics"`
	GenericPacks []GenericTypePack `json:"genericPacks"` // genericPacks always come after the generics
	Value        T                 `json:"value"`
	Exported     bool              `json:"exported"`
}

func (n AstStatTypeAlias[T]) Type() string {
	return "AstStatTypeAlias"
}

func (n AstStatTypeAlias[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nName: %s", n.Name))
	b.WriteString("\nGenerics:")
	for _, g := range n.Generics {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(g), 4))
	}
	b.WriteString("\nGenericPacks:")
	for _, gp := range n.GenericPacks {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(gp), 4))
	}
	b.WriteString("\nValue:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Value), 4))
	b.WriteString(fmt.Sprintf("\nExported: %t\n", n.Exported))

	return b.String()
}

type AstStatWhile[T any] struct {
	NodeLoc
	Condition T               `json:"condition"`
	Body      AstStatBlock[T] `json:"body"`
	HasDo     bool            `json:"hasDo"`
}

func (n AstStatWhile[T]) Type() string {
	return "AstStatWhile"
}

func (n AstStatWhile[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nCondition:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Condition), 4))
	b.WriteString("\nBody:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Body), 4))
	b.WriteString(fmt.Sprintf("\nHasDo: %t\n", n.HasDo))

	return b.String()
}

type AstTableProp[T any] struct {
	Name string `json:"name"`
	NodeLoc
	PropType T `json:"propType"`
}

func (n AstTableProp[T]) Type() string {
	return "AstTableProp"
}

func (n AstTableProp[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nName: %s", n.Name))
	b.WriteString("\nPropType:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.PropType), 4))

	return b.String()
}

type AstTypeFunction[T any] struct {
	NodeLoc
	Attributes   []AstAttr              `json:"attributes"`
	Generics     []GenericType          `json:"generics"`
	GenericPacks []GenericTypePack      `json:"genericPacks"`
	ArgTypes     AstTypeList[T]         `json:"argTypes"`
	ArgNames     []*AstArgumentName     `json:"argNames"`
	ReturnTypes  AstTypePackExplicit[T] `json:"returnTypes"`
}

func (n AstTypeFunction[T]) Type() string {
	return "AstTypeFunction"
}

func (n AstTypeFunction[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nAttributes:")
	for _, attr := range n.Attributes {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(attr), 4))
	}
	b.WriteString("\nGenerics:")
	for _, gen := range n.Generics {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(gen), 4))
	}
	b.WriteString("\nGenericPacks:")
	for _, pack := range n.GenericPacks {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(pack), 4))
	}
	b.WriteString("\nArgTypes:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.ArgTypes), 4))
	b.WriteString("\nArgNames:")
	for _, name := range n.ArgNames {
		b.WriteByte('\n')
		if name == nil {
			b.WriteString(indentStart("<nil>", 4))
			continue
		}
		b.WriteString(indentStart(StringMaybeEvaluated(*name), 4))
	}
	b.WriteString("\nReturnTypes:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.ReturnTypes), 4))

	return b.String()
}

type AstTypeGroup[T any] struct {
	NodeLoc
	Inner T `json:"inner"`
}

func (n AstTypeGroup[T]) Type() string {
	return "AstTypeGroup"
}

func (n AstTypeGroup[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nInner:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Inner), 4))

	return b.String()
}

type AstTypeList[T any] struct {
	ASTNode
	Types    []T `json:"types"`
	TailType *T  `json:"tailType"`
}

func (n AstTypeList[T]) GetLocation() Location {
	return Location{}
}

func (n AstTypeList[T]) Type() string {
	return "AstTypeList"
}

func (n AstTypeList[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString("Types:")

	for _, typ := range n.Types {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(typ), 4))
	}

	b.WriteString("\nTailType:")
	if n.TailType != nil {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(*n.TailType), 4))
	}

	return b.String()
}

type AstTypeOptional struct {
	NodeLoc
}

func (AstTypeOptional) Type() string {
	return "AstTypeOptional"
}

func (n AstTypeOptional) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))

	return b.String()
}

type AstTypePackExplicit[T any] struct {
	NodeLoc
	TypeList AstTypeList[T] `json:"typeList"`
}

func (n AstTypePackExplicit[T]) Type() string {
	return "AstTypePackExplicit"
}

func (n AstTypePackExplicit[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nTypeList:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.TypeList), 4))

	return b.String()
}

type AstTypePackGeneric struct {
	NodeLoc
	GenericName string `json:"genericName"`
}

func (n AstTypePackGeneric) Type() string {
	return "AstTypePackGeneric"
}

func (n AstTypePackGeneric) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nGenericName: %s", n.GenericName))

	return b.String()
}

type AstTypePackVariadic[T any] struct {
	NodeLoc
	VariadicType T `json:"variadicType"`
}

func (n AstTypePackVariadic[T]) Type() string {
	return "AstTypePackVariadic"
}

func (n AstTypePackVariadic[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nVariadicType: %s", StringMaybeEvaluated(n.VariadicType)))

	return b.String()
}

type AstTypeReference[T any] struct {
	NodeLoc
	Name         string   `json:"name"`
	NameLocation Location `json:"nameLocation"`
	Parameters   []T      `json:"parameters"`
}

func (n AstTypeReference[T]) Type() string {
	return "AstTypeReference"
}

func (n AstTypeReference[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nName: %s", n.Name))
	b.WriteString(fmt.Sprintf("\nNameLocation: %s", n.NameLocation))
	b.WriteString("\nParameters:")

	for _, param := range n.Parameters {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(param), 4))
	}

	return b.String()
}

type AstTypeSingletonBool struct {
	NodeLoc
	Value bool `json:"value"`
}

func (n AstTypeSingletonBool) Type() string {
	return "AstTypeSingletonBool"
}

func (n AstTypeSingletonBool) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nValue: %t", n.Value))

	return b.String()
}

type AstTypeSingletonString struct {
	NodeLoc
	Value string `json:"value"`
}

func (n AstTypeSingletonString) Type() string {
	return "AstTypeSingletonString"
}

func (n AstTypeSingletonString) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nValue: %s", n.Value))

	return b.String()
}

type AstTableIndexer[T any] struct {
	Location   Location `json:"location"`
	IndexType  T        `json:"indexType"`
	ResultType T        `json:"resultType"`
}

func (n AstTableIndexer[T]) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nIndexType:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.IndexType), 4))
	b.WriteString("\nResultType:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.ResultType), 4))

	return b.String()
}

type AstTypeTable[T any] struct {
	NodeLoc
	Props   []T                 `json:"props"`
	Indexer *AstTableIndexer[T] `json:"indexer"`
}

func (n AstTypeTable[T]) Type() string {
	return "AstTypeTable"
}

func (n AstTypeTable[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nProps:")

	for _, prop := range n.Props {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(prop), 4))
	}

	b.WriteString("\nIndexer:")
	if n.Indexer != nil {
		b.WriteByte('\n')
		b.WriteString(indentStart(n.Indexer.String(), 4))
	}

	return b.String()
}

// lol
type AstTypeTypeof[T any] struct {
	NodeLoc
	Expr T `json:"expr"`
}

func (n AstTypeTypeof[T]) Type() string {
	return "AstTypeTypeof"
}

func (n AstTypeTypeof[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))

	return b.String()
}

type AstTypeUnion[T any] struct {
	NodeLoc
	Types []T `json:"types"`
}

func (n AstTypeUnion[T]) Type() string {
	return "AstTypeUnion"
}

func (n AstTypeUnion[T]) String() string {
	var b strings.Builder

	b.WriteString(n.ASTNode.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nTypes:")

	for _, typ := range n.Types {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(typ), 4))
	}

	return b.String()
}
