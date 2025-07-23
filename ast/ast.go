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
	Type     string   `json:"type"`
	Location Location `json:"location"`
}

func (n Node) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Type      %s\n", n.Type))
	b.WriteString(fmt.Sprintf("Location  %s\n", n.Location))
	return b.String()
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
}

func (c Comment) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Type      %s\n", c.Type))
	b.WriteString(fmt.Sprintf("Location  %s\n", c.Location))

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

	b.WriteString("Comment Locations:\n")
	for _, c := range ast.CommentLocations {
		b.WriteString(indentStart(c.String(), 4))
		b.WriteString("\n\n")
	}

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

type StatBlock[T any] struct {
	Node
	HasEnd bool `json:"hasEnd"`
	Body   []T  `json:"body"`
}

func (n StatBlock[T]) Type() string {
	return "AstStatBlock"
}

func (n StatBlock[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("HasEnd    %t\n", n.HasEnd))
	b.WriteString("Body:\n")

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
		Node:   raw.Node,
		HasEnd: raw.HasEnd,
		Body:   body,
	}, nil
}

type StatExpr[T any] struct {
	Node
	Expr T `json:"expr"`
}

func (n StatExpr[T]) Type() string {
	return "AstStatExpr"
}

func (n StatExpr[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString("Expr:\n")
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
		Node: raw.Node,
		Expr: n,
	}, nil
}

type StatWhile[T any] struct {
	Node
	Condition T    `json:"condition"`
	Body      T    `json:"body"`
	HasDo     bool `json:"hasDo"`
}

func (n StatWhile[T]) Type() string {
	return "AstStatWhile"
}

func (n StatWhile[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString("Condition:\n")
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
		Condition: condition,
		Body:      body,
		HasDo:     raw.HasDo,
	}, nil
}

type ExprCall[T any] struct {
	Node
	Func T   `json:"func"`
	Args []T `json:"args"`
}

func (n ExprCall[T]) Type() string {
	return "AstExprCall"
}

func (n ExprCall[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString("Func:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Func), 4))
	b.WriteString("\nArgs:\n")

	for _, arg := range n.Args {
		b.WriteString(indentStart(StringMaybeEvaluated(arg), 4))
		b.WriteString("\n")
	}

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
		Node: raw.Node,
		Func: funcNode,
		Args: args,
	}, nil
}

type ExprGlobal struct {
	Node
	Global string `json:"global"`
}

func (n ExprGlobal) Type() string {
	return "AstExprGlobal"
}

func (n ExprGlobal) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Global    %s\n", n.Global))

	return b.String()
}

func DecodeExprGlobal(data json.RawMessage) (INode, error) {
	var raw ExprGlobal
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type ExprConstantString struct {
	Node
	Value string `json:"value"`
}

func (n ExprConstantString) Type() string {
	return "AstExprConstantString"
}

func (n ExprConstantString) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Value     %s\n", n.Value))

	return b.String()
}

func DecodeExprConstantString(data json.RawMessage) (INode, error) {
	var raw ExprConstantString
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type ExprConstantNumber struct {
	Node
	Value float64 `json:"value"`
}

func (n ExprConstantNumber) Type() string {
	return "AstExprConstantNumber"
}

func (n ExprConstantNumber) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Value     %f\n", n.Value))

	return b.String()
}

func DecodeExprConstantNumber(data json.RawMessage) (INode, error) {
	var raw ExprConstantNumber
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
}

type ExprConstantBool struct {
	Node
	Value bool `json:"value"`
}

func (n ExprConstantBool) Type() string {
	return "AstExprConstantBool"
}

func (n ExprConstantBool) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Value     %t\n", n.Value))

	return b.String()
}

func DecodeExprConstantBool(data json.RawMessage) (INode, error) {
	var raw ExprConstantBool
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %v", err)
	}
	return raw, nil
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
	case "AstStatBlock":
		return ret(DecodeStatBlock(data))
	case "AstStatExpr":
		return ret(DecodeStatExpr(data))
	case "AstStatWhile":
		return ret(DecodeStatWhile(data))
	case "AstExprCall":
		return ret(DecodeExprCall(data))
	case "AstExprGlobal":
		return ret(DecodeExprGlobal(data))
	case "AstExprConstantString":
		return ret(DecodeExprConstantString(data))
	case "AstExprConstantNumber":
		return ret(DecodeExprConstantNumber(data))
	case "AstExprConstantBool":
		return ret(DecodeExprConstantBool(data))
	}
	return ret(nil, errors.New("unknown node type"))
}
