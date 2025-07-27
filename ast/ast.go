package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// dot net vibez
type INode interface {
	String() string
	Type() string
}

type Position struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func (p Position) String() string {
	return fmt.Sprintf("%d,%d", p.Line, p.Column)
}

// Location represents a source location range
type Location struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

func (l Location) String() string {
	return fmt.Sprintf("%s - %s", l.Start, l.End)
}

// UnmarshalJSON custom unmarshaler for location strings like "0,0 - 2,0"
func (l *Location) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	_, err := fmt.Sscanf(s, "%d,%d - %d,%d", &l.Start.Line, &l.Start.Column, &l.End.Line, &l.End.Column)
	return err
}

// base for every node
type Node struct {
	Type string `json:"type"`
}

func (n Node) String() string {
	return fmt.Sprintf("Type: %s\n", n.Type)
}

func StringMaybeEvaluated(val any) string {
	if v, ok := val.(json.RawMessage); ok {
		var node Node
		if err := json.Unmarshal(v, &node); err != nil {
			return fmt.Sprintf("Error decoding Node: %v", err)
		}
		return node.String()
	}
	return fmt.Sprintf("%v", val)
}

// ast

type Comment struct {
	Node
	Location Location `json:"location"`
}

func (c Comment) String() string {
	var b strings.Builder

	b.WriteString(c.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s\n", c.Location))

	return b.String()
}

type AST[T any] struct {
	Root             T         `json:"root"`
	CommentLocations []Comment `json:"commentLocations"`
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

func DecodeAST(data json.RawMessage) (AST[INode], error) {
	var raw AST[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return AST[INode]{}, fmt.Errorf("error decoding AST: %v", err)
	}

	rootNode, err := decodeNode(raw.Root)
	if err != nil {
		return AST[INode]{}, fmt.Errorf("error decoding root node: %v", err)
	}

	return AST[INode]{
		Root:             rootNode,
		CommentLocations: raw.CommentLocations,
	}, nil
}

// node types

type ArgumentName struct {
	Node
	Name     string   `json:"name"`
	Location Location `json:"location"`
}

func (a ArgumentName) Type() string {
	return "AstArgumentName"
}

func (a ArgumentName) String() string {
	var b strings.Builder

	b.WriteString(a.Node.String())
	b.WriteString(fmt.Sprintf("Name: %s\n", a.Name))
	b.WriteString(fmt.Sprintf("Location: %s\n", a.Location))

	return b.String()
}

func DecodeArgumentName(data json.RawMessage) (INode, error) {
	var raw ArgumentName
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type Attr struct {
	Node
	Location Location `json:"location"`
	Name     string   `json:"name"`
}

func (a Attr) Type() string {
	return "AstAttr"
}

func (a Attr) String() string {
	var b strings.Builder

	b.WriteString(a.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s\n", a.Location))
	b.WriteString(fmt.Sprintf("Name: %s\n", a.Name))

	return b.String()
}

func DecodeAttr(data json.RawMessage) (INode, error) {
	var raw Attr
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type ExprBinary[T any] struct {
	Node
	Location Location `json:"location"`
	Op       string   `json:"op"`
	Left     T        `json:"left"`
	Right    T        `json:"right"`
}

func (n ExprBinary[T]) Type() string {
	return "AstExprBinary"
}

func (n ExprBinary[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nOp: %s", n.Op))
	b.WriteString("\nLeft:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Left), 4))
	b.WriteString("\nRight:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Right), 4))

	return b.String()
}

func DecodeExprBinary(data json.RawMessage) (INode, error) {
	var raw ExprBinary[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	left, err := decodeNode(raw.Left)
	if err != nil {
		return nil, fmt.Errorf("error decoding left: %v", err)
	}

	right, err := decodeNode(raw.Right)
	if err != nil {
		return nil, fmt.Errorf("error decoding right: %v", err)
	}

	return ExprBinary[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Op:       raw.Op,
		Left:     left,
		Right:    right,
	}, nil
}

type ExprCall[T any] struct {
	Node
	Location    Location `json:"location"`
	Func        T        `json:"func"`
	Args        []T      `json:"args"`
	Self        bool     `json:"self"`
	ArgLocation Location `json:"argLocation"`
}

func (n ExprCall[T]) Type() string {
	return "AstExprCall"
}

func (n ExprCall[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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

func DecodeExprCall(data json.RawMessage) (INode, error) {
	var raw ExprCall[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	funcNode, err := decodeNode(raw.Func)
	if err != nil {
		return nil, fmt.Errorf("error decoding func: %v", err)
	}

	args := make([]INode, len(raw.Args))
	for i, arg := range raw.Args {
		n, err := decodeNode(arg)
		if err != nil {
			return nil, fmt.Errorf("error decoding arg node: %v", err)
		}
		args[i] = n
	}

	return ExprCall[INode]{
		Node:        raw.Node,
		Location:    raw.Location,
		Func:        funcNode,
		Args:        args,
		Self:        raw.Self,
		ArgLocation: raw.ArgLocation,
	}, nil
}

type ExprConstantBool struct {
	Node
	Location Location `json:"location"`
	Value    bool     `json:"value"`
}

func (n ExprConstantBool) Type() string {
	return "AstExprConstantBool"
}

func (n ExprConstantBool) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nValue: %t", n.Value))

	return b.String()
}

func DecodeExprConstantBool(data json.RawMessage) (INode, error) {
	var raw ExprConstantBool
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type ExprConstantNil struct {
	Node
	Location Location `json:"location"`
}

func (n ExprConstantNil) Type() string {
	return "AstExprConstantNil"
}

func (n ExprConstantNil) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))

	return b.String()
}

func DecodeExprConstantNil(data json.RawMessage) (INode, error) {
	var raw ExprConstantNil
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type ExprConstantNumber struct {
	Node
	Location Location `json:"location"`
	Value    float64  `json:"value"`
}

func (n ExprConstantNumber) Type() string {
	return "AstExprConstantNumber"
}

func (n ExprConstantNumber) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nValue: %f", n.Value))

	return b.String()
}

func DecodeExprConstantNumber(data json.RawMessage) (INode, error) {
	var raw ExprConstantNumber
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type ExprConstantString struct {
	Node
	Location Location `json:"location"`
	Value    string   `json:"value"`
}

func (n ExprConstantString) Type() string {
	return "AstExprConstantString"
}

func (n ExprConstantString) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nValue: %s", n.Value))

	return b.String()
}

func DecodeExprConstantString(data json.RawMessage) (INode, error) {
	var raw ExprConstantString
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type ExprFunction[T any] struct {
	Node
	Location       Location `json:"location"`
	Attributes     []T      `json:"attributes"`
	Generics       []T      `json:"generics"`
	GenericPacks   []T      `json:"genericPacks"`
	Args           []T      `json:"args"`
	Vararg         bool     `json:"vararg"`
	VarargLocation Location `json:"varargLocation"`
	Body           T        `json:"body"`
	FunctionDepth  int      `json:"functionDepth"`
	Debugname      string   `json:"debugname"`
}

func (n ExprFunction[T]) Type() string {
	return "AstExprFunction"
}

func (n ExprFunction[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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
	b.WriteString(fmt.Sprintf("\nVararg: %t", n.Vararg))
	b.WriteString(fmt.Sprintf("\nVarargLocation: %s", n.VarargLocation))
	b.WriteString("\nBody:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Body), 4))
	b.WriteString(fmt.Sprintf("\nFunctionDepth: %d", n.FunctionDepth))
	b.WriteString(fmt.Sprintf("\nDebugname: %s", n.Debugname))

	return b.String()
}

func DecodeExprFunction(data json.RawMessage) (INode, error) {
	var raw ExprFunction[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	attributes := make([]INode, len(raw.Attributes))
	for i, attr := range raw.Attributes {
		n, err := decodeNode(attr)
		if err != nil {
			return nil, fmt.Errorf("error decoding attribute node: %v", err)
		}
		attributes[i] = n
	}

	generics := make([]INode, len(raw.Generics))
	for i, gen := range raw.Generics {
		n, err := decodeNode(gen)
		if err != nil {
			return nil, fmt.Errorf("error decoding generic node: %v", err)
		}
		generics[i] = n
	}

	genericPacks := make([]INode, len(raw.GenericPacks))
	for i, pack := range raw.GenericPacks {
		n, err := decodeNode(pack)
		if err != nil {
			return nil, fmt.Errorf("error decoding generic pack node: %v", err)
		}
		genericPacks[i] = n
	}

	args := make([]INode, len(raw.Args))
	for i, arg := range raw.Args {
		n, err := decodeNode(arg)
		if err != nil {
			return nil, fmt.Errorf("error decoding arg node: %v", err)
		}
		args[i] = n
	}

	bodyNode, err := decodeNode(raw.Body)
	if err != nil {
		return nil, fmt.Errorf("error decoding body node: %v", err)
	}

	return ExprFunction[INode]{
		Node:           raw.Node,
		Location:       raw.Location,
		Attributes:     attributes,
		Generics:       generics,
		GenericPacks:   genericPacks,
		Args:           args,
		Vararg:         raw.Vararg,
		VarargLocation: raw.VarargLocation,
		Body:           bodyNode,
		FunctionDepth:  raw.FunctionDepth,
		Debugname:      raw.Debugname,
	}, nil
}

type ExprGlobal struct {
	Node
	Location Location `json:"location"`
	Global   string   `json:"global"`
}

func (n ExprGlobal) Type() string {
	return "AstExprGlobal"
}

func (n ExprGlobal) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nGlobal: %s", n.Global))

	return b.String()
}

func DecodeExprGlobal(data json.RawMessage) (INode, error) {
	var raw ExprGlobal
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type ExprGroup[T any] struct {
	Node
	Location Location `json:"location"`
	Expr     T        `json:"expr"` // only contains one expression right? strange when you first think about it
}

func (n ExprGroup[T]) Type() string {
	return "AstExprGroup"
}

func (n ExprGroup[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))

	return b.String()
}

func DecodeExprGroup(data json.RawMessage) (INode, error) {
	var raw ExprGroup[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	exprNode, err := decodeNode(raw.Expr)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %v", err)
	}

	return ExprGroup[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Expr:     exprNode,
	}, nil
}

type ExprIfElse[T any] struct {
	Node
	Location  Location `json:"location"`
	Condition T        `json:"condition"`
	HasThen   bool     `json:"hasThen"`
	TrueExpr  T        `json:"trueExpr"`
	HasElse   bool     `json:"hasElse"`
	FalseExpr T        `json:"falseExpr"`
}

func (n ExprIfElse[T]) Type() string {
	return "AstExprIfElse"
}

func (n ExprIfElse[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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

func DecodeExprIfElse(data json.RawMessage) (INode, error) {
	var raw ExprIfElse[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	conditionNode, err := decodeNode(raw.Condition)
	if err != nil {
		return nil, fmt.Errorf("error decoding condition: %v", err)
	}

	trueExprNode, err := decodeNode(raw.TrueExpr)
	if err != nil {
		return nil, fmt.Errorf("error decoding true expression: %v", err)
	}

	falseExprNode, err := decodeNode(raw.FalseExpr)
	if err != nil {
		return nil, fmt.Errorf("error decoding false expression: %v", err)
	}

	return ExprIfElse[INode]{
		Node:      raw.Node,
		Location:  raw.Location,
		Condition: conditionNode,
		HasThen:   raw.HasThen,
		TrueExpr:  trueExprNode,
		HasElse:   raw.HasElse,
		FalseExpr: falseExprNode,
	}, nil
}

type ExprIndexExpr[T any] struct {
	Node
	Location Location `json:"location"`
	Expr     T        `json:"expr"`
	Index    T        `json:"index"`
}

func (n ExprIndexExpr[T]) Type() string {
	return "AstExprIndexExpr"
}

func (n ExprIndexExpr[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))
	b.WriteString("\nIndex:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Index), 4))

	return b.String()
}

func DecodeExprIndexExpr(data json.RawMessage) (INode, error) {
	var raw ExprIndexExpr[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	exprNode, err := decodeNode(raw.Expr)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %v", err)
	}

	indexNode, err := decodeNode(raw.Index)
	if err != nil {
		return nil, fmt.Errorf("error decoding index: %v", err)
	}

	return ExprIndexExpr[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Expr:     exprNode,
		Index:    indexNode,
	}, nil
}

type ExprIndexName[T any] struct {
	Node
	Location      Location `json:"location"`
	Expr          T        `json:"expr"`
	Index         string   `json:"index"`
	IndexLocation Location `json:"indexLocation"`
	Op            string   `json:"op"`
}

func (n ExprIndexName[T]) Type() string {
	return "AstExprIndexName"
}

func (n ExprIndexName[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))
	b.WriteString(fmt.Sprintf("\nIndex: %s", n.Index))

	return b.String()
}

func DecodeExprIndexName(data json.RawMessage) (INode, error) {
	var raw ExprIndexName[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	exprNode, err := decodeNode(raw.Expr)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %v", err)
	}

	return ExprIndexName[INode]{
		Node:          raw.Node,
		Location:      raw.Location,
		Expr:          exprNode,
		Index:         raw.Index,
		IndexLocation: raw.IndexLocation,
		Op:            raw.Op,
	}, nil
}

type ExprInterpString[T any] struct {
	Node
	Location    Location `json:"location"`
	Strings     []string `json:"strings"`
	Expressions []T      `json:"expressions"`
}

func (n ExprInterpString[T]) Type() string {
	return "AstExprInterpString"
}

func (n ExprInterpString[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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

func DecodeExprInterpString(data json.RawMessage) (INode, error) {
	var raw ExprInterpString[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	expressions := make([]INode, len(raw.Expressions))
	for i, expr := range raw.Expressions {
		n, err := decodeNode(expr)
		if err != nil {
			return nil, fmt.Errorf("error decoding expression node: %v", err)
		}
		expressions[i] = n
	}

	return ExprInterpString[INode]{
		Node:        raw.Node,
		Location:    raw.Location,
		Strings:     raw.Strings,
		Expressions: expressions,
	}, nil
}

type ExprLocal[T any] struct {
	Node
	Location Location `json:"location"`
	Local    T        `json:"local"`
}

func (n ExprLocal[T]) Type() string {
	return "AstExprLocal"
}

func (n ExprLocal[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nLocal:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Local), 4))

	return b.String()
}

func DecodeExprLocal(data json.RawMessage) (INode, error) {
	var raw ExprLocal[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	localNode, err := decodeNode(raw.Local)
	if err != nil {
		return nil, fmt.Errorf("error decoding local: %v", err)
	}

	return ExprLocal[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Local:    localNode,
	}, nil
}

type ExprTable[T any] struct {
	Node
	Location Location `json:"location"`
	Items    []T      `json:"items"`
}

func (n ExprTable[T]) Type() string {
	return "AstExprTable"
}

func (n ExprTable[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nItems:")

	for _, item := range n.Items {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(item), 4))
	}

	return b.String()
}

func DecodeExprTable(data json.RawMessage) (INode, error) {
	var raw ExprTable[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	items := make([]INode, len(raw.Items))
	for i, item := range raw.Items {
		n, err := decodeNode(item)
		if err != nil {
			return nil, fmt.Errorf("error decoding item node: %v", err)
		}
		items[i] = n
	}

	return ExprTable[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Items:    items,
	}, nil
}

type ExprTableItem[T any] struct {
	Node
	Kind  string `json:"kind"`
	Key   *T     `json:"key"`
	Value T      `json:"value"`
}

func (n ExprTableItem[T]) Type() string {
	return "AstExprTableItem"
}

func (n ExprTableItem[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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

func DecodeExprTableItem(data json.RawMessage) (INode, error) {
	var raw ExprTableItem[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	var keyNodeMaybe *INode
	if raw.Key != nil {
		keyNode, err := decodeNode(*raw.Key)
		if err != nil {
			return nil, fmt.Errorf("error decoding key: %v", err)
		}
		keyNodeMaybe = &keyNode
	}

	valueNode, err := decodeNode(raw.Value)
	if err != nil {
		return nil, fmt.Errorf("error decoding value: %v", err)
	}

	return ExprTableItem[INode]{
		Node:  raw.Node,
		Kind:  raw.Kind,
		Key:   keyNodeMaybe,
		Value: valueNode,
	}, nil
}

type ExprTypeAssertion[T any] struct {
	Node
	Location   Location `json:"location"`
	Expr       T        `json:"expr"`
	Annotation T        `json:"annotation"`
}

func (n ExprTypeAssertion[T]) Type() string {
	return "AstExprTypeAssertion"
}

func (n ExprTypeAssertion[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))
	b.WriteString("\nAnnotation:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Annotation), 4))

	return b.String()
}

func DecodeExprTypeAssertion(data json.RawMessage) (INode, error) {
	var raw ExprTypeAssertion[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	exprNode, err := decodeNode(raw.Expr)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %v", err)
	}

	annotationNode, err := decodeNode(raw.Annotation)
	if err != nil {
		return nil, fmt.Errorf("error decoding annotation: %v", err)
	}

	return ExprTypeAssertion[INode]{
		Node:       raw.Node,
		Location:   raw.Location,
		Expr:       exprNode,
		Annotation: annotationNode,
	}, nil
}

type ExprVarargs struct {
	Node
	Location Location `json:"location"`
}

func (n ExprVarargs) Type() string {
	return "AstExprVarargs"
}

func (n ExprVarargs) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))

	return b.String()
}

func DecodeExprVarargs(data json.RawMessage) (INode, error) {
	var raw ExprVarargs
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type ExprUnary[T any] struct {
	Node
	Location Location `json:"location"`
	Op       string   `json:"op"`
	Expr     T        `json:"expr"`
}

func (n ExprUnary[T]) Type() string {
	return "AstExprUnary"
}

func (n ExprUnary[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nOp: %s", n.Op))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))

	return b.String()
}

func DecodeExprUnary(data json.RawMessage) (INode, error) {
	var raw ExprUnary[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	exprNode, err := decodeNode(raw.Expr)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %v", err)
	}

	return ExprUnary[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Op:       raw.Op,
		Expr:     exprNode,
	}, nil
}

type GenericType struct {
	Node
	Name string `json:"name"`
}

func (g GenericType) Type() string {
	return "AstGenericType"
}

func (g GenericType) String() string {
	var b strings.Builder

	b.WriteString(g.Node.String())
	b.WriteString(fmt.Sprintf("Name: %s", g.Name))

	return b.String()
}

func DecodeGenericType(data json.RawMessage) (INode, error) {
	var raw GenericType
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type Local[T any] struct {
	LuauType *T     `json:"luauType"` // for now it's probably nil?
	Name     string `json:"name"`
	Node
	Location Location `json:"location"`
}

func (n Local[T]) Type() string {
	return "AstLocal"
}

func (n Local[T]) String() string {
	var b strings.Builder

	b.WriteString("LuauType:")
	if n.LuauType != nil {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(*n.LuauType), 4))
	}
	b.WriteString(fmt.Sprintf("\nName: %s\n", n.Name))
	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))

	return b.String()
}

func DecodeLocal(data json.RawMessage) (INode, error) {
	var raw Local[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	var luauTypeMaybe *INode
	if raw.LuauType != nil {
		luauTypeNode, err := decodeNode(*raw.LuauType)
		if err != nil {
			return nil, fmt.Errorf("error decoding luau type: %v", err)
		}
		luauTypeMaybe = &luauTypeNode
	}

	return Local[INode]{
		LuauType: luauTypeMaybe,
		Name:     raw.Name,
		Node:     raw.Node,
		Location: raw.Location,
	}, nil
}

type StatAssign[T any] struct {
	Node
	Location Location `json:"location"`
	Vars     []T      `json:"vars"`
	Values   []T      `json:"values"`
}

func (n StatAssign[T]) Type() string {
	return "AstStatAssign"
}

func (n StatAssign[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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

func DecodeStatAssign(data json.RawMessage) (INode, error) {
	var raw StatAssign[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	vars := make([]INode, len(raw.Vars))
	for i, v := range raw.Vars {
		n, err := decodeNode(v)
		if err != nil {
			return nil, fmt.Errorf("error decoding var node: %v", err)
		}
		vars[i] = n
	}

	values := make([]INode, len(raw.Values))
	for i, v := range raw.Values {
		n, err := decodeNode(v)
		if err != nil {
			return nil, fmt.Errorf("error decoding value node: %v", err)
		}
		values[i] = n
	}

	return StatAssign[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Vars:     vars,
		Values:   values,
	}, nil
}

type StatBlock[T any] struct {
	Node
	Location Location `json:"location"`
	HasEnd   bool     `json:"hasEnd"`
	Body     []T      `json:"body"`
}

func (n StatBlock[T]) Type() string {
	return "AstStatBlock"
}

func (n StatBlock[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nHasEnd: %t", n.HasEnd))
	b.WriteString("\nBody:\n")

	for _, node := range n.Body {
		b.WriteString(indentStart(StringMaybeEvaluated(node), 4))
		b.WriteString("\n\n")
	}

	return b.String()
}

func DecodeStatBlock(data json.RawMessage) (INode, error) {
	var raw StatBlock[json.RawMessage] // rawblocks man
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	body := make([]INode, len(raw.Body))
	for i, bn := range raw.Body {
		n, err := decodeNode(bn)
		if err != nil {
			return nil, fmt.Errorf("error decoding body node: %v", err)
		}
		body[i] = n
	}

	return StatBlock[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		HasEnd:   raw.HasEnd,
		Body:     body,
	}, nil
}

type StatBreak struct {
	Node
	Location Location `json:"location"`
}

func (n StatBreak) Type() string {
	return "AstStatBreak"
}

func (n StatBreak) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))

	return b.String()
}

func DecodeStatBreak(data json.RawMessage) (INode, error) {
	var raw StatBreak
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type StatCompoundAssign[T any] struct {
	Node
	Location Location `json:"location"`
	Op       string   `json:"op"`
	Var      T        `json:"var"`
	Value    T        `json:"value"`
}

func (n StatCompoundAssign[T]) Type() string {
	return "AstStatCompoundAssign"
}

func (n StatCompoundAssign[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nOp: %s", n.Op))
	b.WriteString("\nVar:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Var), 4))
	b.WriteString("\nValue:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Value), 4))

	return b.String()
}

func DecodeStatCompoundAssign(data json.RawMessage) (INode, error) {
	var raw StatCompoundAssign[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	varNode, err := decodeNode(raw.Var)
	if err != nil {
		return nil, fmt.Errorf("error decoding var: %v", err)
	}

	valueNode, err := decodeNode(raw.Value)
	if err != nil {
		return nil, fmt.Errorf("error decoding value: %v", err)
	}

	return StatCompoundAssign[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Op:       raw.Op,
		Var:      varNode,
		Value:    valueNode,
	}, nil
}

type StatContinue struct {
	Node
	Location Location `json:"location"`
}

func (n StatContinue) Type() string {
	return "AstStatContinue"
}

func (n StatContinue) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))

	return b.String()
}

func DecodeStatContinue(data json.RawMessage) (INode, error) {
	var raw StatContinue
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type StatExpr[T any] struct {
	Node
	Location Location `json:"location"`
	Expr     T        `json:"expr"`
}

func (n StatExpr[T]) Type() string {
	return "AstStatExpr"
}

func (n StatExpr[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))
	b.WriteByte('\n')

	return b.String()
}

func DecodeStatExpr(data json.RawMessage) (INode, error) {
	var raw StatExpr[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	n, err := decodeNode(raw.Expr)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %v", err)
	}

	return StatExpr[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Expr:     n,
	}, nil
}

type StatFor[T any] struct {
	Node
	Location Location `json:"location"`
	Var      T        `json:"var"`
	From     T        `json:"from"`
	To       T        `json:"to"`
	Body     T        `json:"body"`
	HasDo    bool     `json:"hasDo"`
}

func (n StatFor[T]) Type() string {
	return "AstStatFor"
}

func (n StatFor[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nVar:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Var), 4))
	b.WriteString("\nFrom:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.From), 4))
	b.WriteString("\nTo:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.To), 4))
	b.WriteString("\nBody:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Body), 4))
	b.WriteString(fmt.Sprintf("\nHasDo: %t\n", n.HasDo))

	return b.String()
}

func DecodeStatFor(data json.RawMessage) (INode, error) {
	var raw StatFor[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	varNode, err := decodeNode(raw.Var)
	if err != nil {
		return nil, fmt.Errorf("error decoding var: %v", err)
	}

	fromNode, err := decodeNode(raw.From)
	if err != nil {
		return nil, fmt.Errorf("error decoding from: %v", err)
	}

	toNode, err := decodeNode(raw.To)
	if err != nil {
		return nil, fmt.Errorf("error decoding to: %v", err)
	}

	bodyNode, err := decodeNode(raw.Body)
	if err != nil {
		return nil, fmt.Errorf("error decoding body: %v", err)
	}

	return StatFor[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Var:      varNode,
		From:     fromNode,
		To:       toNode,
		Body:     bodyNode,
		HasDo:    raw.HasDo,
	}, nil
}

type StatForIn[T any] struct {
	Node
	Location Location `json:"location"`
	Vars     []T      `json:"vars"`
	Values   []T      `json:"values"`
	Body     T        `json:"body"`
	HasIn    bool     `json:"hasIn"`
	HasDo    bool     `json:"hasDo"`
}

func (n StatForIn[T]) Type() string {
	return "AstStatForIn"
}

func (n StatForIn[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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

func DecodeStatForIn(data json.RawMessage) (INode, error) {
	var raw StatForIn[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	vars := make([]INode, len(raw.Vars))
	for i, v := range raw.Vars {
		n, err := decodeNode(v)
		if err != nil {
			return nil, fmt.Errorf("error decoding var node: %v", err)
		}
		vars[i] = n
	}

	values := make([]INode, len(raw.Values))
	for i, v := range raw.Values {
		n, err := decodeNode(v)
		if err != nil {
			return nil, fmt.Errorf("error decoding value node: %v", err)
		}
		values[i] = n
	}

	bodyNode, err := decodeNode(raw.Body)
	if err != nil {
		return nil, fmt.Errorf("error decoding body: %v", err)
	}

	return StatForIn[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Vars:     vars,
		Values:   values,
		Body:     bodyNode,
		HasIn:    raw.HasIn,
		HasDo:    raw.HasDo,
	}, nil
}

type StatFunction[T any] struct {
	Node
	Location Location `json:"location"`
	Name     T        `json:"name"`
	Func     T        `json:"func"`
}

func (n StatFunction[T]) Type() string {
	return "AstStatFunction"
}

func (n StatFunction[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nName:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Name), 4))
	b.WriteString("\nFunc:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Func), 4))

	return b.String()
}

func DecodeStatFunction(data json.RawMessage) (INode, error) {
	var raw StatFunction[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	nameNode, err := decodeNode(raw.Name)
	if err != nil {
		return nil, fmt.Errorf("error decoding name: %v", err)
	}

	funcNode, err := decodeNode(raw.Func)
	if err != nil {
		return nil, fmt.Errorf("error decoding func: %v", err)
	}

	return StatFunction[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Name:     nameNode,
		Func:     funcNode,
	}, nil
}

type StatIf[T any] struct {
	Node
	Location  Location `json:"location"`
	Condition T        `json:"condition"`
	ThenBody  T        `json:"thenbody"`
	ElseBody  *T       `json:"elsebody"`
	HasThen   bool     `json:"hasThen"`
}

func (n StatIf[T]) Type() string {
	return "AstStatIf"
}

func (n StatIf[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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

func DecodeStatIf(data json.RawMessage) (INode, error) {
	var raw StatIf[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	condition, err := decodeNode(raw.Condition)
	if err != nil {
		return nil, fmt.Errorf("error decoding condition: %v", err)
	}

	thenBody, err := decodeNode(raw.ThenBody)
	if err != nil {
		return nil, fmt.Errorf("error decoding then body: %v", err)
	}

	var elseBodyMaybe *INode
	if raw.ElseBody != nil {
		elseBody, err := decodeNode(*raw.ElseBody)
		if err != nil {
			return nil, fmt.Errorf("error decoding else body: %v", err)
		}
		elseBodyMaybe = &elseBody
	}

	return StatIf[INode]{
		Node:      raw.Node,
		Location:  raw.Location,
		Condition: condition,
		ThenBody:  thenBody,
		ElseBody:  elseBodyMaybe,
		HasThen:   raw.HasThen,
	}, nil
}

type StatLocal[T any] struct {
	Node
	Location Location `json:"location"`
	Vars     []T      `json:"vars"`
	Values   []T      `json:"values"`
}

func (n StatLocal[T]) Type() string {
	return "AstStatLocal"
}

func (n StatLocal[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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

func DecodeStatLocal(data json.RawMessage) (INode, error) {
	var raw StatLocal[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	vars := make([]INode, len(raw.Vars))
	for i, v := range raw.Vars {
		n, err := decodeNode(v)
		if err != nil {
			return nil, fmt.Errorf("error decoding var node: %v", err)
		}
		vars[i] = n
	}

	values := make([]INode, len(raw.Values))
	for i, v := range raw.Values {
		n, err := decodeNode(v)
		if err != nil {
			return nil, fmt.Errorf("error decoding value node: %v", err)
		}
		values[i] = n
	}

	return StatLocal[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Vars:     vars,
		Values:   values,
	}, nil
}

type StatLocalFunction[T any] struct {
	Node
	Location Location `json:"location"`
	Name     T        `json:"name"`
	Func     T        `json:"func"`
}

func (n StatLocalFunction[T]) Type() string {
	return "AstStatLocalFunction"
}

func (n StatLocalFunction[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nName:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Name), 4))
	b.WriteString("\nFunc:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Func), 4))

	return b.String()
}

func DecodeStatLocalFunction(data json.RawMessage) (INode, error) {
	var raw StatLocalFunction[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	nameNode, err := decodeNode(raw.Name)
	if err != nil {
		return nil, fmt.Errorf("error decoding name: %v", err)
	}

	funcNode, err := decodeNode(raw.Func)
	if err != nil {
		return nil, fmt.Errorf("error decoding func: %v", err)
	}

	return StatLocalFunction[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Name:     nameNode,
		Func:     funcNode,
	}, nil
}

type StatRepeat[T any] struct {
	Node
	Location  Location `json:"location"`
	Condition T        `json:"condition"`
	Body      T        `json:"body"`
}

func (n StatRepeat[T]) Type() string {
	return "AstStatRepeat"
}

func (n StatRepeat[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nCondition:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Condition), 4))
	b.WriteString("\nBody:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Body), 4))

	return b.String()
}

func DecodeStatRepeat(data json.RawMessage) (INode, error) {
	var raw StatRepeat[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	condition, err := decodeNode(raw.Condition)
	if err != nil {
		return nil, fmt.Errorf("error decoding condition: %v", err)
	}

	body, err := decodeNode(raw.Body)
	if err != nil {
		return nil, fmt.Errorf("error decoding body: %v", err)
	}

	return StatRepeat[INode]{
		Node:      raw.Node,
		Location:  raw.Location,
		Condition: condition,
		Body:      body,
	}, nil
}

type StatReturn[T any] struct {
	Node
	Location Location `json:"location"`
	List     []T      `json:"list"`
}

func (n StatReturn[T]) Type() string {
	return "AstStatReturn"
}

func (n StatReturn[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nList:")

	for _, item := range n.List {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(item), 4))
	}

	return b.String()
}

func DecodeStatReturn(data json.RawMessage) (INode, error) {
	var raw StatReturn[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	list := make([]INode, len(raw.List))
	for i, item := range raw.List {
		n, err := decodeNode(item)
		if err != nil {
			return nil, fmt.Errorf("error decoding list item: %v", err)
		}
		list[i] = n
	}

	return StatReturn[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		List:     list,
	}, nil
}

type StatTypeAlias[T any] struct {
	Node
	Location     Location `json:"location"`
	Name         string   `json:"name"`
	Generics     []T      `json:"generics"`
	GenericPacks []T      `json:"genericPacks"`
	Value        T        `json:"value"`
	Exported     bool     `json:"exported"`
}

func (n StatTypeAlias[T]) Type() string {
	return "AstStatTypeAlias"
}

func (n StatTypeAlias[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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

func DecodeStatTypeAlias(data json.RawMessage) (INode, error) {
	var raw StatTypeAlias[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	generics := make([]INode, len(raw.Generics))
	for i, g := range raw.Generics {
		n, err := decodeNode(g)
		if err != nil {
			return nil, fmt.Errorf("error decoding generic node: %v", err)
		}
		generics[i] = n
	}

	genericPacks := make([]INode, len(raw.GenericPacks))
	for i, gp := range raw.GenericPacks {
		n, err := decodeNode(gp)
		if err != nil {
			return nil, fmt.Errorf("error decoding generic pack node: %v", err)
		}
		genericPacks[i] = n
	}

	valueNode, err := decodeNode(raw.Value)
	if err != nil {
		return nil, fmt.Errorf("error decoding value: %v", err)
	}

	return StatTypeAlias[INode]{
		Node:         raw.Node,
		Location:     raw.Location,
		Name:         raw.Name,
		Generics:     generics,
		GenericPacks: genericPacks,
		Value:        valueNode,
		Exported:     raw.Exported,
	}, nil
}

type StatWhile[T any] struct {
	Node
	Location  Location `json:"location"`
	Condition T        `json:"condition"`
	Body      T        `json:"body"`
	HasDo     bool     `json:"hasDo"`
}

func (n StatWhile[T]) Type() string {
	return "AstStatWhile"
}

func (n StatWhile[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nCondition:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Condition), 4))
	b.WriteString("\nBody:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Body), 4))
	b.WriteString(fmt.Sprintf("\nHasDo: %t\n", n.HasDo))

	return b.String()
}

func DecodeStatWhile(data json.RawMessage) (INode, error) {
	var raw StatWhile[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	condition, err := decodeNode(raw.Condition)
	if err != nil {
		return nil, fmt.Errorf("error decoding condition: %v", err)
	}

	body, err := decodeNode(raw.Body)
	if err != nil {
		return nil, fmt.Errorf("error decoding body: %v", err)
	}

	return StatWhile[INode]{
		Node:      raw.Node,
		Location:  raw.Location,
		Condition: condition,
		Body:      body,
		HasDo:     raw.HasDo,
	}, nil
}

type TableProp[T any] struct {
	Name string `json:"name"`
	Node
	Location Location `json:"location"`
	PropType T        `json:"propType"`
}

func (n TableProp[T]) Type() string {
	return "AstTableProp"
}

func (n TableProp[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nName: %s", n.Name))
	b.WriteString("\nPropType:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.PropType), 4))

	return b.String()
}

func DecodeTableProp(data json.RawMessage) (INode, error) {
	var raw TableProp[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	propTypeNode, err := decodeNode(raw.PropType)
	if err != nil {
		return nil, fmt.Errorf("error decoding prop type: %v", err)
	}

	return TableProp[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Name:     raw.Name,
		PropType: propTypeNode,
	}, nil
}

type TypeFunction[T any] struct {
	Node
	Location     Location `json:"location"`
	Attributes   []T      `json:"attributes"`
	Generics     []T      `json:"generics"`
	GenericPacks []T      `json:"genericPacks"`
	ArgTypes     T        `json:"argTypes"`
	ArgNames     []T      `json:"argNames"`
	ReturnTypes  T        `json:"returnTypes"`
}

func (n TypeFunction[T]) Type() string {
	return "AstTypeFunction"
}

func (n TypeFunction[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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
		b.WriteString(indentStart(StringMaybeEvaluated(name), 4))
	}
	b.WriteString("\nReturnTypes:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.ReturnTypes), 4))

	return b.String()
}

func DecodeTypeFunction(data json.RawMessage) (INode, error) {
	var raw TypeFunction[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	attributes := make([]INode, len(raw.Attributes))
	for i, attr := range raw.Attributes {
		n, err := decodeNode(attr)
		if err != nil {
			return nil, fmt.Errorf("error decoding attribute node: %v", err)
		}
		attributes[i] = n
	}

	generics := make([]INode, len(raw.Generics))
	for i, gen := range raw.Generics {
		n, err := decodeNode(gen)
		if err != nil {
			return nil, fmt.Errorf("error decoding generic node: %v", err)
		}
		generics[i] = n
	}

	genericPacks := make([]INode, len(raw.GenericPacks))
	for i, pack := range raw.GenericPacks {
		n, err := decodeNode(pack)
		if err != nil {
			return nil, fmt.Errorf("error decoding generic pack node: %v", err)
		}
		genericPacks[i] = n
	}

	argTypesNode, err := decodeNode(raw.ArgTypes)
	if err != nil {
		return nil, fmt.Errorf("error decoding arg types: %v", err)
	}

	argNames := make([]INode, len(raw.ArgNames))
	for i, name := range raw.ArgNames {
		n, err := decodeNode(name)
		if err != nil {
			return nil, fmt.Errorf("error decoding arg name node: %v", err)
		}
		argNames[i] = n
	}

	returnTypesNode, err := decodeNode(raw.ReturnTypes)
	if err != nil {
		return nil, fmt.Errorf("error decoding return types: %v", err)
	}

	return TypeFunction[INode]{
		Node:         raw.Node,
		Location:     raw.Location,
		Attributes:   attributes,
		Generics:     generics,
		GenericPacks: genericPacks,
		ArgTypes:     argTypesNode,
		ArgNames:     argNames,
		ReturnTypes:  returnTypesNode,
	}, nil
}

type TypeList[T any] struct {
	Node
	Types []T `json:"types"`
}

func (n TypeList[T]) Type() string {
	return "AstTypeList"
}

func (n TypeList[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString("Types:")

	for _, typ := range n.Types {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(typ), 4))
	}

	return b.String()
}

func DecodeTypeList(data json.RawMessage) (INode, error) {
	var raw TypeList[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	types := make([]INode, len(raw.Types))
	for i, typ := range raw.Types {
		n, err := decodeNode(typ)
		if err != nil {
			return nil, fmt.Errorf("error decoding type node: %v", err)
		}
		types[i] = n
	}

	return TypeList[INode]{
		Node:  raw.Node,
		Types: types,
	}, nil
}

type TypePackExplicit[T any] struct {
	Node
	Location Location `json:"location"`
	TypeList T        `json:"typeList"`
}

func (n TypePackExplicit[T]) Type() string {
	return "AstTypePackExplicit"
}

func (n TypePackExplicit[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nTypeList:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.TypeList), 4))

	return b.String()
}

func DecodeTypePackExplicit(data json.RawMessage) (INode, error) {
	var raw TypePackExplicit[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	typeListNode, err := decodeNode(raw.TypeList)
	if err != nil {
		return nil, fmt.Errorf("error decoding type list: %v", err)
	}

	return TypePackExplicit[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		TypeList: typeListNode,
	}, nil
}

type TypeReference[T any] struct {
	Node
	Location     Location `json:"location"`
	Name         string   `json:"name"`
	NameLocation Location `json:"nameLocation"`
	Parameters   []T      `json:"parameters"`
}

func (n TypeReference[T]) Type() string {
	return "AstTypeReference"
}

func (n TypeReference[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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

func DecodeTypeReference(data json.RawMessage) (INode, error) {
	var raw TypeReference[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	parameters := make([]INode, len(raw.Parameters))
	for i, param := range raw.Parameters {
		n, err := decodeNode(param)
		if err != nil {
			return nil, fmt.Errorf("error decoding parameter node: %v", err)
		}
		parameters[i] = n
	}

	return TypeReference[INode]{
		Node:         raw.Node,
		Location:     raw.Location,
		Name:         raw.Name,
		NameLocation: raw.NameLocation,
		Parameters:   parameters,
	}, nil
}

type Indexer[T any] struct {
	Location   Location `json:"location"`
	IndexType  T        `json:"indexType"`
	ResultType T        `json:"resultType"`
}

func (n Indexer[T]) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nIndexType:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.IndexType), 4))
	b.WriteString("\nResultType:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.ResultType), 4))

	return b.String()
}

type TypeTable[T any] struct {
	Node
	Location Location    `json:"location"`
	Props    []T         `json:"props"`
	Indexer  *Indexer[T] `json:"indexer"`
}

func (n TypeTable[T]) Type() string {
	return "AstTypeTable"
}

func (n TypeTable[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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

func DecodeTypeTable(data json.RawMessage) (INode, error) {
	var raw TypeTable[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	props := make([]INode, len(raw.Props))
	for i, prop := range raw.Props {
		n, err := decodeNode(prop)
		if err != nil {
			return nil, fmt.Errorf("error decoding prop node: %v", err)
		}
		props[i] = n
	}

	var indexerMaybe *Indexer[INode]
	if raw.Indexer != nil {
		indexerNode, err := decodeNode(raw.Indexer.IndexType)
		if err != nil {
			return nil, fmt.Errorf("error decoding indexer index type: %v", err)
		}
		resultNode, err := decodeNode(raw.Indexer.ResultType)
		if err != nil {
			return nil, fmt.Errorf("error decoding indexer result type: %v", err)
		}
		indexerMaybe = &Indexer[INode]{
			Location:   raw.Indexer.Location,
			IndexType:  indexerNode,
			ResultType: resultNode,
		}
	}

	return TypeTable[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Props:    props,
		Indexer:  indexerMaybe,
	}, nil
}

type TypeUnion[T any] struct {
	Node
	Location Location `json:"location"`
	Types    []T      `json:"types"`
}

func (n TypeUnion[T]) Type() string {
	return "AstTypeUnion"
}

func (n TypeUnion[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("Types:")

	for _, typ := range n.Types {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(typ), 4))
	}

	return b.String()
}

func DecodeTypeUnion(data json.RawMessage) (INode, error) {
	var raw TypeUnion[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}

	types := make([]INode, len(raw.Types))
	for i, typ := range raw.Types {
		n, err := decodeNode(typ)
		if err != nil {
			return nil, fmt.Errorf("error decoding type node: %v", err)
		}
		types[i] = n
	}

	return TypeUnion[INode]{
		Node:     raw.Node,
		Location: raw.Location,
		Types:    types,
	}, nil
}

// decoding

func decodeNode(data json.RawMessage) (INode, error) {
	var node Node
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("error decoding node: %v", err)
	}

	// helper for now
	ret := func(n INode, err error) (INode, error) {
		if err != nil {
			return nil, fmt.Errorf("\n%-22s %v", node.Type, err)
		}
		return n, nil
	}

	switch t := node.Type; t {
	case "AstArgumentName":
		return ret(DecodeArgumentName(data))
	case "AstAttr":
		return ret(DecodeAttr(data))
	case "AstExprBinary":
		return ret(DecodeExprBinary(data))
	case "AstExprCall":
		return ret(DecodeExprCall(data))
	case "AstExprConstantBool":
		return ret(DecodeExprConstantBool(data))
	case "AstExprConstantNil":
		return ret(DecodeExprConstantNil(data))
	case "AstExprConstantNumber":
		return ret(DecodeExprConstantNumber(data))
	case "AstExprConstantString":
		return ret(DecodeExprConstantString(data))
	case "AstExprFunction":
		return ret(DecodeExprFunction(data))
	case "AstExprGlobal":
		return ret(DecodeExprGlobal(data))
	case "AstExprGroup":
		return ret(DecodeExprGroup(data))
	case "AstExprIfElse":
		return ret(DecodeExprIfElse(data))
	case "AstExprIndexExpr":
		return ret(DecodeExprIndexExpr(data))
	case "AstExprIndexName":
		return ret(DecodeExprIndexName(data))
	case "AstExprInterpString":
		return ret(DecodeExprInterpString(data))
	case "AstExprLocal":
		return ret(DecodeExprLocal(data))
	case "AstExprTable":
		return ret(DecodeExprTable(data))
	case "AstExprTableItem":
		return ret(DecodeExprTableItem(data))
	case "AstExprTypeAssertion":
		return ret(DecodeExprTypeAssertion(data))
	case "AstExprVarargs":
		return ret(DecodeExprVarargs(data))
	case "AstExprUnary":
		return ret(DecodeExprUnary(data))
	case "AstGenericType":
		return ret(DecodeGenericType(data))
	case "AstLocal":
		return ret(DecodeLocal(data))
	case "AstStatAssign":
		return ret(DecodeStatAssign(data))
	case "AstStatBlock":
		return ret(DecodeStatBlock(data))
	case "AstStatBreak":
		return ret(DecodeStatBreak(data))
	case "AstStatCompoundAssign":
		return ret(DecodeStatCompoundAssign(data))
	case "AstStatContinue":
		return ret(DecodeStatContinue(data))
	case "AstStatExpr":
		return ret(DecodeStatExpr(data))
	case "AstStatFor":
		return ret(DecodeStatFor(data))
	case "AstStatForIn":
		return ret(DecodeStatForIn(data))
	case "AstStatFunction":
		return ret(DecodeStatFunction(data))
	case "AstStatIf":
		return ret(DecodeStatIf(data))
	case "AstStatLocal":
		return ret(DecodeStatLocal(data))
	case "AstStatLocalFunction":
		return ret(DecodeStatLocalFunction(data))
	case "AstStatRepeat":
		return ret(DecodeStatRepeat(data))
	case "AstStatReturn":
		return ret(DecodeStatReturn(data))
	case "AstStatTypeAlias":
		return ret(DecodeStatTypeAlias(data))
	case "AstStatWhile":
		return ret(DecodeStatWhile(data))
	case "AstTableProp":
		return ret(DecodeTableProp(data))
	case "AstTypeFunction":
		return ret(DecodeTypeFunction(data))
	case "AstTypeList":
		return ret(DecodeTypeList(data))
	case "AstTypePackExplicit":
		return ret(DecodeTypePackExplicit(data))
	case "AstTypeReference":
		return ret(DecodeTypeReference(data))
	case "AstTypeTable":
		return ret(DecodeTypeTable(data))
	case "AstTypeUnion":
		return ret(DecodeTypeUnion(data))
	}
	return ret(nil, errors.New("unknown node type"))
}
