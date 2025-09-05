package ast

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"slices"
	"strings"
)

const (
	Ext            = ".luau"
	AstDir         = "../test/ast"
	BenchmarkDir   = "../test/benchmark"
	ConformanceDir = "../test/conformance"
)

func LuauAst(path string) (out []byte, err error) {
	cmd := exec.Command("luau-ast", path)
	out, err = cmd.Output()
	if err != nil {
		return
	}

	strout := string(out)
	strout = strings.ReplaceAll(strout, `"value":Infinity`, `"value":"Infinity"`)

	return []byte(strout), nil
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

// dot net vibez
type INode interface {
	GetLocation() Location
	String() string
	Source(og string, indent int) (string, error)
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
	Start Position `json:"start"`
	End   Position `json:"end"`
}

func (l Location) Contains(ol Location) bool {
	return ol.Start.After(l.Start) && l.End.After(ol.End)
}

// infinite gfs (or, well, 9223372036854775807)
// var gfsCount int

// we've pretty much eliminated calls to this function lel, it'll always be needed for comments though
func (l Location) GetFromSource(source string) (string, error) {
	lines := strings.Split(source, "\n")
	if l.Start.Line < 0 || l.End.Line >= len(lines) {
		return "", errors.New("location out of bounds")
	}

	var b strings.Builder
	for i := l.Start.Line; i <= l.End.Line; i++ {
		line := lines[i]
		if i == l.Start.Line && i == l.End.Line {
			line = line[l.Start.Column:l.End.Column]
		} else if i == l.Start.Line {
			line = line[l.Start.Column:]
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

type NodeLoc struct {
	Node
	Location Location `json:"location"`
}

func (nl NodeLoc) GetLocation() Location {
	return nl.Location
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
	NodeLoc
}

func (c Comment) String() string {
	var b strings.Builder

	b.WriteString(c.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s\n", c.Location))

	return b.String()
}

func (c Comment) Source(og string) (string, error) {
	return c.Location.GetFromSource(og)
}

// statblocks are the most important node for comments
type StatBlockDepth struct {
	StatBlock[INode]
	Depth int
}

type AST[T any] struct {
	Root             StatBlock[T] `json:"root"`
	CommentLocations []Comment    `json:"commentLocations"`
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

func (ast AST[T]) Source(og string) (string, error) {
	iroot, ok := any(ast.Root).(StatBlock[INode])
	if !ok {
		return "", fmt.Errorf("expected Root to be StatBlock[INode], got %T", ast)
	}

	rs, err := iroot.Source(og, 0) // ewh, rs
	if err != nil {
		return "", fmt.Errorf("error getting root source: %w", err)
	}

	return rs + "\n", nil
}

type AddStatBlock func(StatBlock[INode], int)

func DecodeAST(data json.RawMessage) (AST[INode], error) {
	var raw AST[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return AST[INode]{}, fmt.Errorf("error unmarshalling AST: %w", err)
	}

	var statBlocks []StatBlockDepth
	addStatBlock := func(sb StatBlock[INode], depth int) {
		statBlocks = append(statBlocks, StatBlockDepth{StatBlock: sb, Depth: depth})
	}

	rootNode, err := DecodeStatBlockKnown(raw.Root, addStatBlock, 0)
	if err != nil {
		return AST[INode]{}, fmt.Errorf("error decoding root node: %w", err)
	}

	slices.SortFunc(statBlocks, func(a, b StatBlockDepth) int {
		return b.Depth - a.Depth
	})

	// for each comment, add it to the deepest statblock that fully contains it
	for _, comment := range raw.CommentLocations {
		for _, sb := range statBlocks {
			if !sb.Location.Contains(comment.Location) {
				continue
			}
			*sb.CommentsContained = append(*sb.CommentsContained, comment)
			break
		}
	}

	return AST[INode]{
		Root:             rootNode,
		CommentLocations: raw.CommentLocations,
	}, nil
}

// node types (ok, real ast now)

type ArgumentName struct {
	Node
	Name     string   `json:"name"`
	Location Location `json:"location"`
}

func (a ArgumentName) GetLocation() Location {
	return a.Location
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

func (a ArgumentName) Source(string, int) (string, error) {
	return a.Name, nil
}

func DecodeArgumentName(data json.RawMessage) (INode, error) {
	var raw ArgumentName
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

type Attr struct {
	NodeLoc
	Name string `json:"name"`
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

func (a Attr) Source(string, int) (string, error) {
	// return a.Location.GetFromSource(og)
	return fmt.Sprintf("@%s", a.Name), nil // they're called @ributes for a reason
}

func DecodeAttr(data json.RawMessage) (INode, error) {
	var raw Attr
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

type DeclaredClassProp[T any] struct {
	Name         string   `json:"name"`
	NameLocation Location `json:"nameLocation"`
	Node
	LuauType T        `json:"luauType"`
	Location Location `json:"location"`
}

func (d DeclaredClassProp[T]) GetLocation() Location {
	return d.Location
}

func (d DeclaredClassProp[T]) Type() string {
	return "AstDeclaredClassProp"
}

func (d DeclaredClassProp[T]) String() string {
	var b strings.Builder

	b.WriteString(d.Node.String())
	b.WriteString(fmt.Sprintf("Name: %s", d.Name))
	b.WriteString(fmt.Sprintf("\nNameLocation: %s", d.NameLocation))
	b.WriteString(fmt.Sprintf("\nLocation: %s", d.Location))
	b.WriteString("\nLuauType:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(d.LuauType), 4))

	return b.String()
}

func (d DeclaredClassProp[T]) Source(og string, indent int) (string, error) {
	ilt, ok := any(d.LuauType).(INode)
	if !ok {
		return "", fmt.Errorf("expected LuauType to be INode, got %T", d.LuauType)
	}

	// TODO: we have no way of knowing whether the method has a self parameter {;-;}
	lts, err := ilt.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting LuauType source: %w", err)
	}

	return IndentSize(indent) + fmt.Sprintf("%s: %s", d.Name, lts), nil
}

func DecodeDeclaredClassProp(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw DeclaredClassProp[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	luauTypeNode, err := decodeNode(raw.LuauType, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding luauType: %w", err)
	}

	return DeclaredClassProp[INode]{
		Name:         raw.Name,
		NameLocation: raw.NameLocation,
		Node:         raw.Node,
		LuauType:     luauTypeNode,
		Location:     raw.Location,
	}, nil
}

type ExprBinary[T any] struct {
	NodeLoc
	Op    string `json:"op"`
	Left  T      `json:"left"`
	Right T      `json:"right"`
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

func (n ExprBinary[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(ExprBinary[INode])
	if !ok {
		return "", fmt.Errorf("expected ExprBinary[INode], got %T", n)
	}

	l, err := in.Left.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting left source: %w", err)
	}

	r, err := in.Right.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting right source: %w", err)
	}

	op := BinopToSource(in.Op)

	return fmt.Sprintf("%s %s %s", l, op, r), nil
}

func DecodeExprBinary(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprBinary[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	left, err := decodeNode(raw.Left, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding left: %w", err)
	}

	right, err := decodeNode(raw.Right, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding right: %w", err)
	}

	return ExprBinary[INode]{
		NodeLoc: raw.NodeLoc,
		Op:      raw.Op,
		Left:    left,
		Right:   right,
	}, nil
}

type ExprCall[T any] struct {
	NodeLoc
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

func (n ExprCall[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(ExprCall[INode])
	if !ok {
		return "", fmt.Errorf("expected ExprCall[INode], got %T", n)
	}

	ns, err := in.Func.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting function source: %w", err)
	}

	// if len(in.Args) == 0 {
	// 	return fmt.Sprintf("%s()", ns), nil
	// }

	// if len(in.Args) == 1 {
	// 	if str, ok := in.Args[0].(ExprConstantString); ok {
	// 		strs, err := str.Source(og, indent)
	// 		if err != nil {
	// 			return "", fmt.Errorf("error getting string source: %w", err)
	// 		}
	// 		return fmt.Sprintf("%s %s", ns, strs), nil
	// 	}
	// }

	argStrings := make([]string, len(in.Args))
	for i, arg := range in.Args {
		ns, err := arg.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting argument source: %w", err)
		}
		argStrings[i] = ns
	}

	// fmt.Println(ns, argStrings)
	return fmt.Sprintf("%s(%s)", ns, strings.Join(argStrings, ", ")), nil
}

func DecodeExprCall(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprCall[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	funcNode, err := decodeNode(raw.Func, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding func: %w", err)
	}

	args := make([]INode, len(raw.Args))
	for i, arg := range raw.Args {
		n, err := decodeNode(arg, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding arg node: %w", err)
		}
		args[i] = n
	}

	return ExprCall[INode]{
		NodeLoc:     raw.NodeLoc,
		Func:        funcNode,
		Args:        args,
		Self:        raw.Self,
		ArgLocation: raw.ArgLocation,
	}, nil
}

type ExprConstantBool struct {
	NodeLoc
	Value bool `json:"value"`
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

func (n ExprConstantBool) Source(string, int) (string, error) {
	if n.Value {
		return "true", nil
	}
	return "false", nil
}

func DecodeExprConstantBool(data json.RawMessage) (INode, error) {
	var raw ExprConstantBool
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

type ExprConstantNil struct {
	NodeLoc
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

func (n ExprConstantNil) Source(string, int) (string, error) {
	return "nil", nil // nil
}

func DecodeExprConstantNil(data json.RawMessage) (INode, error) {
	var raw ExprConstantNil
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

type ExprConstantNumber struct {
	NodeLoc
	Value Number `json:"value"`
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

func (n ExprConstantNumber) Source(og string, _ int) (string, error) {
	return NumberToSource(n.Value), nil
}

func DecodeExprConstantNumber(data json.RawMessage) (INode, error) {
	var raw ExprConstantNumber
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

type ExprConstantString struct {
	NodeLoc
	Value string `json:"value"`
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

func (n ExprConstantString) Source(og string, _ int) (string, error) {
	// return n.Location.GetFromSource(og)
	return StringToSource(n.Value), nil
}

func DecodeExprConstantString(data json.RawMessage) (INode, error) {
	var raw ExprConstantString
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

type ExprFunction[T any] struct {
	NodeLoc
	Attributes     []Attr            `json:"attributes"`
	Generics       []GenericType     `json:"generics"`
	GenericPacks   []GenericTypePack `json:"genericPacks"`
	Args           []T               `json:"args"`
	Vararg         bool              `json:"vararg"`
	VarargLocation Location          `json:"varargLocation"`
	Body           StatBlock[T]      `json:"body"`
	FunctionDepth  int               `json:"functionDepth"`
	Debugname      string            `json:"debugname"`
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

func (n ExprFunction[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(ExprFunction[INode])
	if !ok {
		return "", fmt.Errorf("expected ExprFunction[INode], got %T", n)
	}
	// return in.Location.GetFromSource(og)
	l := len(in.Args)
	if in.Vararg {
		l++
	}

	argStrings := make([]string, l)
	for i, arg := range in.Args {
		iarg, ok := arg.(Local[INode])
		if !ok {
			return "", fmt.Errorf("expected Local, got %T", arg)
		}

		as, err := iarg.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting source for arg %d: %w", i, err)
		}

		if iarg.LuauType == nil {
			argStrings[i] = as
			continue
		}

		lts, err := (*iarg.LuauType).Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting source for arg %d LuauType: %w", i, err)
		}

		argStrings[i] = fmt.Sprintf("%s: %s", as, lts)
	}

	if in.Vararg {
		argStrings[l-1] = "..."
	}

	bs, err := in.Body.Source(og, indent+1)
	if err != nil {
		return "", fmt.Errorf("error getting body source: %w", err)
	}

	var b strings.Builder
	b.WriteString("function")
	if in.Debugname != "" {
		b.WriteByte(' ')
		b.WriteString(in.Debugname)
	}

	if len(in.Generics) > 0 || len(in.GenericPacks) > 0 {
		genericStrings := make([]string, len(in.Generics))
		for i, g := range in.Generics {
			genericStrings[i] = g.Name
		}

		genericPackStrings := make([]string, len(in.GenericPacks))
		for i, gp := range in.GenericPacks {
			genericPackStrings[i] = gp.Name + "..."
		}

		allGenerics := append(genericStrings, genericPackStrings...)

		b.WriteString(fmt.Sprintf("<%s>", strings.Join(allGenerics, ", ")))
	}

	b.WriteString(fmt.Sprintf("(%s)\n%s\n", strings.Join(argStrings, ", "), bs))
	b.WriteString(IndentSize(indent) + "end")
	return b.String(), nil
}

func DecodeExprFunctionKnown(raw ExprFunction[json.RawMessage], addStatBlock AddStatBlock, depth int) (ExprFunction[INode], error) {
	args := make([]INode, len(raw.Args))
	for i, arg := range raw.Args {
		n, err := decodeNode(arg, addStatBlock, depth+1)
		if err != nil {
			return ExprFunction[INode]{}, fmt.Errorf("error decoding arg node: %w", err)
		}
		args[i] = n
	}

	bodyNode, err := DecodeStatBlockKnown(raw.Body, addStatBlock, depth+1)
	if err != nil {
		return ExprFunction[INode]{}, fmt.Errorf("error decoding body node: %w", err)
	}

	return ExprFunction[INode]{
		NodeLoc:        raw.NodeLoc,
		Attributes:     raw.Attributes,
		Generics:       raw.Generics,
		GenericPacks:   raw.GenericPacks,
		Args:           args,
		Vararg:         raw.Vararg,
		VarargLocation: raw.VarargLocation,
		Body:           bodyNode,
		FunctionDepth:  raw.FunctionDepth,
		Debugname:      raw.Debugname,
	}, nil
}

func DecodeExprFunction(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprFunction[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	return DecodeExprFunctionKnown(raw, addStatBlock, depth+1)
}

type ExprGlobal struct {
	NodeLoc
	Global string `json:"global"`
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

func (n ExprGlobal) Source(string, int) (string, error) {
	// return n.Location.GetFromSource(og)
	return n.Global, nil
}

func DecodeExprGlobal(data json.RawMessage) (INode, error) {
	var raw ExprGlobal
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

type ExprGroup[T any] struct {
	NodeLoc
	Expr T `json:"expr"` // only contains one expression right? strange when you first think about it
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

func (n ExprGroup[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	iexpr, ok := any(n.Expr).(INode)
	if !ok {
		return "", fmt.Errorf("expected Expr to be INode, got %T", n.Expr)
	}

	sexpr, err := iexpr.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting source for expr: %w", err)
	}

	return fmt.Sprintf("(%s)", sexpr), nil
}

func DecodeExprGroup(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprGroup[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	exprNode, err := decodeNode(raw.Expr, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %w", err)
	}

	return ExprGroup[INode]{
		NodeLoc: raw.NodeLoc,
		Expr:    exprNode,
	}, nil
}

type ExprIfElse[T any] struct {
	NodeLoc
	Condition T    `json:"condition"`
	HasThen   bool `json:"hasThen"`
	TrueExpr  T    `json:"trueExpr"`
	HasElse   bool `json:"hasElse"`
	FalseExpr T    `json:"falseExpr"`
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

func (n ExprIfElse[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(ExprIfElse[INode])
	if !ok {
		return "", fmt.Errorf("expected ExprIfElse[INode], got %T", n)
	}

	scond, err := in.Condition.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting condition source: %w", err)
	}

	strue, err := in.TrueExpr.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting true expression source: %w", err)
	}

	sfalse, err := in.FalseExpr.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting false expression source: %w", err)
	}

	if in.FalseExpr.Type() == "AstExprIfElse" {
		return fmt.Sprintf("if %s then %s else%s", scond, strue, sfalse), nil
	}
	return fmt.Sprintf("if %s then %s else %s", scond, strue, sfalse), nil
}

func DecodeExprIfElse(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprIfElse[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	conditionNode, err := decodeNode(raw.Condition, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding condition: %w", err)
	}

	trueExprNode, err := decodeNode(raw.TrueExpr, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding true expression: %w", err)
	}

	falseExprNode, err := decodeNode(raw.FalseExpr, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding false expression: %w", err)
	}

	return ExprIfElse[INode]{
		NodeLoc:   raw.NodeLoc,
		Condition: conditionNode,
		HasThen:   raw.HasThen,
		TrueExpr:  trueExprNode,
		HasElse:   raw.HasElse,
		FalseExpr: falseExprNode,
	}, nil
}

type ExprIndexExpr[T any] struct {
	NodeLoc
	Expr  T `json:"expr"`
	Index T `json:"index"`
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

func (n ExprIndexExpr[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(ExprIndexExpr[INode])
	if !ok {
		return "", fmt.Errorf("expected ExprIndexExpr[INode], got %T", n)
	}

	exprSource, err := in.Expr.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting expr source: %w", err)
	}

	indexSource, err := in.Index.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting index source: %w", err)
	}

	return fmt.Sprintf("%s[%s]", exprSource, indexSource), nil
}

func DecodeExprIndexExpr(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprIndexExpr[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	exprNode, err := decodeNode(raw.Expr, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %w", err)
	}

	indexNode, err := decodeNode(raw.Index, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding index: %w", err)
	}

	return ExprIndexExpr[INode]{
		NodeLoc: raw.NodeLoc,
		Expr:    exprNode,
		Index:   indexNode,
	}, nil
}

type ExprIndexName[T any] struct {
	NodeLoc
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

func (n ExprIndexName[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(ExprIndexName[INode])
	if !ok {
		return "", fmt.Errorf("expected ExprIndexName[INode], got %T", n)
	}

	es, err := in.Expr.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting expr source: %w", err)
	}

	return fmt.Sprintf("%s.%s", es, in.Index), nil
}

func DecodeExprIndexName(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprIndexName[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	exprNode, err := decodeNode(raw.Expr, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %w", err)
	}

	return ExprIndexName[INode]{
		NodeLoc:       raw.NodeLoc,
		Expr:          exprNode,
		Index:         raw.Index,
		IndexLocation: raw.IndexLocation,
		Op:            raw.Op,
	}, nil
}

type ExprInterpString[T any] struct {
	NodeLoc
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

func (n ExprInterpString[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	iexprs, ok := any(n.Expressions).([]INode)
	if !ok {
		return "", fmt.Errorf("expected Expressions to be []INode, got %T", n.Expressions)
	}

	ls := len(n.Strings)
	if len(iexprs) != ls-1 {
		return "", fmt.Errorf("mismatched string and expression counts: %d vs %d", len(n.Strings), len(iexprs))
	}

	parts := n.Strings
	for i, expr := range iexprs {
		ies, err := expr.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting expr source: %w", err)
		}
		parts[i] += fmt.Sprintf("{%s}", ies)
	}

	return fmt.Sprintf("`%s`", strings.Join(parts, "")), nil
}

func DecodeExprInterpString(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprInterpString[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	expressions := make([]INode, len(raw.Expressions))
	for i, expr := range raw.Expressions {
		n, err := decodeNode(expr, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding expression node: %w", err)
		}
		expressions[i] = n
	}

	return ExprInterpString[INode]{
		NodeLoc:     raw.NodeLoc,
		Strings:     raw.Strings,
		Expressions: expressions,
	}, nil
}

type ExprLocal[T any] struct {
	NodeLoc
	Local T `json:"local"`
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

func (n ExprLocal[T]) Source(og string, indent int) (string, error) {
	ilocal, ok := any(n.Local).(Local[INode])
	if !ok {
		return "", fmt.Errorf("expected Local to be Local[INode], got %T", n.Local)
	}
	return ilocal.Source(og, indent)
}

func DecodeExprLocal(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprLocal[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	localNode, err := decodeNode(raw.Local, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding local: %w", err)
	}

	return ExprLocal[INode]{
		NodeLoc: raw.NodeLoc,
		Local:   localNode,
	}, nil
}

type ExprTable[T any] struct {
	NodeLoc
	Items []T `json:"items"`
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

func (n ExprTable[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	iitems, ok := any(n.Items).([]INode)
	if !ok {
		return "", fmt.Errorf("expected Items to be []INode, got %T", n.Items)
	}

	if len(iitems) == 0 {
		return "{}", nil
	}

	itemStrings := make([]string, len(iitems))
	for i, item := range iitems {
		is, err := item.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting source for item %d: %w", i, err)
		}
		itemStrings[i] = is
	}

	return fmt.Sprintf("{ %s }", strings.Join(itemStrings, ", ")), nil
}

func DecodeExprTable(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprTable[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	items := make([]INode, len(raw.Items))
	for i, item := range raw.Items {
		n, err := decodeNode(item, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding item node: %w", err)
		}
		items[i] = n
	}

	return ExprTable[INode]{
		NodeLoc: raw.NodeLoc,
		Items:   items,
	}, nil
}

type ExprTableItem[T any] struct {
	Node
	Kind  string `json:"kind"`
	Key   *T     `json:"key"`
	Value T      `json:"value"`
}

func (n ExprTableItem[T]) GetLocation() Location {
	return Location{}
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

func (n ExprTableItem[T]) Source(og string, indent int) (string, error) {
	// TableItem doesn't seem to have a Location field, using Value's location if possible
	// return "", errors.New("table item has no direct location")
	in, ok := any(n).(ExprTableItem[INode])
	if !ok {
		return "", fmt.Errorf("expected ExprTableItem[INode], got %T", n)
	}

	vs, err := in.Value.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting value source: %w", err)
	}

	if in.Key == nil {
		return vs, nil
	}

	ks, err := (*in.Key).Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting key source: %w", err)
	}

	switch in.Kind {
	case "record":
		return fmt.Sprintf("%s = %s", ks, vs), nil
	case "general":
		return fmt.Sprintf("[%s] = %s", ks, vs), nil
	}

	return "", fmt.Errorf("unknown table item kind: %s", in.Kind)
}

func DecodeExprTableItem(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprTableItem[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	var keyNodeMaybe *INode
	if raw.Key != nil {
		keyNode, err := decodeNode(*raw.Key, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding key: %w", err)
		}
		keyNodeMaybe = &keyNode
	}

	valueNode, err := decodeNode(raw.Value, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding value: %w", err)
	}

	return ExprTableItem[INode]{
		Node:  raw.Node,
		Kind:  raw.Kind,
		Key:   keyNodeMaybe,
		Value: valueNode,
	}, nil
}

type ExprTypeAssertion[T any] struct {
	NodeLoc
	Expr       T `json:"expr"`
	Annotation T `json:"annotation"`
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

func (n ExprTypeAssertion[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(ExprTypeAssertion[INode])
	if !ok {
		return "", fmt.Errorf("expected ExprTypeAssertion[INode], got %T", n)
	}

	sexpr, err := in.Expr.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting expr source: %w", err)
	}

	sannotation, err := in.Annotation.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting annotation source: %w", err)
	}

	return fmt.Sprintf("%s :: %s", sexpr, sannotation), nil
}

func DecodeExprTypeAssertion(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprTypeAssertion[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	exprNode, err := decodeNode(raw.Expr, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %w", err)
	}

	annotationNode, err := decodeNode(raw.Annotation, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding annotation: %w", err)
	}

	return ExprTypeAssertion[INode]{
		NodeLoc:    raw.NodeLoc,
		Expr:       exprNode,
		Annotation: annotationNode,
	}, nil
}

type ExprVarargs struct {
	NodeLoc
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

func (n ExprVarargs) Source(string, int) (string, error) {
	// return n.Location.GetFromSource(og)
	return "...", nil
}

func DecodeExprVarargs(data json.RawMessage) (INode, error) {
	var raw ExprVarargs
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

type ExprUnary[T any] struct {
	NodeLoc
	Op   string `json:"op"`
	Expr T      `json:"expr"`
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

func (n ExprUnary[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(ExprUnary[INode])
	if !ok {
		return "", fmt.Errorf("expected ExprUnary[INode], got %T", n)
	}

	exprSource, err := in.Expr.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting expr source: %w", err)
	}

	op := UnopToSource(in.Op)

	return op + exprSource, nil
}

func DecodeExprUnary(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw ExprUnary[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	exprNode, err := decodeNode(raw.Expr, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %w", err)
	}

	return ExprUnary[INode]{
		NodeLoc: raw.NodeLoc,
		Op:      raw.Op,
		Expr:    exprNode,
	}, nil
}

type GenericType struct {
	Node
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

	b.WriteString(g.Node.String())
	b.WriteString(fmt.Sprintf("Name: %s", g.Name))

	return b.String()
}

func DecodeGenericType(data json.RawMessage) (INode, error) {
	var raw GenericType
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

func (g GenericType) Source(string, int) (string, error) {
	// GenericType doesn't have a Location field
	return "", errors.New("generic type has no location")
}

type GenericTypePack struct {
	Node
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

	b.WriteString(g.Node.String())
	b.WriteString(fmt.Sprintf("Name: %s", g.Name))

	return b.String()
}

func DecodeGenericTypePack(data json.RawMessage) (INode, error) {
	var raw GenericTypePack
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

func (g GenericTypePack) Source(string, int) (string, error) {
	// GenericTypePack doesn't have a Location field
	return "", errors.New("generic type pack has no location")
}

type Local[T any] struct {
	LuauType *T     `json:"luauType"` // for now it's probably nil?
	Name     string `json:"name"`
	NodeLoc
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

func (n Local[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(Local[INode])
	if !ok {
		return "", fmt.Errorf("expected Local[INode], got %T", n)
	}

	// if in.LuauType == nil {
	return in.Name, nil
	// }

	// lts, err := (*in.LuauType).Source(og, indent)
	// if err != nil {
	// 	return "", fmt.Errorf("error getting luau type source: %w", err)
	// }

	// return fmt.Sprintf("%s: %s", in.Name, lts), nil
}

func DecodeLocalKnown(raw Local[json.RawMessage], addStatBlock AddStatBlock, depth int) (Local[INode], error) {
	var luauTypeMaybe *INode
	if raw.LuauType != nil {
		luauTypeNode, err := decodeNode(*raw.LuauType, addStatBlock, depth+1)
		if err != nil {
			return Local[INode]{}, fmt.Errorf("error decoding luau type: %w", err)
		}
		luauTypeMaybe = &luauTypeNode
	}

	return Local[INode]{
		LuauType: luauTypeMaybe,
		Name:     raw.Name,
		NodeLoc:  raw.NodeLoc,
	}, nil
}

func DecodeLocal(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw Local[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	return DecodeLocalKnown(raw, addStatBlock, depth+1)
}

type StatAssign[T any] struct {
	NodeLoc
	Vars   []T `json:"vars"`
	Values []T `json:"values"`
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

func (n StatAssign[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatAssign[INode])
	if !ok {
		return "", fmt.Errorf("expected StatAssign[INode], got %T", n)
	}

	VarStrings := make([]string, len(in.Vars))
	for i, node := range in.Vars {
		ns, err := node.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting var source: %w", err)
		}
		VarStrings[i] = ns
	}

	ValueStrings := make([]string, len(in.Values))
	for i, node := range in.Values {
		ns, err := node.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting value source: %w", err)
		}
		ValueStrings[i] = ns
	}

	return IndentSize(indent) + fmt.Sprintf("%s = %s", strings.Join(VarStrings, ", "), strings.Join(ValueStrings, ", ")), nil
}

func DecodeStatAssign(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatAssign[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	vars := make([]INode, len(raw.Vars))
	for i, v := range raw.Vars {
		n, err := decodeNode(v, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding var node: %w", err)
		}
		vars[i] = n
	}

	values := make([]INode, len(raw.Values))
	for i, v := range raw.Values {
		n, err := decodeNode(v, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding value node: %w", err)
		}
		values[i] = n
	}

	return StatAssign[INode]{
		NodeLoc: raw.NodeLoc,
		Vars:    vars,
		Values:  values,
	}, nil
}

type StatBlock[T any] struct {
	NodeLoc
	HasEnd            bool       `json:"hasEnd"`
	Body              []T        `json:"body"`
	CommentsContained *[]Comment // not in json
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

func (n StatBlock[T]) Source(og string, indent int) (string, error) {
	in, ok := any(n).(StatBlock[INode])
	if !ok {
		// return n.Location.GetFromSource(og)
		return "", fmt.Errorf("expected StatBlock[INode], got %T", n)
	}

	ccs := *in.CommentsContained

	var b strings.Builder
	for i, bnode := range in.Body {
		loc := bnode.GetLocation()

		var commentsDone int
		for _, c := range ccs {
			if c.Location.Start.After(loc.End) {
				break
			}

			cs, err := c.Source(og)
			if err != nil {
				return "", fmt.Errorf("error getting comment source: %w", err)
			}

			b.WriteString(IndentSize(indent) + cs)
			b.WriteByte('\n')
			commentsDone++
		}
		ccs = ccs[commentsDone:]

		if bnode.Type() == "AstStatBlock" {
			bs, err := bnode.Source(og, indent+1)
			if err != nil {
				return "", err
			}

			b.WriteString(IndentSize(indent) + fmt.Sprintf("do\n%s\n", bs))
			b.WriteString(IndentSize(indent) + "end")
			continue
		}

		bs, err := bnode.Source(og, indent)
		if err != nil {
			return "", err
		}

		b.WriteString(bs)
		if i < len(in.Body)-1 {
			b.WriteByte('\n')
		}
	}

	for _, c := range ccs {
		cs, err := c.Source(og)
		if err != nil {
			return "", fmt.Errorf("error getting comment source: %w", err)
		}

		b.WriteByte('\n')
		b.WriteString(IndentSize(indent) + cs)
	}

	return b.String(), nil
}

func DecodeStatBlockKnown(raw StatBlock[json.RawMessage], addStatBlock AddStatBlock, depth int) (StatBlock[INode], error) {
	body := make([]INode, len(raw.Body))
	for i, bn := range raw.Body {
		n, err := decodeNode(bn, addStatBlock, depth+1)
		if err != nil {
			return StatBlock[INode]{}, fmt.Errorf("error decoding body node: %w", err)
		}
		body[i] = n
	}

	// hi sb, my good old friend
	sb := StatBlock[INode]{
		NodeLoc:           raw.NodeLoc,
		HasEnd:            raw.HasEnd,
		Body:              body,
		CommentsContained: &[]Comment{},
	}
	addStatBlock(sb, depth)
	return sb, nil
}

func DecodeStatBlock(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatBlock[json.RawMessage] // rawblocks man
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	return DecodeStatBlockKnown(raw, addStatBlock, depth+1)
}

type StatBreak struct {
	NodeLoc
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

func (n StatBreak) Source(_ string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	return IndentSize(indent) + "break", nil
}

func DecodeStatBreak(data json.RawMessage) (INode, error) {
	var raw StatBreak
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

type StatCompoundAssign[T any] struct {
	NodeLoc
	Op    string `json:"op"`
	Var   T      `json:"var"`
	Value T      `json:"value"`
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

func (n StatCompoundAssign[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatCompoundAssign[INode])
	if !ok {
		return "", fmt.Errorf("expected StatCompoundAssign[INode], got %T", n)
	}

	svar, err := in.Var.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting var source: %w", err)
	}

	svalue, err := in.Value.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting value source: %w", err)
	}

	op := BinopToSource(in.Op)

	return IndentSize(indent) + fmt.Sprintf("%s %s= %s", svar, op, svalue), nil
}

func DecodeStatCompoundAssign(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatCompoundAssign[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	varNode, err := decodeNode(raw.Var, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding var: %w", err)
	}

	valueNode, err := decodeNode(raw.Value, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding value: %w", err)
	}

	return StatCompoundAssign[INode]{
		NodeLoc: raw.NodeLoc,
		Op:      raw.Op,
		Var:     varNode,
		Value:   valueNode,
	}, nil
}

type StatContinue struct {
	NodeLoc
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

func (n StatContinue) Source(_ string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	return IndentSize(indent) + "continue", nil
}

func DecodeStatContinue(data json.RawMessage) (INode, error) {
	var raw StatContinue
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

type StatDeclareClass[T any] struct {
	NodeLoc
	Name      string  `json:"name"`
	SuperName *string `json:"superName"`
	Props     []T     `json:"props"`
	Indexer   *T      `json:"indexer"`
}

func (n StatDeclareClass[T]) Type() string {
	return "AstStatDeclareClass"
}

func (n StatDeclareClass[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
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

func (n StatDeclareClass[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatDeclareClass[INode])
	if !ok {
		return "", fmt.Errorf("expected StatDeclareClass[INode], got %T", n)
	}

	propStrings := make([]string, len(in.Props))
	for i, prop := range in.Props {
		sprop, err := prop.Source(og, indent+1)
		if err != nil {
			return "", fmt.Errorf("error getting prop source: %w", err)
		}
		propStrings[i] = sprop
	}

	psi := strings.Join(propStrings, "\n")

	if in.SuperName == nil {
		return fmt.Sprintf("declare class %s\n%s\nend", in.Name, psi), nil
	}
	return fmt.Sprintf("declare class %s extends %s\n%s\nend", in.Name, *in.SuperName, psi), nil
}

func DecodeStatDeclareClass(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatDeclareClass[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	props := make([]INode, len(raw.Props))
	for i, prop := range raw.Props {
		n, err := decodeNode(prop, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding prop node: %w", err)
		}
		props[i] = n
	}

	var indexerNodeMaybe *INode
	if raw.Indexer != nil {
		indexerNode, err := decodeNode(*raw.Indexer, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding indexer: %w", err)
		}
		indexerNodeMaybe = &indexerNode
	}

	return StatDeclareClass[INode]{
		NodeLoc:   raw.NodeLoc,
		Name:      raw.Name,
		SuperName: raw.SuperName,
		Props:     props,
		Indexer:   indexerNodeMaybe,
	}, nil
}

type StatExpr[T any] struct {
	NodeLoc
	Expr T `json:"expr"`
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

func (n StatExpr[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatExpr[INode])
	if !ok {
		return "", fmt.Errorf("expected StatExpr[INode], got %T", n)
	}

	sexpr, err := in.Expr.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting expr source: %w", err)
	}

	return IndentSize(indent) + sexpr, nil
}

func DecodeStatExpr(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatExpr[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	n, err := decodeNode(raw.Expr, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %w", err)
	}

	return StatExpr[INode]{
		NodeLoc: raw.NodeLoc,
		Expr:    n,
	}, nil
}

type StatFor[T any] struct {
	NodeLoc
	Var   T            `json:"var"`
	From  T            `json:"from"`
	To    T            `json:"to"`
	Step  *T           `json:"step"`
	Body  StatBlock[T] `json:"body"`
	HasDo bool         `json:"hasDo"`
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
	b.WriteString("\nStep:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Step), 4))
	b.WriteString("\nBody:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Body), 4))
	b.WriteString(fmt.Sprintf("\nHasDo: %t\n", n.HasDo))

	return b.String()
}

func (n StatFor[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatFor[INode])
	if !ok {
		return "", fmt.Errorf("expected StatFor[INode], got %T", n)
	}

	svar, err := in.Var.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting var source: %w", err)
	}

	sfrom, err := in.From.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting from source: %w", err)
	}

	sto, err := in.To.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting to source: %w", err)
	}

	sbody, err := in.Body.Source(og, indent+1)
	if err != nil {
		return "", fmt.Errorf("error getting body source: %w", err)
	}

	var b strings.Builder
	b.WriteString(IndentSize(indent) + fmt.Sprintf("for %s = %s, %s", svar, sfrom, sto))

	if in.Step != nil {
		sstep, err := (*in.Step).Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting step source: %w", err)
		}

		b.WriteString(fmt.Sprintf(", %s\n", sstep))
	}

	b.WriteString(fmt.Sprintf(" do\n%s\n", sbody))
	b.WriteString(IndentSize(indent) + "end")
	return b.String(), nil
}

func DecodeStatFor(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatFor[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	svar, err := decodeNode(raw.Var, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding var: %w", err)
	}

	sfrom, err := decodeNode(raw.From, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding from: %w", err)
	}

	sto, err := decodeNode(raw.To, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding to: %w", err)
	}

	sbody, err := DecodeStatBlockKnown(raw.Body, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding body: %w", err)
	}

	return StatFor[INode]{
		NodeLoc: raw.NodeLoc,
		Var:     svar,
		From:    sfrom,
		To:      sto,
		Body:    sbody,
		HasDo:   raw.HasDo,
	}, nil
}

type StatForIn[T any] struct {
	NodeLoc
	Vars   []Local[T]   `json:"vars"`
	Values []T          `json:"values"`
	Body   StatBlock[T] `json:"body"`
	HasIn  bool         `json:"hasIn"`
	HasDo  bool         `json:"hasDo"`
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

func (n StatForIn[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatForIn[INode])
	if !ok {
		return "", fmt.Errorf("expected StatForIn[INode], got %T", n)
	}

	vars := make([]string, len(in.Vars))
	for i, v := range in.Vars {
		svar, err := v.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting source for var: %w", err)
		}
		vars[i] = svar
	}

	values := make([]string, len(in.Values))
	for i, v := range in.Values {
		svalue, err := v.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting source for value: %w", err)
		}
		values[i] = svalue
	}

	sbody, err := in.Body.Source(og, indent+1)
	if err != nil {
		return "", fmt.Errorf("error getting source for body: %w", err)
	}

	return fmt.Sprintf("for %s in %s do\n%s\nend",
		strings.Join(vars, ", "), strings.Join(values, ", "), sbody), nil
}

func DecodeStatForIn(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatForIn[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	vars := make([]Local[INode], len(raw.Vars))
	for i, v := range raw.Vars {
		n, err := DecodeLocalKnown(v, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding var node: %w", err)
		}
		vars[i] = n
	}

	values := make([]INode, len(raw.Values))
	for i, v := range raw.Values {
		n, err := decodeNode(v, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding value node: %w", err)
		}
		values[i] = n
	}

	bodyNode, err := DecodeStatBlockKnown(raw.Body, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding body: %w", err)
	}

	return StatForIn[INode]{
		NodeLoc: raw.NodeLoc,
		Vars:    vars,
		Values:  values,
		Body:    bodyNode,
		HasIn:   raw.HasIn,
		HasDo:   raw.HasDo,
	}, nil
}

type StatFunction[T any] struct {
	NodeLoc
	Name T               `json:"name"`
	Func ExprFunction[T] `json:"func"`
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

func (n StatFunction[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatFunction[INode])
	if !ok {
		return "", fmt.Errorf("expected StatFunction[INode], got %T", n)
	}

	ifunc, ok := any(in.Func).(ExprFunction[INode])
	if !ok {
		return "", fmt.Errorf("expected Func to be ExprFunction[INode], got %T", in.Func)
	}

	fs, err := ifunc.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting function source: %w", err)
	}

	if len(ifunc.Attributes) == 0 {
		return fs, err
	}

	attributeStrings := make([]string, len(ifunc.Attributes))
	for i, attr := range ifunc.Attributes {
		ns, err := attr.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting attribute source: %w", err)
		}
		attributeStrings[i] = ns
	}

	return fmt.Sprintf("%s\n%s", strings.Join(attributeStrings, "\n"), fs), nil
}

func DecodeStatFunction(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatFunction[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	nameNode, err := decodeNode(raw.Name, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding name: %w", err)
	}

	funcNode, err := DecodeExprFunctionKnown(raw.Func, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding func: %w", err)
	}

	return StatFunction[INode]{
		NodeLoc: raw.NodeLoc,
		Name:    nameNode,
		Func:    funcNode,
	}, nil
}

type StatIf[T any] struct {
	NodeLoc
	Condition T            `json:"condition"`
	ThenBody  StatBlock[T] `json:"thenbody"`
	ElseBody  *T           `json:"elsebody"` // StatBlock[T] | StatIf[T]
	HasThen   bool         `json:"hasThen"`
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

func (n StatIf[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatIf[INode])
	if !ok {
		return "", fmt.Errorf("expected StatIf[INode], got %T", n)
	}

	scond, err := in.Condition.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting condition source: %w", err)
	}

	sthen, err := in.ThenBody.Source(og, indent+1)
	if err != nil {
		return "", fmt.Errorf("error getting then body source: %w", err)
	}

	var b strings.Builder
	b.WriteString(IndentSize(indent) + fmt.Sprintf("if %s then\n%s\n", scond, sthen))

	if in.ElseBody == nil {
		b.WriteString(IndentSize(indent) + "end")
		return b.String(), nil
	}
	eb := *in.ElseBody

	if eb.Type() == "AstStatIf" {
		selse, err := eb.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting else-if source: %w", err)
		}
		b.WriteString(IndentSize(indent) + "else" + selse)
		return b.String(), nil
	} else if eb.Type() == "AstStatBlock" {
		ebb := eb.(StatBlock[INode])
		if len(ebb.Body) == 1 && ebb.Body[0].Type() == "AstStatIf" {
			selse, err := ebb.Body[0].Source(og, indent)
			if err != nil {
				return "", fmt.Errorf("error getting else-if source: %w", err)
			}
			b.WriteString(IndentSize(indent) + "else" + selse)
			return b.String(), nil
		}
	}

	selse, err := eb.Source(og, indent+1)
	if err != nil {
		return "", fmt.Errorf("error getting else body source: %w", err)
	}

	b.WriteString(IndentSize(indent) + fmt.Sprintf("else\n%s\n", selse))
	b.WriteString(IndentSize(indent) + "end")
	return b.String(), nil
}

func DecodeStatIf(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatIf[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	condition, err := decodeNode(raw.Condition, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding condition: %w", err)
	}

	thenBody, err := DecodeStatBlockKnown(raw.ThenBody, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding then body: %w", err)
	}

	var elseBodyMaybe *INode
	if raw.ElseBody != nil {
		elseBody, err := decodeNode(*raw.ElseBody, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding else body: %w", err)
		}
		elseBodyMaybe = &elseBody
	}

	return StatIf[INode]{
		NodeLoc:   raw.NodeLoc,
		Condition: condition,
		ThenBody:  thenBody,
		ElseBody:  elseBodyMaybe,
		HasThen:   raw.HasThen,
	}, nil
}

type StatLocal[T any] struct {
	NodeLoc
	Vars   []Local[T] `json:"vars"`
	Values []T        `json:"values"`
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

func (n StatLocal[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatLocal[INode])
	if !ok {
		return "", fmt.Errorf("expected StatLocal[INode], got %T", n)
	}

	VarStrings := make([]string, len(in.Vars))
	for i, node := range in.Vars {
		ns, err := node.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting var source: %w", err)
		}
		VarStrings[i] = ns
	}

	if len(in.Values) == 0 {
		return fmt.Sprintf("local %s", strings.Join(VarStrings, ", ")), nil
	}

	ValueStrings := make([]string, len(in.Values))
	for i, node := range in.Values {
		ns, err := node.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting value source: %w", err)
		}
		ValueStrings[i] = ns
	}

	return IndentSize(indent) + fmt.Sprintf("local %s = %s", strings.Join(VarStrings, ", "), strings.Join(ValueStrings, ", ")), nil
}

func DecodeStatLocal(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatLocal[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	vars := make([]Local[INode], len(raw.Vars))
	for i, v := range raw.Vars {
		n, err := DecodeLocalKnown(v, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding var node: %w", err)
		}
		vars[i] = n
	}

	values := make([]INode, len(raw.Values))
	for i, v := range raw.Values {
		n, err := decodeNode(v, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding value node: %w", err)
		}
		values[i] = n
	}

	return StatLocal[INode]{
		NodeLoc: raw.NodeLoc,
		Vars:    vars,
		Values:  values,
	}, nil
}

type StatLocalFunction[T any] struct {
	NodeLoc
	Name Local[T]        `json:"name"`
	Func ExprFunction[T] `json:"func"`
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

func (n StatLocalFunction[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatLocalFunction[INode])
	if !ok {
		return "", fmt.Errorf("expected StatLocalFunction[INode], got %T", n)
	}

	ifunc, ok := any(in.Func).(ExprFunction[INode])
	if !ok {
		return "", fmt.Errorf("expected Func to be ExprFunction[INode], got %T", in.Func)
	}

	fs, err := ifunc.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting function source: %w", err)
	}

	if len(ifunc.Attributes) == 0 {
		return fmt.Sprintf("local %s", fs), nil
	}

	attributeStrings := make([]string, len(ifunc.Attributes))
	for i, attr := range ifunc.Attributes {
		ns, err := attr.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting attribute source: %w", err)
		}
		attributeStrings[i] = ns
	}

	return fmt.Sprintf("%s\nlocal %s", strings.Join(attributeStrings, "\n"), fs), nil
}

func DecodeStatLocalFunction(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatLocalFunction[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	nameNode, err := DecodeLocalKnown(raw.Name, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding name: %w", err)
	}

	funcNode, err := DecodeExprFunctionKnown(raw.Func, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding func: %w", err)
	}

	return StatLocalFunction[INode]{
		NodeLoc: raw.NodeLoc,
		Name:    nameNode,
		Func:    funcNode,
	}, nil
}

type StatRepeat[T any] struct {
	NodeLoc
	Condition T            `json:"condition"`
	Body      StatBlock[T] `json:"body"`
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

func (n StatRepeat[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatRepeat[INode])
	if !ok {
		return "", fmt.Errorf("expected StatRepeat[INode], got %T", n)
	}

	scond, err := in.Condition.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting condition source: %w", err)
	}

	sbody, err := in.Body.Source(og, indent+1)
	if err != nil {
		return "", fmt.Errorf("error getting body source: %w", err)
	}

	// return fmt.Sprintf("repeat\n%s\nuntil %s", sbody, scond), nil
	var b strings.Builder
	b.WriteString(IndentSize(indent) + fmt.Sprintf("repeat\n%s\nuntil %s", sbody, scond))
	b.WriteString(IndentSize(indent) + "end")
	return b.String(), nil
}

func DecodeStatRepeat(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatRepeat[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	condition, err := decodeNode(raw.Condition, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding condition: %w", err)
	}

	body, err := DecodeStatBlockKnown(raw.Body, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding body: %w", err)
	}

	return StatRepeat[INode]{
		NodeLoc:   raw.NodeLoc,
		Condition: condition,
		Body:      body,
	}, nil
}

type StatReturn[T any] struct {
	NodeLoc
	List []T `json:"list"`
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

func (n StatReturn[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatReturn[INode])
	if !ok {
		return "", fmt.Errorf("error asserting node type")
	}

	if len(in.List) == 0 {
		return IndentSize(indent) + "return", nil
	}

	listStrings := make([]string, len(in.List))
	for i, item := range in.List {
		sitem, err := item.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting item source: %w", err)
		}
		listStrings[i] = sitem
	}
	return IndentSize(indent) + fmt.Sprintf("return %s", strings.Join(listStrings, ", ")), nil
}

func DecodeStatReturn(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatReturn[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	list := make([]INode, len(raw.List))
	for i, item := range raw.List {
		n, err := decodeNode(item, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding list item: %w", err)
		}
		list[i] = n
	}

	return StatReturn[INode]{
		NodeLoc: raw.NodeLoc,
		List:    list,
	}, nil
}

type StatTypeAlias[T any] struct {
	NodeLoc
	Name         string            `json:"name"`
	Generics     []GenericType     `json:"generics"`
	GenericPacks []GenericTypePack `json:"genericPacks"` // genericPacks always come after the generics
	Value        T                 `json:"value"`
	Exported     bool              `json:"exported"`
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

func (n StatTypeAlias[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatTypeAlias[INode])
	if !ok {
		return "", fmt.Errorf("expected StatTypeAlias[INode], got %T", n)
	}

	svalue, err := in.Value.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting value source: %w", err)
	}

	if len(in.Generics) == 0 && len(in.GenericPacks) == 0 {
		return fmt.Sprintf("type %s = %s", in.Name, svalue), nil
	}

	genericStrings := make([]string, len(in.Generics))
	for i, g := range in.Generics {
		genericStrings[i] = g.Name
	}

	genericPackStrings := make([]string, len(in.GenericPacks))
	for i, gp := range in.GenericPacks {
		genericPackStrings[i] = gp.Name + "..."
	}

	allGenerics := append(genericStrings, genericPackStrings...)

	return fmt.Sprintf("type %s<%s> = %s", in.Name, strings.Join(allGenerics, ", "), svalue), nil
}

func DecodeStatTypeAlias(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatTypeAlias[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	valueNode, err := decodeNode(raw.Value, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding value: %w", err)
	}

	return StatTypeAlias[INode]{
		NodeLoc:      raw.NodeLoc,
		Name:         raw.Name,
		Generics:     raw.Generics,
		GenericPacks: raw.GenericPacks,
		Value:        valueNode,
		Exported:     raw.Exported,
	}, nil
}

type StatWhile[T any] struct {
	NodeLoc
	Condition T            `json:"condition"`
	Body      StatBlock[T] `json:"body"`
	HasDo     bool         `json:"hasDo"`
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

func (n StatWhile[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(StatWhile[INode])
	if !ok {
		return "", fmt.Errorf("expected StatWhile[INode], got %T", n)
	}
	// return in.Location.GetFromSource(og)

	scond, err := in.Condition.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting condition source: %w", err)
	}

	sbody, err := in.Body.Source(og, indent+1)
	if err != nil {
		return "", fmt.Errorf("error getting body source: %w", err)
	}

	var b strings.Builder
	b.WriteString(IndentSize(indent) + fmt.Sprintf("while %s do\n%s\n", scond, sbody))
	b.WriteString(IndentSize(indent) + "end")
	return b.String(), nil
}

func DecodeStatWhile(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw StatWhile[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	condition, err := decodeNode(raw.Condition, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding condition: %w", err)
	}

	body, err := DecodeStatBlockKnown(raw.Body, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding body: %w", err)
	}

	return StatWhile[INode]{
		NodeLoc:   raw.NodeLoc,
		Condition: condition,
		Body:      body,
		HasDo:     raw.HasDo,
	}, nil
}

type TableProp[T any] struct {
	Name string `json:"name"`
	NodeLoc
	PropType T `json:"propType"`
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

func (n TableProp[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(TableProp[INode])
	if !ok {
		return "", fmt.Errorf("expected TableProp[INode], got %T", n)
	}

	propTypeSource, err := in.PropType.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting prop type source: %w", err)
	}

	return fmt.Sprintf("%s: %s", in.Name, propTypeSource), nil
}

func DecodeTableProp(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw TableProp[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	propTypeNode, err := decodeNode(raw.PropType, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding prop type: %w", err)
	}

	return TableProp[INode]{
		NodeLoc:  raw.NodeLoc,
		Name:     raw.Name,
		PropType: propTypeNode,
	}, nil
}

type TypeFunction[T any] struct {
	NodeLoc
	Attributes   []Attr              `json:"attributes"`
	Generics     []GenericType       `json:"generics"`
	GenericPacks []GenericTypePack   `json:"genericPacks"`
	ArgTypes     TypeList[T]         `json:"argTypes"`
	ArgNames     []*ArgumentName     `json:"argNames"`
	ReturnTypes  TypePackExplicit[T] `json:"returnTypes"`
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

func (n TypeFunction[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(TypeFunction[INode])
	if !ok {
		return "", fmt.Errorf("expected TypeFunction[INode], got %T", n)
	}

	argTypeStrings := make([]string, len(in.ArgTypes.Types))
	for i, argType := range in.ArgTypes.Types {
		sargType, err := argType.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting arg type source: %w", err)
		}
		argTypeStrings[i] = sargType
	}

	if in.ArgTypes.TailType != nil {
		ts, err := (*in.ArgTypes.TailType).Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting source for tail type: %w", err)
		}
		argTypeStrings = append(argTypeStrings, ts)
	}

	for i, argName := range in.ArgNames {
		if argName != nil {
			argTypeStrings[i] = fmt.Sprintf("%s: %s", argName.Name, argTypeStrings[i])
		}
	}

	sreturn, err := in.ReturnTypes.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting return types source: %w", err)
	}

	if len(in.Generics) == 0 && len(in.GenericPacks) == 0 {
		return fmt.Sprintf("(%s) -> %s", strings.Join(argTypeStrings, ", "), sreturn), nil
	}

	genericStrings := make([]string, len(in.Generics))
	for i, g := range in.Generics {
		genericStrings[i] = g.Name
	}

	genericPackStrings := make([]string, len(in.GenericPacks))
	for i, gp := range in.GenericPacks {
		genericPackStrings[i] = gp.Name + "..."
	}

	allGenerics := append(genericStrings, genericPackStrings...)

	return fmt.Sprintf("<%s>(%s) -> %s", strings.Join(allGenerics, ", "), strings.Join(argTypeStrings, ", "), sreturn), nil
}

func DecodeTypeFunction(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw TypeFunction[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	argTypesNode, err := DecodeTypeListKnown(raw.ArgTypes, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding arg types: %w", err)
	}

	returnTypesNode, err := DecodeTypePackExplicitKnown(raw.ReturnTypes, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding return types: %w", err)
	}

	return TypeFunction[INode]{
		NodeLoc:      raw.NodeLoc,
		Attributes:   raw.Attributes,
		Generics:     raw.Generics,
		GenericPacks: raw.GenericPacks,
		ArgTypes:     argTypesNode,
		ArgNames:     raw.ArgNames,
		ReturnTypes:  returnTypesNode,
	}, nil
}

type TypeGroup[T any] struct {
	NodeLoc
	Inner T `json:"inner"`
}

func (n TypeGroup[T]) Type() string {
	return "AstTypeGroup"
}

func (n TypeGroup[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nInner:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Inner), 4))

	return b.String()
}

func (n TypeGroup[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	iinner, ok := any(n.Inner).(INode)
	if !ok {
		return "", fmt.Errorf("expected Inner to be INode, got %T", n.Inner)
	}

	// now you can parse your way out of hell, only 40 000 ast nodes
	sinner, err := iinner.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting source for inner node: %w", err)
	}

	return fmt.Sprintf("(%s)", sinner), nil
}

func DecodeTypeGroup(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw TypeGroup[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	innerNode, err := decodeNode(raw.Inner, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding inner node: %w", err)
	}

	return TypeGroup[INode]{
		NodeLoc: raw.NodeLoc,
		Inner:   innerNode,
	}, nil
}

type TypeList[T any] struct {
	Node
	Types    []T `json:"types"`
	TailType *T  `json:"tailType"`
}

func (n TypeList[T]) GetLocation() Location {
	return Location{}
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

	b.WriteString("\nTailType:")
	if n.TailType != nil {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(*n.TailType), 4))
	}

	return b.String()
}

func (n TypeList[T]) Source(og string, indent int) (string, error) {
	// TypeList doesn't seem to have a Location field
	// return "", errors.New("type list has no location")
	in, ok := any(n).(TypeList[INode])
	if !ok {
		return "", fmt.Errorf("expected Types to be []INode, got %T", n.Types)
	}

	var typeStrings []string
	for _, typ := range in.Types {
		s, err := typ.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting source for type: %w", err)
		}
		typeStrings = append(typeStrings, s)
	}

	if in.TailType != nil {
		ts, err := (*in.TailType).Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting source for tail type: %w", err)
		}
		typeStrings = append(typeStrings, ts)
	}

	return strings.Join(typeStrings, ", "), nil
}

func DecodeTypeListKnown(raw TypeList[json.RawMessage], addStatBlock AddStatBlock, depth int) (TypeList[INode], error) {
	types := make([]INode, len(raw.Types))
	for i, typ := range raw.Types {
		n, err := decodeNode(typ, addStatBlock, depth+1)
		if err != nil {
			return TypeList[INode]{}, fmt.Errorf("error decoding type node: %w", err)
		}
		types[i] = n
	}

	var tailType *INode
	if raw.TailType != nil {
		n, err := decodeNode(*raw.TailType, addStatBlock, depth+1)
		if err != nil {
			return TypeList[INode]{}, fmt.Errorf("error decoding tail type node: %w", err)
		}
		tailType = &n
	}

	return TypeList[INode]{
		Node:     raw.Node,
		Types:    types,
		TailType: tailType,
	}, nil
}

func DecodeTypeList(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw TypeList[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	return DecodeTypeListKnown(raw, addStatBlock, depth+1)
}

type TypeOptional struct {
	NodeLoc
}

func (TypeOptional) Type() string {
	return "AstTypeOptional"
}

func (n TypeOptional) Source(string, int) (string, error) {
	// return n.Location.GetFromSource(og)
	return "?", nil
}

func (n TypeOptional) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))

	return b.String()
}

func DecodeTypeOptional(data json.RawMessage) (INode, error) {
	var raw TypeOptional
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	return TypeOptional{
		NodeLoc: raw.NodeLoc,
	}, nil
}

type TypePackExplicit[T any] struct {
	NodeLoc
	TypeList TypeList[T] `json:"typeList"`
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

func (n TypePackExplicit[T]) Source(og string, indent int) (string, error) {
	itypelist, ok := any(n.TypeList).(TypeList[INode])
	if !ok {
		return "", fmt.Errorf("expected TypeList[INode], got %T", n.TypeList)
	}

	stypelist, err := itypelist.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting type list source: %w", err)
	}

	if len(itypelist.Types) == 1 {
		return stypelist, nil
	}

	return fmt.Sprintf("(%s)", stypelist), nil
}

func DecodeTypePackExplicitKnown(raw TypePackExplicit[json.RawMessage], addStatBlock AddStatBlock, depth int) (TypePackExplicit[INode], error) {
	typeListNode, err := DecodeTypeListKnown(raw.TypeList, addStatBlock, depth+1)
	if err != nil {
		return TypePackExplicit[INode]{}, fmt.Errorf("error decoding type list: %w", err)
	}

	return TypePackExplicit[INode]{
		NodeLoc:  raw.NodeLoc,
		TypeList: typeListNode,
	}, nil
}

func DecodeTypePackExplicit(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw TypePackExplicit[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	return DecodeTypePackExplicitKnown(raw, addStatBlock, depth+1)
}

type TypePackGeneric struct {
	NodeLoc
	GenericName string `json:"genericName"`
}

func (n TypePackGeneric) Type() string {
	return "AstTypePackGeneric"
}

func (n TypePackGeneric) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nGenericName: %s", n.GenericName))

	return b.String()
}

func (n TypePackGeneric) Source(string, int) (string, error) {
	// return n.Location.GetFromSource(og)
	return n.GenericName + "...", nil
}

func DecodeTypePackGeneric(data json.RawMessage) (INode, error) {
	var raw TypePackGeneric
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}
	return raw, nil
}

type TypePackVariadic[T any] struct {
	NodeLoc
	VariadicType T `json:"variadicType"`
}

func (n TypePackVariadic[T]) Type() string {
	return "AstTypePackVariadic"
}

func (n TypePackVariadic[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nVariadicType: %s", StringMaybeEvaluated(n.VariadicType)))

	return b.String()
}

func (n TypePackVariadic[T]) Source(og string, indent int) (string, error) {
	ivt, ok := any(n.VariadicType).(INode)
	if !ok {
		return "", fmt.Errorf("expected VariadicType to be INode, got %T", n.VariadicType)
	}

	svt, err := ivt.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting VariadicType source: %w", err)
	}

	return "..." + svt, nil
}

func DecodeTypePackVariadic(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw TypePackVariadic[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	variadicType, err := decodeNode(raw.VariadicType, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding VariadicType: %w", err)
	}

	return TypePackVariadic[INode]{
		NodeLoc:      raw.NodeLoc,
		VariadicType: variadicType,
	}, nil
}

type TypeReference[T any] struct {
	NodeLoc
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

func (n TypeReference[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(TypeReference[INode])
	if !ok {
		return "", fmt.Errorf("expected TypeReference[INode], got %T", n)
	}

	if len(in.Parameters) == 0 {
		return in.Name, nil
	}

	paramStrings := make([]string, len(in.Parameters))
	for i, param := range in.Parameters {
		ns, err := param.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting parameter source: %w", err)
		}
		paramStrings[i] = ns
	}

	return fmt.Sprintf("%s<%s>",
		in.Name,
		strings.Join(paramStrings, ", ")), nil
}

func DecodeTypeReference(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw TypeReference[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	parameters := make([]INode, len(raw.Parameters))
	for i, param := range raw.Parameters {
		n, err := decodeNode(param, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding parameter node: %w", err)
		}
		parameters[i] = n
	}

	return TypeReference[INode]{
		NodeLoc:      raw.NodeLoc,
		Name:         raw.Name,
		NameLocation: raw.NameLocation,
		Parameters:   parameters,
	}, nil
}

type TypeSingletonBool struct {
	NodeLoc
	Value bool `json:"value"`
}

func (n TypeSingletonBool) Type() string {
	return "AstTypeSingletonBool"
}

func (n TypeSingletonBool) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nValue: %t", n.Value))

	return b.String()
}

func (n TypeSingletonBool) Source(string, int) (string, error) {
	// return n.Location.GetFromSource(og)
	if n.Value {
		return "true", nil
	}
	return "false", nil
}

func DecodeTypeSingletonBool(data json.RawMessage) (INode, error) {
	var raw TypeSingletonBool
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	return TypeSingletonBool{
		NodeLoc: raw.NodeLoc,
		Value:   raw.Value,
	}, nil
}

type TypeSingletonString struct {
	NodeLoc
	Value string `json:"value"`
}

func (n TypeSingletonString) Type() string {
	return "AstTypeSingletonString"
}

func (n TypeSingletonString) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString(fmt.Sprintf("\nValue: %s", n.Value))

	return b.String()
}

func (n TypeSingletonString) Source(og string, _ int) (string, error) {
	return StringToSource(n.Value), nil
}

func DecodeTypeSingletonString(data json.RawMessage) (INode, error) {
	var raw TypeSingletonString
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	return TypeSingletonString{
		NodeLoc: raw.NodeLoc,
		Value:   raw.Value,
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

func (n Indexer[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(Indexer[INode])
	if !ok {
		return "", fmt.Errorf("expected Indexer[INode], got %T", n)
	}

	sindexType, err := in.IndexType.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting source for index type: %w", err)
	}

	sresultType, err := in.ResultType.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting source for result type: %w", err)
	}

	return fmt.Sprintf("[%s]: %s", sindexType, sresultType), nil
}

type TypeTable[T any] struct {
	NodeLoc
	Props   []T         `json:"props"`
	Indexer *Indexer[T] `json:"indexer"`
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

func (n TypeTable[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, ok := any(n).(TypeTable[INode])
	if !ok {
		return "", fmt.Errorf("expected TypeTable[INode], got %T", n)
	}

	if len(in.Props) == 0 {
		if in.Indexer == nil {
			return "{}", nil
		}

		// { number } etc, these don't work with additional object/hash fields or whatever
		ixr := *in.Indexer

		ixrr, ok := any(ixr.IndexType).(TypeReference[INode])
		if ok && ixrr.Name == "number" { // it doesn't matter if the type is *actually* number. Fun experiment: do `type number = string` in your Luau script and see how much you can upset the typechecker!
			rt, err := ixr.ResultType.Source(og, indent)
			if err != nil {
				return "", fmt.Errorf("error getting source for result type: %w", err)
			}

			return fmt.Sprintf("{ %s }", rt), nil

		}
	}

	var parts []string
	if in.Indexer != nil {
		sindexer, err := (*in.Indexer).Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting indexer index type source: %w", err)
		}
		parts = append(parts, sindexer)
	}

	if len(in.Props) > 0 {
		var propStrings []string
		for _, prop := range in.Props {
			ps, err := prop.Source(og, indent)
			if err != nil {
				return "", fmt.Errorf("error getting prop source: %w", err)
			}
			propStrings = append(propStrings, ps)
		}
		parts = append(parts, propStrings...)
	}

	return fmt.Sprintf("{ %s }", strings.Join(parts, ", ")), nil
}

func DecodeTypeTable(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw TypeTable[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	props := make([]INode, len(raw.Props))
	for i, prop := range raw.Props {
		n, err := decodeNode(prop, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding prop node: %w", err)
		}
		props[i] = n
	}

	var indexerMaybe *Indexer[INode]
	if raw.Indexer != nil {
		indexerNode, err := decodeNode(raw.Indexer.IndexType, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding indexer index type: %w", err)
		}
		resultNode, err := decodeNode(raw.Indexer.ResultType, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding indexer result type: %w", err)
		}
		indexerMaybe = &Indexer[INode]{
			Location:   raw.Indexer.Location,
			IndexType:  indexerNode,
			ResultType: resultNode,
		}
	}

	return TypeTable[INode]{
		NodeLoc: raw.NodeLoc,
		Props:   props,
		Indexer: indexerMaybe,
	}, nil
}

// lol
type TypeTypeof[T any] struct {
	NodeLoc
	Expr T `json:"expr"`
}

func (n TypeTypeof[T]) Type() string {
	return "AstTypeTypeof"
}

func (n TypeTypeof[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nExpr:\n")
	b.WriteString(indentStart(StringMaybeEvaluated(n.Expr), 4))

	return b.String()
}

func (n TypeTypeof[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	iexpr, ok := any(n.Expr).(INode)
	if !ok {
		return "", fmt.Errorf("expected INode, got %T", n.Expr)
	}

	sexpr, err := iexpr.Source(og, indent)
	if err != nil {
		return "", fmt.Errorf("error getting source for expr: %w", err)
	}

	return fmt.Sprintf("typeof(%s)", sexpr), nil
}

func DecodeTypeTypeof(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw TypeTypeof[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	exprNode, err := decodeNode(raw.Expr, addStatBlock, depth+1)
	if err != nil {
		return nil, fmt.Errorf("error decoding expr: %w", err)
	}

	return TypeTypeof[INode]{
		NodeLoc: raw.NodeLoc,
		Expr:    exprNode,
	}, nil
}

type TypeUnion[T any] struct {
	NodeLoc
	Types []T `json:"types"`
}

func (n TypeUnion[T]) Type() string {
	return "AstTypeUnion"
}

func (n TypeUnion[T]) String() string {
	var b strings.Builder

	b.WriteString(n.Node.String())
	b.WriteString(fmt.Sprintf("Location: %s", n.Location))
	b.WriteString("\nTypes:")

	for _, typ := range n.Types {
		b.WriteByte('\n')
		b.WriteString(indentStart(StringMaybeEvaluated(typ), 4))
	}

	return b.String()
}

func (n TypeUnion[T]) Source(og string, indent int) (string, error) {
	// return n.Location.GetFromSource(og)
	in, err := any(n).(TypeUnion[INode])
	if !err {
		return "", fmt.Errorf("expected TypeUnion[INode], got %T", n)
	}

	var seenOptional bool
	var newTypes []INode
	for _, typ := range in.Types {
		if typ.Type() == "AstTypeOptional" {
			seenOptional = true
			continue
		}
		newTypes = append(newTypes, typ)
	}

	if len(newTypes) == 1 {
		ts, err := newTypes[0].Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting source for type: %w", err)
		}

		if seenOptional {
			return ts + "?", nil
		}
		return ts, nil
	}

	var sources []string
	for _, typ := range newTypes {
		src, err := typ.Source(og, indent)
		if err != nil {
			return "", fmt.Errorf("error getting type source: %w", err)
		}
		sources = append(sources, src)
	}

	joined := strings.Join(sources, " | ")

	if seenOptional {
		return fmt.Sprintf("(%s)?", joined), nil
	}
	return joined, nil
}

func DecodeTypeUnion(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var raw TypeUnion[json.RawMessage]
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("error decoding: %w", err)
	}

	types := make([]INode, len(raw.Types))
	for i, typ := range raw.Types {
		n, err := decodeNode(typ, addStatBlock, depth+1)
		if err != nil {
			return nil, fmt.Errorf("error decoding type node: %w", err)
		}
		types[i] = n
	}

	return TypeUnion[INode]{
		NodeLoc: raw.NodeLoc,
		Types:   types,
	}, nil
}

// decoding

func decodeNode(data json.RawMessage, addStatBlock AddStatBlock, depth int) (INode, error) {
	var node Node
	if err := json.Unmarshal(data, &node); err != nil {
		return nil, fmt.Errorf("error decoding node: %w", err)
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
	case "AstDeclaredClassProp":
		return ret(DecodeDeclaredClassProp(data, addStatBlock, depth))
	case "AstExprBinary":
		return ret(DecodeExprBinary(data, addStatBlock, depth))
	case "AstExprCall":
		return ret(DecodeExprCall(data, addStatBlock, depth))
	case "AstExprConstantBool":
		return ret(DecodeExprConstantBool(data))
	case "AstExprConstantNil":
		return ret(DecodeExprConstantNil(data))
	case "AstExprConstantNumber":
		return ret(DecodeExprConstantNumber(data))
	case "AstExprConstantString":
		return ret(DecodeExprConstantString(data))
	case "AstExprFunction":
		return ret(DecodeExprFunction(data, addStatBlock, depth))
	case "AstExprGlobal":
		return ret(DecodeExprGlobal(data))
	case "AstExprGroup":
		return ret(DecodeExprGroup(data, addStatBlock, depth))
	case "AstExprIfElse":
		return ret(DecodeExprIfElse(data, addStatBlock, depth))
	case "AstExprIndexExpr":
		return ret(DecodeExprIndexExpr(data, addStatBlock, depth))
	case "AstExprIndexName":
		return ret(DecodeExprIndexName(data, addStatBlock, depth))
	case "AstExprInterpString":
		return ret(DecodeExprInterpString(data, addStatBlock, depth))
	case "AstExprLocal":
		return ret(DecodeExprLocal(data, addStatBlock, depth))
	case "AstExprTable":
		return ret(DecodeExprTable(data, addStatBlock, depth))
	case "AstExprTableItem":
		return ret(DecodeExprTableItem(data, addStatBlock, depth))
	case "AstExprTypeAssertion":
		return ret(DecodeExprTypeAssertion(data, addStatBlock, depth))
	case "AstExprVarargs":
		return ret(DecodeExprVarargs(data))
	case "AstExprUnary":
		return ret(DecodeExprUnary(data, addStatBlock, depth))
	case "AstGenericType":
		return ret(DecodeGenericType(data))
	case "AstGenericTypePack":
		return ret(DecodeGenericTypePack(data))
	case "AstLocal":
		return ret(DecodeLocal(data, addStatBlock, depth))
	case "AstStatAssign":
		return ret(DecodeStatAssign(data, addStatBlock, depth))
	case "AstStatBlock":
		return ret(DecodeStatBlock(data, addStatBlock, depth))
	case "AstStatBreak":
		return ret(DecodeStatBreak(data))
	case "AstStatCompoundAssign":
		return ret(DecodeStatCompoundAssign(data, addStatBlock, depth))
	case "AstStatContinue":
		return ret(DecodeStatContinue(data))
	case "AstStatDeclareClass":
		return ret(DecodeStatDeclareClass(data, addStatBlock, depth))
	case "AstStatExpr":
		return ret(DecodeStatExpr(data, addStatBlock, depth))
	case "AstStatFor":
		return ret(DecodeStatFor(data, addStatBlock, depth))
	case "AstStatForIn":
		return ret(DecodeStatForIn(data, addStatBlock, depth))
	case "AstStatFunction":
		return ret(DecodeStatFunction(data, addStatBlock, depth))
	case "AstStatIf":
		return ret(DecodeStatIf(data, addStatBlock, depth))
	case "AstStatLocal":
		return ret(DecodeStatLocal(data, addStatBlock, depth))
	case "AstStatLocalFunction":
		return ret(DecodeStatLocalFunction(data, addStatBlock, depth))
	case "AstStatRepeat":
		return ret(DecodeStatRepeat(data, addStatBlock, depth))
	case "AstStatReturn":
		return ret(DecodeStatReturn(data, addStatBlock, depth))
	case "AstStatTypeAlias":
		return ret(DecodeStatTypeAlias(data, addStatBlock, depth))
	case "AstStatWhile":
		return ret(DecodeStatWhile(data, addStatBlock, depth))
	case "AstTableProp":
		return ret(DecodeTableProp(data, addStatBlock, depth))
	case "AstTypeFunction":
		return ret(DecodeTypeFunction(data, addStatBlock, depth))
	case "AstTypeGroup":
		return ret(DecodeTypeGroup(data, addStatBlock, depth))
	case "AstTypeList":
		return ret(DecodeTypeList(data, addStatBlock, depth))
	case "AstTypeOptional":
		return ret(DecodeTypeOptional(data))
	case "AstTypePackExplicit":
		return ret(DecodeTypePackExplicit(data, addStatBlock, depth))
	case "AstTypePackGeneric":
		return ret(DecodeTypePackGeneric(data))
	case "AstTypePackVariadic":
		return ret(DecodeTypePackVariadic(data, addStatBlock, depth))
	case "AstTypeReference":
		return ret(DecodeTypeReference(data, addStatBlock, depth))
	case "AstTypeSingletonBool":
		return ret(DecodeTypeSingletonBool(data))
	case "AstTypeSingletonString":
		return ret(DecodeTypeSingletonString(data))
	case "AstTypeTable":
		return ret(DecodeTypeTable(data, addStatBlock, depth))
	case "AstTypeTypeof":
		return ret(DecodeTypeTypeof(data, addStatBlock, depth))
	case "AstTypeUnion":
		return ret(DecodeTypeUnion(data, addStatBlock, depth))
	}
	return ret(nil, errors.New("unknown node type"))
}
