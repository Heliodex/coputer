package main

import (
	"fmt"
	"strings"

	"github.com/Heliodex/coputer/ast/lex"
)

// This file implements Minify() for every AST node. Each Minify() renders the node as compact Luau, stripping non-directive comments and all optional whitespace.
//
// Like Source(), every method renders its node starting at column zero and delegates to the Minify() methods of its children. Unlike Source(), statements are concatenated directly wherever Luau's grammar accepts it: a newline is only inserted when two adjacent tokens would otherwise merge, and a semicolon is only inserted when a newline still wouldn't separate two statements (a callable expression followed by a statement starting with `(`). Directives (line comments whose body starts with `!`) are kept because they change how Luau compiles the file.

// minifyBuilder builds a token stream, inserting a single space only when two adjacent tokens would otherwise merge. Child Minify() results are written as single tokens so that every join is checked.
type minifyBuilder struct {
	b    strings.Builder
	last byte
}

func minifyWordByte(c byte) bool {
	return c == '_' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9'
}

// minifyNeedsSpace reports whether a space must separate two adjacent bytes so that they don't lex as a different token.
func minifyNeedsSpace(prev, next byte) bool {
	switch {
	case minifyWordByte(prev) && minifyWordByte(next):
		return true
	case '0' <= prev && prev <= '9' && next == '.':
		return true
	case prev == '-' && next == '-':
		return true
	case prev == '[' && next == '[':
		return true
	case prev == '>' && next == '=':
		return true
	}
	return false
}

func (m *minifyBuilder) write(s string) {
	if s == "" {
		return
	}
	if m.b.Len() > 0 && minifyNeedsSpace(m.last, s[0]) {
		m.b.WriteByte(' ')
	}
	m.b.WriteString(s)
	m.last = s[len(s)-1]
}

func (m *minifyBuilder) writeByte(c byte) {
	if m.b.Len() > 0 && minifyNeedsSpace(m.last, c) {
		m.b.WriteByte(' ')
	}
	m.b.WriteByte(c)
	m.last = c
}

func (m *minifyBuilder) String() string {
	return m.b.String()
}

// minifyDirectives renders the directive comments among comments, preserving their original order, each on its own line.
func minifyDirectives(comments []Comment) string {
	var lines []string
	for _, c := range comments {
		if _, ok := directiveCategory(c); ok {
			lines = append(lines, "--"+c.Content)
		}
	}
	return strings.Join(lines, "\n")
}

// minifyLastTokenCallable reports whether the final token of a rendered statement is a prefix expression that a following `(` would be parsed as calling.
func minifyLastTokenCallable(s string) bool {
	l := lex.NewLexer(s)
	var last lex.Lexeme

	for {
		tok := l.Next0()
		if tok.Type == lex.Eof {
			break
		}
		if tok.Type != lex.Comment && tok.Type != lex.BlockComment {
			last = tok
		}
	}

	switch last.Type {
	case ')', ']', lex.Name:
		return true
	}
	return false
}

// minifyStatementSeparator returns the separator to place between two rendered statements: none when they concatenate safely, a newline when the tokens would merge, and a semicolon when a newline still wouldn't split them.
func minifyStatementSeparator(prev, next string) string {
	if prev == "" || next == "" {
		return ""
	}

	p := prev[len(prev)-1]
	c := next[0]

	// a line already ends the previous statement, so the next one starts fresh
	if p == '\n' {
		return ""
	}

	// `f` followed by `(` is parsed as the call `f(...)`, even across a newline
	if c == '(' && minifyLastTokenCallable(prev) {
		return ";"
	}

	// otherwise a newline separates the tokens, which direct concatenation wouldn't
	if minifyNeedsSpace(p, c) {
		return "\n"
	}

	return ""
}

// minifyStatList renders statements with the minimal separator between each pair.
func minifyStatList(stats []AstStat) string {
	if len(stats) == 0 {
		return ""
	}

	var b strings.Builder
	prev := stats[0].Minify()
	b.WriteString(prev)

	for _, stat := range stats[1:] {
		next := stat.Minify()
		b.WriteString(minifyStatementSeparator(prev, next))
		b.WriteString(next)
		prev = next
	}

	return b.String()
}

// minifyBlockBody renders a block's statements, preceded by any directives inside the block. Directive comments must start on their own line, so a leading newline is emitted when present.
func minifyBlockBody(block AstStatBlock) string {
	body := minifyStatList(block.Body)
	directives := minifyDirectives(block.Comments)
	if directives == "" {
		return body
	}
	return "\n" + directives + "\n" + body
}

// minifyRoot renders the top-level chunk, placing any file directives above the code.
func minifyRoot(block AstStatBlock) string {
	directives := minifyDirectives(block.Comments)
	body := minifyStatList(block.Body)

	switch {
	case directives == "":
		return body
	case body == "":
		return directives
	default:
		return directives + "\n" + body
	}
}

// minifyParens wraps an expression in parentheses.
func minifyParens(s string) string {
	return "(" + s + ")"
}

// minifyExpr renders expr in a context requiring at least minPrec, mirroring sourceExpr.
func minifyExpr(expr AstExpr, minPrec int, expand bool) string {
	var group *AstExprGroup
	switch e := expr.(type) {
	case AstExprGroup:
		group = &e
	case *AstExprGroup:
		group = e
	}

	if group != nil {
		inner := group.Expr

		if expand && sourceExprExpands(inner) {
			return minifyParens(minifyExpr(inner, 0, false))
		}

		return minifyExpr(inner, minPrec, expand)
	}

	s := expr.Minify()
	if sourceExprPrecedence(expr) < minPrec {
		return minifyParens(s)
	}
	return s
}

// minifyExprPrec renders expr in a single-value context requiring at least minPrec.
func minifyExprPrec(expr AstExpr, minPrec int) string {
	return minifyExpr(expr, minPrec, false)
}

// minifyExprPostfix renders an expression used as the base of a postfix operation, mirroring sourceExprPostfix.
func minifyExprPostfix(expr AstExpr) string {
	var inner AstExpr
	switch e := expr.(type) {
	case AstExprGroup:
		inner = e.Expr
	case *AstExprGroup:
		inner = e.Expr
	}

	if inner != nil {
		unwrapped := sourceUnwrapGroup(inner)
		if !sourcePostfixBase(unwrapped) {
			return minifyParens(minifyExprPrec(unwrapped, 0))
		}
	}

	return minifyExprPrec(expr, precPrimary)
}

// minifyExprList renders comma-separated expressions, allowing the last one to expand.
func minifyExprList(exprs []AstExpr) string {
	parts := make([]string, len(exprs))
	for i, expr := range exprs {
		parts[i] = minifyExpr(expr, 0, i == len(exprs)-1)
	}
	return strings.Join(parts, ",")
}

// minifyTypeOrPackList renders comma-separated types/type packs.
func minifyTypeOrPackList(items []AstTypeOrPack) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = item.Minify()
	}
	return strings.Join(parts, ",")
}

// minifyGenerics renders a generic list, e.g. `T,U...=string`.
func minifyGenerics(generics []AstGenericType, packs []AstGenericTypePack) string {
	parts := make([]string, 0, len(generics)+len(packs))
	for _, generic := range generics {
		parts = append(parts, generic.Minify())
	}
	for _, pack := range packs {
		parts = append(parts, pack.Minify())
	}
	return strings.Join(parts, ",")
}

// minifyAttrs renders attributes separated by spaces, or "" when there are none.
func minifyAttrs(attrs []AstAttr) string {
	parts := make([]string, len(attrs))
	for i, attr := range attrs {
		parts[i] = attr.Minify()
	}
	return strings.Join(parts, " ")
}

// minifyVarargAnnotation renders a vararg annotation without the leading `...`, which the parameter list already writes.
func minifyVarargAnnotation(pack AstTypePack) string {
	switch p := pack.(type) {
	case AstTypePackVariadic:
		return p.VariadicType.Minify()
	case *AstTypePackVariadic:
		return p.VariadicType.Minify()
	}
	return pack.Minify()
}

// minifyDeclareParams renders the parameter list of a declare function.
func minifyDeclareParams(params AstTypeList, names []AstArgumentName, vararg bool) string {
	parts := make([]string, 0, len(params.Types)+1)

	for i, param := range params.Types {
		p := param.Minify()
		if i < len(names) && names[i].Name != "" {
			p = names[i].Name + ":" + p
		}
		parts = append(parts, p)
	}

	if vararg {
		p := "..."
		if params.TailType != nil {
			p += ":" + minifyVarargAnnotation(*params.TailType)
		}
		parts = append(parts, p)
	} else if params.TailType != nil {
		parts = append(parts, (*params.TailType).Minify())
	}

	return strings.Join(parts, ",")
}

// minifyFunctionRest renders a function body from the generics onwards, i.e. without the leading `function` keyword.
func minifyFunctionRest(n AstExprFunction) string {
	var b minifyBuilder

	if generics := minifyGenerics(n.Generics, n.GenericPacks); generics != "" {
		b.writeByte('<')
		b.write(generics)
		b.writeByte('>')
	}

	b.writeByte('(')
	args := make([]string, 0, len(n.Args)+1)
	for i := range n.Args {
		args = append(args, n.Args[i].Minify())
	}
	if n.Vararg {
		if n.VarargAnnotation != nil {
			args = append(args, "...:"+minifyVarargAnnotation(*n.VarargAnnotation))
		} else {
			args = append(args, "...")
		}
	}
	b.write(strings.Join(args, ","))
	b.writeByte(')')

	if n.ReturnAnnotation != nil {
		b.writeByte(':')
		b.write((*n.ReturnAnnotation).Minify())
	}

	b.write(minifyBlockBody(n.Body))
	b.write("end")

	return b.String()
}

// minifyParamsAndReturn renders the parameter list and return type of a function type, so callers (e.g. declared methods) can splice a name in.
func minifyParamsAndReturn(n AstTypeFunction) (params string, ret string) {
	parts := make([]string, 0, len(n.ArgTypes.Types)+1)

	for i, argType := range n.ArgTypes.Types {
		p := argType.Minify()
		if i < len(n.ArgNames) && n.ArgNames[i] != nil && n.ArgNames[i].Name != "" {
			p = n.ArgNames[i].Name + ":" + p
		}
		parts = append(parts, p)
	}

	if n.ArgTypes.TailType != nil {
		parts = append(parts, (*n.ArgTypes.TailType).Minify())
	}

	return strings.Join(parts, ","), n.ReturnTypes.Minify()
}

// minifyTableItem renders a table item, allowing the value of a trailing list item to expand into multiple values.
func minifyTableItem(item AstExprTableItem, expand bool) string {
	if item.Key == nil {
		return minifyExpr(item.Value, 0, expand)
	}

	value := minifyExpr(item.Value, 0, expand)

	switch item.Kind {
	case Record:
		key := ""
		switch k := (*item.Key).(type) {
		case AstExprConstantString:
			key = k.Value
		case *AstExprConstantString:
			key = k.Value
		}

		var b minifyBuilder
		b.write(key)
		b.writeByte('=')
		b.write(value)
		return b.String()
	case General:
		if key, ok := sourceRecordKey(*item.Key); ok {
			var b minifyBuilder
			b.write(key)
			b.writeByte('=')
			b.write(value)
			return b.String()
		}

		var b minifyBuilder
		b.writeByte('[')
		b.write(minifyExprPrec(*item.Key, 0))
		b.writeByte(']')
		b.writeByte('=')
		b.write(value)
		return b.String()
	}

	return value
}

// minifyUnionIntersection renders union and intersection types, dropping the marker types and collapsing the optional marker.
func minifyUnionIntersection(types []AstType, op string) string {
	var nonOptional []AstType
	seenOptional := false

	for _, typ := range types {
		if _, ok := typ.(AstTypeOptional); ok {
			seenOptional = true
			continue
		}
		nonOptional = append(nonOptional, typ)
	}

	if len(nonOptional) == 0 {
		return "?"
	}

	parts := make([]string, len(nonOptional))
	for i, typ := range nonOptional {
		parts[i] = typ.Minify()
	}

	if len(nonOptional) == 1 {
		if seenOptional {
			return parts[0] + "?"
		}
		return parts[0]
	}

	var b minifyBuilder
	for i, part := range parts {
		if i > 0 {
			b.write(op)
		}
		b.write(part)
	}

	if seenOptional {
		return "(" + b.String() + ")?"
	}
	return b.String()
}

// -------------------------------------------------------------------------------- -- EXPRESSIONS --------------------------------------------------------------------------------

func (n AstExprBinary) Minify() string {
	op := BinaryOp(n.Op)
	priorities := BinaryPriority[op]

	leftMin := priorities[0]
	rightMin := priorities[1]

	if priorities[1] < priorities[0] {
		leftMin++
	} else {
		rightMin++
	}

	var b minifyBuilder
	b.write(minifyExpr(n.Left, leftMin, false))
	b.write(sourceBinaryOp(op))
	b.write(minifyExpr(n.Right, rightMin, false))

	return b.String()
}

func (n AstExprCall) Minify() string {
	var b minifyBuilder

	b.write(minifyExprPostfix(n.Func))

	if n.TypeArguments != nil && len(*n.TypeArguments) > 0 {
		b.write("<<")
		b.write(minifyTypeOrPackList(*n.TypeArguments))
		b.write(">>")
	}

	if len(n.Args) == 1 && sourceCallSugar(n.Args[0]) {
		b.write(minifyExpr(n.Args[0], 0, true))
		return b.String()
	}

	b.writeByte('(')
	b.write(minifyExprList(n.Args))
	b.writeByte(')')

	return b.String()
}

func (n AstExprConstantBool) Minify() string {
	if n.Value {
		return "true"
	}
	return "false"
}

func (n AstExprConstantNil) Minify() string {
	return "nil"
}

func (n AstExprConstantNumber) Minify() string {
	return sourceNumber(n.Value)
}

func (n AstExprConstantInteger) Minify() string {
	if n.Value >= 0x1000 {
		if hex, ok := cleanHex(uint64(n.Value)); ok {
			return hex + "i"
		}
	}
	return fmt.Sprintf("%di", n.Value)
}

func (n AstExprConstantString) Minify() string {
	return sourceString(n.Value, n.QuoteStyle)
}

func (n AstExprError) Minify() string {
	if len(n.Expressions) > 0 {
		return minifyExprPrec(n.Expressions[0], 0)
	}
	return "nil"
}

func (n AstExprFunction) Minify() string {
	var b minifyBuilder

	if attrs := minifyAttrs(n.Attributes); attrs != "" {
		b.write(attrs)
	}
	b.write("function")
	b.write(minifyFunctionRest(n))

	return b.String()
}

func (n AstExprGlobal) Minify() string {
	return n.Name
}

func (n AstExprGroup) Minify() string {
	return minifyParens(n.Expr.Minify())
}

func (n AstExprIfElse) Minify() string {
	return n.minifyIf("if")
}

// minifyIf renders an if-then-else expression with the given leading keyword, so nested else-if expressions can be rendered as `elseif` chains.
func (n AstExprIfElse) minifyIf(keyword string) string {
	var b minifyBuilder

	b.write(keyword)
	if n.ConditionLocal != nil {
		if n.ConditionIsConst {
			b.write("const")
		} else {
			b.write("local")
		}
		b.write(n.ConditionLocal.Name)
		b.writeByte('=')
	}
	b.write(minifyExprPrec(n.Condition, 0))
	b.write("then")
	b.write(minifyExprOrNil(n.TrueExpr))

	if falseExpr, ok := asIfElseExpr(sourceUnwrapGroup(n.FalseExpr)); ok {
		b.write(falseExpr.minifyIf("elseif"))
	} else {
		b.write("else")
		b.write(minifyExprOrNil(n.FalseExpr))
	}

	return b.String()
}

func (n AstExprIndexExpr) Minify() string {
	var b minifyBuilder
	b.write(minifyExprPostfix(n.Expr))
	b.writeByte('[')
	b.write(minifyExprPrec(n.Index, 0))
	b.writeByte(']')
	return b.String()
}

func (n AstExprIndexName) Minify() string {
	var b minifyBuilder
	b.write(minifyExprPostfix(n.Expr))
	b.writeByte(byte(n.Op))
	b.write(n.Index)
	return b.String()
}

func (n AstExprInterpString) Minify() string {
	var b minifyBuilder

	b.writeByte('`')
	for i, str := range n.Strings {
		b.write(sourceInterpStringPart(str))
		if i < len(n.Expressions) {
			b.writeByte('{')
			b.write(minifyExprPrec(n.Expressions[i], 0))
			b.writeByte('}')
		}
	}
	b.writeByte('`')

	return b.String()
}

func (n AstExprInstantiate) Minify() string {
	var b minifyBuilder
	b.write(minifyExprPostfix(n.Expr))
	b.write("<<")
	b.write(minifyTypeOrPackList(n.TypeArguments))
	b.write(">>")
	return b.String()
}

func (n AstExprLocal) Minify() string {
	return n.Local.Name
}

func (n AstExprTable) Minify() string {
	if len(n.Items) == 0 {
		return "{}"
	}

	parts := make([]string, len(n.Items))
	for i, item := range n.Items {
		parts[i] = minifyTableItem(item, i == len(n.Items)-1 && item.Kind == List)
	}

	var b minifyBuilder
	b.writeByte('{')
	b.write(strings.Join(parts, ","))
	b.writeByte('}')

	return b.String()
}

func (n AstExprTableItem) Minify() string {
	return minifyTableItem(n, false)
}

func (n AstExprTypeAssertion) Minify() string {
	var b minifyBuilder
	b.write(minifyExpr(n.Expr, precPrimary, false))
	b.write("::")
	b.write(n.Annotation.Minify())
	return b.String()
}

func (n AstExprVarargs) Minify() string {
	return "..."
}

func (n AstExprUnary) Minify() string {
	var b minifyBuilder
	b.write(minifyUnaryOp(n.Op))
	b.write(minifyExpr(n.Expr, precUnaryOperand, false))
	return b.String()
}

// minifyUnaryOp renders a unary operator without a trailing space, which the builder inserts when needed.
func minifyUnaryOp(op UnaryOp) string {
	switch op {
	case UnaryOp_Not:
		return "not"
	case UnaryOp_Minus:
		return "-"
	case UnaryOp_Len:
		return "#"
	}
	return "?"
}

// minifyExprOrNil renders expr, or nil when it is missing.
func minifyExprOrNil(expr AstExpr) string {
	if expr == nil {
		return "nil"
	}
	return minifyExprPrec(expr, 0)
}

// -------------------------------------------------------------------------------- -- LOCALS, GENERICS AND ATTRIBUTES --------------------------------------------------------------------------------

func (n AstAttr) Minify() string {
	name := n.Type
	if n.Name != nil {
		name = *n.Name
	}

	if len(n.Args) > 0 {
		var b minifyBuilder
		b.write("@[")
		b.write(name)
		b.writeByte('(')
		b.write(minifyExprList(n.Args))
		b.writeByte(')')
		b.writeByte(']')
		return b.String()
	}

	return "@" + name
}

func (n AstArgumentName) Minify() string {
	return n.Name
}

func (n AstGenericType) Minify() string {
	if n.DefaultValue == nil {
		return n.Name
	}
	return n.Name + "=" + (*n.DefaultValue).Minify()
}

func (n AstGenericTypePack) Minify() string {
	s := n.Name + "..."
	if n.DefaultValue != nil {
		s += "=" + (*n.DefaultValue).Minify()
	}
	return s
}

func (n AstLocal) Minify() string {
	if n.Annotation == nil {
		return n.Name
	}

	var b minifyBuilder
	b.write(n.Name)
	b.writeByte(':')
	b.write(n.Annotation.Minify())
	return b.String()
}

// -------------------------------------------------------------------------------- -- STATEMENTS --------------------------------------------------------------------------------

func (n AstStatAssign) Minify() string {
	var b minifyBuilder
	b.write(minifyExprList(n.Vars))
	b.writeByte('=')
	b.write(minifyExprList(n.Values))
	return b.String()
}

func (n AstStatBlock) Minify() string {
	if !n.HasEnd {
		return minifyRoot(n)
	}

	var b minifyBuilder
	b.write("do")
	b.write(minifyBlockBody(n))
	b.write("end")
	return b.String()
}

func (n AstStatBreak) Minify() string {
	return "break"
}

func (n AstStatCompoundAssign) Minify() string {
	var b minifyBuilder
	b.write(minifyExprPrec(n.Var, 0))
	b.write(sourceBinaryOp(n.Op))
	b.writeByte('=')
	b.write(minifyExprPrec(n.Value, 0))
	return b.String()
}

func (n AstStatContinue) Minify() string {
	return "continue"
}

func (n AstStatDeclareFunction) Minify() string {
	var b minifyBuilder

	if attrs := minifyAttrs(n.Attributes); attrs != "" {
		b.write(attrs)
	}
	b.write("declare")
	b.write("function")
	b.write(n.Name)

	if generics := minifyGenerics(n.Generics, n.GenericPacks); generics != "" {
		b.writeByte('<')
		b.write(generics)
		b.writeByte('>')
	}

	b.writeByte('(')
	b.write(minifyDeclareParams(n.Params, n.ParamNames, n.Vararg))
	b.writeByte(')')

	if n.RetTypes != nil {
		b.writeByte(':')
		b.write(n.RetTypes.Minify())
	}

	return b.String()
}

func (n AstStatDeclareGlobal) Minify() string {
	var b minifyBuilder
	b.write("declare")
	b.write(n.Name)
	b.writeByte(':')
	b.write(n.Type.Minify())
	return b.String()
}

func (n AstStatDeclareExternType) Minify() string {
	var b minifyBuilder

	b.write("declare")
	b.write("extern")
	b.write("type")
	b.write(n.Name)
	if n.SuperName != nil {
		b.write("extends")
		b.write(*n.SuperName)
	}
	b.write("with")

	parts := make([]string, 0, len(n.Props)+1)
	for _, prop := range n.Props {
		parts = append(parts, prop.Minify())
	}
	if n.Indexer != nil {
		parts = append(parts, n.Indexer.Minify())
	}
	if props := strings.Join(parts, " "); props != "" {
		b.write(props)
	}

	b.write("end")

	return b.String()
}

func (n AstDeclaredExternTypeProperty) Minify() string {
	name := n.Name.Value

	if !n.IsMethod {
		var b minifyBuilder
		b.write(name)
		b.writeByte(':')
		b.write(n.Ty.Minify())
		return b.String()
	}

	if fn, ok := n.Ty.(AstTypeFunction); ok {
		params, ret := minifyParamsAndReturn(fn)
		if params == "" {
			params = "self"
		} else {
			params = "self," + params
		}

		var b minifyBuilder
		b.write("function")
		b.write(name)
		b.writeByte('(')
		b.write(params)
		b.writeByte(')')
		b.writeByte(':')
		b.write(ret)
		return b.String()
	}

	var b minifyBuilder
	b.write("function")
	b.write(name)
	b.writeByte(':')
	b.write(n.Ty.Minify())
	return b.String()
}

func (n AstStatError) Minify() string {
	if len(n.Statements) > 0 {
		return minifyStatList(n.Statements)
	}
	if len(n.Expressions) > 0 {
		return minifyExprList(n.Expressions)
	}
	return "nil"
}

func (n AstStatExpr) Minify() string {
	return minifyExprPrec(n.Expr, 0)
}

func (n AstStatFor) Minify() string {
	var b minifyBuilder

	b.write("for")
	if n.Var != nil {
		b.write(n.Var.Minify())
	}
	b.writeByte('=')
	b.write(minifyExprPrec(n.From, 0))
	b.writeByte(',')
	b.write(minifyExprPrec(n.To, 0))
	if n.Step != nil {
		b.writeByte(',')
		b.write(minifyExprPrec(n.Step, 0))
	}
	b.write("do")
	if n.Body != nil {
		b.write(minifyBlockBody(*n.Body))
	}
	b.write("end")

	return b.String()
}

func (n AstStatForIn) Minify() string {
	var b minifyBuilder

	b.write("for")
	vars := make([]string, len(n.Vars))
	for i, v := range n.Vars {
		if v != nil {
			vars[i] = v.Minify()
		}
	}
	b.write(strings.Join(vars, ","))
	b.write("in")
	b.write(minifyExprList(n.Values))
	b.write("do")
	if n.Body != nil {
		b.write(minifyBlockBody(*n.Body))
	}
	b.write("end")

	return b.String()
}

func (n AstStatFunction) Minify() string {
	var b minifyBuilder

	if attrs := minifyAttrs(n.Func.Attributes); attrs != "" {
		b.write(attrs)
	}
	b.write("function")
	b.write(minifyExprPrec(n.Name, precPrimary))
	b.write(minifyFunctionRest(n.Func))

	return b.String()
}

func (n AstStatIf) Minify() string {
	return n.minifyIf("if")
}

// minifyIf renders an if statement with the given leading keyword, so nested else-if statements can be rendered as `elseif` chains.
func (n AstStatIf) minifyIf(keyword string) string {
	var b minifyBuilder

	b.write(keyword)
	if n.ConditionLocal != nil {
		if n.ConditionIsConst {
			b.write("const")
		} else {
			b.write("local")
		}
		b.write(n.ConditionLocal.Name)
		b.writeByte('=')
	}
	b.write(minifyExprPrec(n.Condition, 0))
	b.write("then")
	b.write(minifyBlockBody(n.ThenBody))

	switch {
	case n.ElseBody == nil:
		b.write("end")
	case isIfStat(n.ElseBody):
		// the nested else-if writes its own `end`
		elseIf, _ := asIfStat(n.ElseBody)
		b.write(elseIf.minifyIf("elseif"))
	default:
		b.write("else")
		if elseBlock, ok := asBlock(n.ElseBody); ok {
			b.write(minifyBlockBody(*elseBlock))
		} else {
			b.write(n.ElseBody.Minify())
		}
		b.write("end")
	}

	return b.String()
}

func (n AstStatLocal) Minify() string {
	var b minifyBuilder

	switch {
	case n.IsExported && n.IsConst:
		b.write("export")
		b.write("const")
	case n.IsExported:
		b.write("export")
		b.write("local")
	case n.IsConst:
		b.write("const")
	default:
		b.write("local")
	}

	vars := make([]string, len(n.Vars))
	for i := range n.Vars {
		vars[i] = n.Vars[i].Minify()
	}
	b.write(strings.Join(vars, ","))

	if len(n.Values) > 0 {
		b.writeByte('=')
		b.write(minifyExprList(n.Values))
	}

	return b.String()
}

func (n AstStatLocalFunction) Minify() string {
	var b minifyBuilder

	if attrs := minifyAttrs(n.Func.Attributes); attrs != "" {
		b.write(attrs)
	}

	switch {
	case n.Name.IsExported:
		b.write("export")
		b.write("function")
	case n.IsConst:
		b.write("const")
		b.write("function")
	default:
		b.write("local")
		b.write("function")
	}

	b.write(n.Name.Name)
	b.write(minifyFunctionRest(n.Func))

	return b.String()
}

func (n AstStatRepeat) Minify() string {
	var b minifyBuilder

	b.write("repeat")
	if n.Body != nil {
		b.write(minifyBlockBody(*n.Body))
	}
	b.write("until")
	b.write(minifyExprPrec(n.Condition, 0))

	return b.String()
}

func (n AstStatReturn) Minify() string {
	if len(n.List) == 0 {
		return "return"
	}

	var b minifyBuilder
	b.write("return")
	b.write(minifyExprList(n.List))
	return b.String()
}

func (n AstStatTypeAlias) Minify() string {
	var b minifyBuilder

	if n.Exported {
		b.write("export")
	}
	b.write("type")
	b.write(n.Name)

	if generics := minifyGenerics(n.Generics, n.GenericPacks); generics != "" {
		b.writeByte('<')
		b.write(generics)
		b.writeByte('>')
	}

	b.writeByte('=')
	b.write(n.Type.Minify())

	return b.String()
}

func (n AstStatTypeFunction) Minify() string {
	var b minifyBuilder

	if n.Exported {
		b.write("export")
	}
	b.write("type")
	b.write("function")
	b.write(n.Name)
	b.write(minifyFunctionRest(n.Body))

	return b.String()
}

func (n AstStatWhile) Minify() string {
	var b minifyBuilder

	b.write("while")
	b.write(minifyExprPrec(n.Condition, 0))
	b.write("do")
	if n.Body != nil {
		b.write(minifyBlockBody(*n.Body))
	}
	b.write("end")

	return b.String()
}

// -------------------------------------------------------------------------------- -- TYPES --------------------------------------------------------------------------------

func (n AstTableIndexer) Minify() string {
	var b minifyBuilder

	switch n.Access {
	case "Read":
		b.write("read")
	case "Write":
		b.write("write")
	}

	b.writeByte('[')
	b.write(n.IndexType.Minify())
	b.writeByte(']')
	b.writeByte(':')
	b.write(n.ResultType.Minify())

	return b.String()
}

func (n AstTableProp) Minify() string {
	var b minifyBuilder

	switch n.Access {
	case "Read":
		b.write("read")
	case "Write":
		b.write("write")
	}

	b.write(n.Name.Value)
	b.writeByte(':')
	b.write(n.Type.Minify())

	return b.String()
}

func (n AstTypeError) Minify() string {
	if len(n.Types) > 0 {
		return n.Types[0].Minify()
	}
	return "any"
}

func (n AstTypeFunction) Minify() string {
	params, ret := minifyParamsAndReturn(n)

	var b minifyBuilder
	if generics := minifyGenerics(n.Generics, n.GenericPacks); generics != "" {
		b.writeByte('<')
		b.write(generics)
		b.writeByte('>')
	}
	b.writeByte('(')
	b.write(params)
	b.writeByte(')')
	b.write("->")
	b.write(ret)

	return b.String()
}

func (n AstTypeGroup) Minify() string {
	return minifyParens(n.Type.Minify())
}

func (n AstTypeIntersection) Minify() string {
	return minifyUnionIntersection(n.Types, "&")
}

func (n AstTypeList) Minify() string {
	parts := make([]string, 0, len(n.Types)+1)
	for _, typ := range n.Types {
		parts = append(parts, typ.Minify())
	}
	if n.TailType != nil {
		parts = append(parts, (*n.TailType).Minify())
	}
	return strings.Join(parts, ",")
}

func (n AstTypeOptional) Minify() string {
	return "?"
}

func (n AstTypeOrPack) Minify() string {
	if n.Type != nil {
		return (*n.Type).Minify()
	}
	if n.Pack != nil {
		return (*n.Pack).Minify()
	}
	return ""
}

func (n AstTypePackExplicit) Minify() string {
	list := n.Types.Minify()
	if n.TailType != nil {
		if list != "" {
			list += ","
		}
		list += (*n.TailType).Minify()
	}

	single := len(n.Types.Types) == 1 && n.Types.TailType == nil && n.TailType == nil
	if single {
		return list
	}
	return "(" + list + ")"
}

func (n AstTypePackGeneric) Minify() string {
	return n.GenericName + "..."
}

func (n AstTypePackVariadic) Minify() string {
	return "..." + n.VariadicType.Minify()
}

func (n AstTypeReference) Minify() string {
	var b minifyBuilder

	if n.Prefix != nil {
		b.write(*n.Prefix)
		b.writeByte('.')
	}
	b.write(n.Name)

	if n.HasParameterList || len(n.Parameters) > 0 {
		b.writeByte('<')
		b.write(minifyTypeOrPackList(n.Parameters))
		b.writeByte('>')
	}

	return b.String()
}

func (n AstTypeSingletonBool) Minify() string {
	if n.Value {
		return "true"
	}
	return "false"
}

func (n AstTypeSingletonString) Minify() string {
	return sourceString(n.Value, QuoteStyle_QuotedSimple)
}

func (n AstTypeTable) Minify() string {
	if len(n.Props) == 0 && n.Indexer != nil {
		if ref, ok := n.Indexer.IndexType.(AstTypeReference); ok && ref.Prefix == nil && ref.Name == "number" && len(ref.Parameters) == 0 {
			var b minifyBuilder
			b.writeByte('{')
			b.write(n.Indexer.ResultType.Minify())
			b.writeByte('}')
			return b.String()
		}
	}

	parts := make([]string, 0, len(n.Props)+1)
	if n.Indexer != nil {
		parts = append(parts, n.Indexer.Minify())
	}
	for _, prop := range n.Props {
		parts = append(parts, prop.Minify())
	}

	if len(parts) == 0 {
		return "{}"
	}

	var b minifyBuilder
	b.writeByte('{')
	b.write(strings.Join(parts, ","))
	b.writeByte('}')
	return b.String()
}

func (n AstTypeTypeof) Minify() string {
	var b minifyBuilder
	b.write("typeof")
	b.writeByte('(')
	b.write(minifyExprPrec(n.Expr, 0))
	b.writeByte(')')
	return b.String()
}

func (n AstTypeUnion) Minify() string {
	return minifyUnionIntersection(n.Types, "|")
}
