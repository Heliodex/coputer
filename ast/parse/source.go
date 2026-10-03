package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/Heliodex/coputer/ast/lex"
)

// This file implements Source() for every AST node. Each Source() renders the node as formatted Luau code starting at column zero, delegating to the Source() methods of its children and re-indenting their output where needed.
//
// Statements/expressions that span multiple lines indent their own contents with tabs relative to their first line, so a parent can embed them by indenting every line once more with sourceIndent.

// sourceIndent prefixes every non-empty line of s with one tab per level.
func sourceIndent(s string, levels int) string {
	if s == "" || levels <= 0 {
		return s
	}

	pad := strings.Repeat("\t", levels)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		lines[i] = pad + line
	}

	return strings.Join(lines, "\n")
}

func leadingWhitespace(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return s[:i]
}

func commonLeadingWhitespace(a, b string) string {
	n := min(len(a), len(b))
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return a[:i]
}

// sourceLongString renders a long (`[[...]]`) string with enough equals signs to avoid an accidental terminator.
func sourceLongString(value string) string {
	for eq := 0; ; eq++ {
		eqs := strings.Repeat("=", eq)
		strend := "]" + eqs + "]"
		if !strings.Contains(value, strend) {
			return "[" + eqs + "[" + value + strend
		}
	}
}

// normalizeCommentContent removes the common leading indentation from the continuation lines of a block comment. Without this, re-indenting a comment when rendering would keep adding whitespace every time the output is parsed and rendered again.
func normalizeCommentContent(content string) string {
	if !strings.Contains(content, "\n") {
		return content
	}

	lines := strings.Split(content, "\n")

	common := ""
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}

		indent := leadingWhitespace(lines[i])
		if common == "" {
			common = indent
			continue
		}
		common = commonLeadingWhitespace(common, indent)
	}

	if common == "" {
		return content
	}

	for i := 1; i < len(lines); i++ {
		lines[i] = strings.TrimPrefix(lines[i], common)
	}

	return strings.Join(lines, "\n")
}

// sourceComment renders a comment.
func sourceComment(comment Comment) string {
	if comment.Type == lex.BlockComment || comment.Type == lex.BrokenComment {
		return "--" + sourceLongString(normalizeCommentContent(comment.Content))
	}
	return "--" + comment.Content
}

// sourceString renders a string literal in the requested quote style, falling back to a safer representation when needed.
func sourceString(value string, style QuoteStyle) string {
	// long strings are avoided when they contain newlines: a multi-line literal can't be re-indented without changing its value. They can't escape bytes either, so invalid UTF-8 falls back to a quoted string.
	if style == QuoteStyle_QuotedRaw && !strings.ContainsAny(value, "\r\n") && utf8.ValidString(value) {
		return sourceLongString(value)
	}

	quote := byte('"')
	if style == QuoteStyle_QuotedSingle {
		quote = '\''
	}

	// if the value contains the preferred quote but not the other one, switch
	other := byte('"')
	if quote == '"' {
		other = '\''
	}
	if strings.IndexByte(value, quote) >= 0 && strings.IndexByte(value, other) < 0 {
		quote = other
	}

	var b strings.Builder
	b.WriteByte(quote)
	for i := 0; i < len(value); {
		ch := value[i]

		// valid UTF-8 sequences are kept as-is, invalid bytes are escaped
		if ch >= 0x80 {
			r, size := utf8.DecodeRuneInString(value[i:])
			if r == utf8.RuneError && size <= 1 {
				fmt.Fprintf(&b, "\\%03d", ch)
				i++
				continue
			}

			b.WriteString(value[i : i+size])
			i += size
			continue
		}

		i++
		switch ch {
		case '\\':
			b.WriteString(`\\`)
		case quote:
			b.WriteByte('\\')
			b.WriteByte(quote)
		case '\a':
			b.WriteString(`\a`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\v':
			b.WriteString(`\v`)
		default:
			if ch < 0x20 || ch == 0x7f {
				// pad to three digits so a following digit isn't absorbed
				fmt.Fprintf(&b, "\\%03d", ch)
			} else {
				b.WriteByte(ch)
			}
		}
	}
	b.WriteByte(quote)

	return b.String()
}

// sourceNumber renders a numeric literal.
func sourceNumber(value float64) string {
	if math.IsInf(value, 1) {
		return "math.huge"
	}
	if math.IsInf(value, -1) {
		return "-math.huge"
	}

	exp := fmt.Sprintf("%g", value)
	exp = strings.Replace(exp, "e+", "e", 1)

	for strings.Contains(exp, "e0") {
		exp = strings.Replace(exp, "e0", "e", 1)
	}
	for strings.Contains(exp, "e-0") {
		exp = strings.Replace(exp, "e-0", "e-", 1)
	}

	// prefer the plain decimal form when it isn't longer than the exponent form
	plain := strconv.FormatFloat(value, 'f', -1, 64)
	if len(plain) <= len(exp) {
		return plain
	}

	return exp
}

// sourceBinaryOp renders a binary operator.
func sourceBinaryOp(op BinaryOp) string {
	switch op {
	case BinaryOp_Add:
		return "+"
	case BinaryOp_Sub:
		return "-"
	case BinaryOp_Mul:
		return "*"
	case BinaryOp_Div:
		return "/"
	case BinaryOp_FloorDiv:
		return "//"
	case BinaryOp_Mod:
		return "%"
	case BinaryOp_Pow:
		return "^"
	case BinaryOp_Concat:
		return ".."
	case BinaryOp_CompareNe:
		return "~="
	case BinaryOp_CompareEq:
		return "=="
	case BinaryOp_CompareLt:
		return "<"
	case BinaryOp_CompareLe:
		return "<="
	case BinaryOp_CompareGt:
		return ">"
	case BinaryOp_CompareGe:
		return ">="
	case BinaryOp_And:
		return "and"
	case BinaryOp_Or:
		return "or"
	}
	return "?"
}

// sourceUnaryOp renders a unary operator.
func sourceUnaryOp(op UnaryOp) string {
	switch op {
	case UnaryOp_Not:
		return "not "
	case UnaryOp_Minus:
		return "-"
	case UnaryOp_Len:
		return "#"
	}
	return "?"
}

func sourceHasSemicolon(stat AstStat) bool {
	switch s := stat.(type) {
	case *AstStatAssign: // more uncollapsible type switches yayyy
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatBlock:
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatBreak:
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatCompoundAssign:
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatContinue:
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatDeclareFunction:
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatDeclareGlobal:
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatDeclareExternType:
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatError: // i
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatExpr: // could
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatFor: // use
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatForIn: // an
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatFunction: // interface
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatIf: // for
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatLocal: // this
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatLocalFunction: // but
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatRepeat: // can't
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatReturn: // be
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatTypeAlias: // arsed
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatTypeFunction:
		return s.HasSemicolon != nil && *s.HasSemicolon
	case *AstStatWhile:
		return s.HasSemicolon != nil && *s.HasSemicolon
	}
	return false
}

// sourceStat renders a statement, preserving a trailing semicolon when the source had one (semicolons can be semantically significant).
func sourceStat(stat AstStat) string {
	s := stat.Source()
	if sourceHasSemicolon(stat) {
		s += ";"
	}
	return s
}

// sourceStatList renders statements separated by newlines, interleaving the block's comments in their original positions: leading comments stay above the statement they precede and trailing comments stay on the same line.
func sourceStatList(stats []AstStat, comments []Comment) string {
	var lines []string
	ci := 0

	for i, stat := range stats {
		statLoc := stat.GetLocation()

		// comments that start at or before this statement go on their own line
		for ci < len(comments) && !comments[ci].Location.Begin.After(statLoc.Begin) {
			lines = append(lines, sourceComment(comments[ci]))
			ci++
		}

		src := sourceStat(stat)

		// comments on the same line as the end of the statement stay there
		for ci < len(comments) &&
			comments[ci].Location.Begin.Line == statLoc.End.Line &&
			(i+1 >= len(stats) || comments[ci].Location.Begin.Before(stats[i+1].GetLocation().Begin)) {
			src += " " + sourceComment(comments[ci])
			ci++
		}

		lines = append(lines, src)
	}

	// comments after the last statement
	for ci < len(comments) {
		lines = append(lines, sourceComment(comments[ci]))
		ci++
	}

	return strings.Join(lines, "\n")
}

// sourceBlockBody renders a block's statements indented one level, preceded by a newline. Empty blocks render as an empty string.
func sourceBlockBody(body AstStatBlock) string {
	s := sourceStatList(body.Body, body.Comments)
	if s == "" {
		return ""
	}
	return "\n" + sourceIndent(s, 1)
}

// sourceEndChain reports whether line consists solely of `end` keywords, e.g.
// "end" or "end end end".
func sourceEndChain(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}

	for _, field := range fields {
		if field != "end" {
			return false
		}
	}
	return true
}

// appendEnd appends an `end` to a block body, collapsing it onto the previous line when that line already ends a nested block. The collapsed line is unindented one level, so a chain of ends lines up with the outermost block.
func appendEnd(body string) string {
	if body == "" {
		return "\nend"
	}

	lines := strings.Split(body, "\n")
	last := len(lines) - 1

	if sourceEndChain(lines[last]) {
		lines[last] = strings.TrimPrefix(lines[last], "\t") + " end"
		return strings.Join(lines, "\n")
	}

	return body + "\nend"
}

// Expression precedence levels, mirroring the parser's binding powers. A group can be dropped when its contents bind at least as tightly as the context.
const (
	precIfElse       = 0
	precUnaryOperand = 8
	precUnary        = 9
	precAssert       = 11
	precFunction     = 11
	precPrimary      = 12
)

// sourceExprPrecedence returns the binding power of an expression, used to decide whether a parenthesised group can be dropped.
func sourceExprPrecedence(expr AstExpr) int {
	switch e := expr.(type) {
	case AstExprBinary:
		return BinaryPriority[BinaryOp(e.Op)][0]
	case AstExprUnary:
		return precUnary
	case AstExprIfElse:
		return precIfElse
	case AstExprTypeAssertion:
		return precAssert
	case AstExprFunction, *AstExprFunction:
		return precFunction
	}
	return precPrimary
}

// sourceExprExpands reports whether a bare expression can expand to multiple values, i.e. a call or `...`.
func sourceExprExpands(expr AstExpr) bool {
	switch expr.(type) {
	case AstExprCall, *AstExprCall, AstExprVarargs, *AstExprVarargs:
		return true
	}
	return false
}

// sourceParens wraps an expression in parentheses, indenting multi-line contents onto their own lines.
func sourceParens(s string) string {
	if strings.Contains(s, "\n") {
		return "(\n" + sourceIndent(s, 1) + "\n)"
	}
	return "(" + s + ")"
}

// sourceUnwrapGroup removes any parentheses groups around an expression.
func sourceUnwrapGroup(expr AstExpr) AstExpr {
	for {
		switch e := expr.(type) {
		case AstExprGroup:
			expr = e.Expr
		case *AstExprGroup:
			expr = e.Expr
		default:
			return expr
		}
	}
}

// sourceExpr renders expr in a context requiring at least minPrec. When expand is true the expression sits where a bare call or `...` would expand to multiple values, so parentheses around one are kept. Groups are dropped when they affect neither operator precedence nor the number of values.
func sourceExpr(expr AstExpr, minPrec int, expand bool) string {
	var group *AstExprGroup
	switch e := expr.(type) {
	case AstExprGroup:
		group = &e
	case *AstExprGroup:
		group = e
	}

	if group != nil {
		inner := group.Expr

		// a group around a call or `...` truncates it to a single value
		if expand && sourceExprExpands(inner) {
			return sourceParens(sourceExpr(inner, 0, false))
		}

		return sourceExpr(inner, minPrec, expand)
	}

	s := expr.Source()
	if sourceExprPrecedence(expr) < minPrec {
		return sourceParens(s)
	}
	return s
}

// sourceExprPrec renders expr in a single-value context requiring at least minPrec.
func sourceExprPrec(expr AstExpr, minPrec int) string {
	return sourceExpr(expr, minPrec, false)
}

// sourcePostfixBase reports whether an expression can be indexed or called without parentheses. (on lhs not rhs)
func sourcePostfixBase(expr AstExpr) bool {
	switch expr.(type) {
	case AstExprLocal, *AstExprLocal,
		AstExprGlobal, *AstExprGlobal,
		AstExprIndexName, *AstExprIndexName,
		AstExprIndexExpr, *AstExprIndexExpr,
		AstExprCall, *AstExprCall,
		AstExprInstantiate, *AstExprInstantiate:
		return true
	}
	return false
}

// sourceExprPostfix renders an expression used as the base of a postfix operation. Unlike other primary expressions, literals, tables, functions and `...` can't be indexed or called without parentheses.
func sourceExprPostfix(expr AstExpr) string {
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
			return sourceParens(sourceExprPrec(unwrapped, 0))
		}
	}

	return sourceExprPrec(expr, precPrimary)
}

// sourceExprList renders comma-separated expressions, allowing the last one to expand into multiple values.
func sourceExprList(exprs []AstExpr) string {
	parts := make([]string, len(exprs))
	for i, expr := range exprs {
		parts[i] = sourceExpr(expr, 0, i == len(exprs)-1)
	}
	return strings.Join(parts, ", ")
}

// sourceTypeOrPackList renders comma-separated types/type packs.
func sourceTypeOrPackList(items []AstTypeOrPack) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = item.Source()
	}
	return strings.Join(parts, ", ")
}

// sourceGenerics renders a generic list, e.g. `T, U..., V = string`.
func sourceGenerics(generics []AstGenericType, packs []AstGenericTypePack) string {
	parts := make([]string, 0, len(generics)+len(packs))
	for _, generic := range generics {
		parts = append(parts, generic.Source())
	}
	for _, pack := range packs {
		parts = append(parts, pack.Source())
	}
	return strings.Join(parts, ", ")
}

// sourceAttrs renders attributes followed by newlines, or "" when there are none.
func sourceAttrs(attrs []AstAttr) string {
	if len(attrs) == 0 {
		return ""
	}

	var b strings.Builder
	for _, attr := range attrs {
		b.WriteString(attr.Source())
		b.WriteByte('\n')
	}
	return b.String()
}

// sourceCallSugar reports whether a call argument can be passed without parentheses, i.e. `f "string"` or `f { table }`.
func sourceCallSugar(expr AstExpr) bool {
	switch sourceUnwrapGroup(expr).(type) {
	case AstExprConstantString, *AstExprConstantString:
		return true
	case AstExprTable, *AstExprTable:
		return true
	}
	return false
}

// sourceInterpStringPart escapes a raw segment of an interpolated string.
func sourceInterpStringPart(s string) string {
	var b strings.Builder

	for i := 0; i < len(s); {
		ch := s[i]

		// valid UTF-8 sequences are kept as-is, invalid bytes are escaped
		if ch >= 0x80 {
			r, size := utf8.DecodeRuneInString(s[i:])
			if r == utf8.RuneError && size <= 1 {
				fmt.Fprintf(&b, "\\%03d", ch)
				i++
				continue
			}

			b.WriteString(s[i : i+size])
			i += size
			continue
		}

		i++
		switch ch {
		case '\\':
			b.WriteString(`\\`)
		case '`':
			b.WriteString("\\`")
		case '{':
			b.WriteString("\\{")
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if ch < 0x20 || ch == 0x7f {
				// pad to three digits so a following digit isn't absorbed
				fmt.Fprintf(&b, "\\%03d", ch)
			} else {
				b.WriteByte(ch)
			}
		}
	}

	return b.String()
}

func asIfElseExpr(expr AstExpr) (AstExprIfElse, bool) {
	switch e := expr.(type) {
	case AstExprIfElse:
		return e, true
	case *AstExprIfElse:
		return *e, true
	}
	return AstExprIfElse{}, false
}

// sourceHasIfExpr reports whether any of the expressions is an if-then-else expression.
func sourceHasIfExpr(exprs []AstExpr) bool {
	for _, expr := range exprs {
		if _, ok := asIfElseExpr(sourceUnwrapGroup(expr)); ok {
			return true
		}
	}
	return false
}

// sourceCompoundOp returns the compound assignment operator for a binary operator, if it has one.
func sourceCompoundOp(op BinaryOp) (BinaryOp, bool) {
	switch op {
	case BinaryOp_Add, BinaryOp_Sub, BinaryOp_Mul, BinaryOp_Div,
		BinaryOp_FloorDiv, BinaryOp_Mod, BinaryOp_Pow, BinaryOp_Concat:
		return op, true
	}
	return 0, false
}

// sourcePureExpr reports whether evaluating an expression has no observable effects. Only locals and constants qualify: global reads can invoke environment metamethods and indexing can invoke table metamethods.
func sourcePureExpr(expr AstExpr) bool {
	switch sourceUnwrapGroup(expr).(type) {
	case AstExprLocal, *AstExprLocal,
		AstExprConstantNil, *AstExprConstantNil,
		AstExprConstantBool, *AstExprConstantBool,
		AstExprConstantNumber, *AstExprConstantNumber,
		AstExprConstantInteger, *AstExprConstantInteger,
		AstExprConstantString, *AstExprConstantString:
		return true
	}
	return false
}

// sourceSameLocal reports whether two local bindings are the same. Bindings share the NodeLoc of their declaration, so comparing it identifies them without touching the uncomparable Annotation field.
func sourceSameLocal(a, b AstLocal) bool {
	if a.NodeLoc == nil || b.NodeLoc == nil {
		return false
	}
	return a.NodeLoc == b.NodeLoc
}

// sourceSameExpr reports whether two expressions refer to the same value or location. Groups are ignored, so `(x)` and `x` compare equal.
func sourceSameExpr(a, b AstExpr) bool {
	a, b = sourceUnwrapGroup(a), sourceUnwrapGroup(b)

	switch av := a.(type) {
	case AstExprLocal:
		bv, ok := b.(AstExprLocal)
		return ok && sourceSameLocal(av.Local, bv.Local)
	case *AstExprLocal:
		bv, ok := b.(*AstExprLocal)
		return ok && sourceSameLocal(av.Local, bv.Local)
	case AstExprGlobal:
		bv, ok := b.(AstExprGlobal)
		return ok && av.Name == bv.Name
	case *AstExprGlobal:
		bv, ok := b.(*AstExprGlobal)
		return ok && av.Name == bv.Name
	case AstExprConstantNil:
		_, ok := b.(AstExprConstantNil)
		return ok
	case *AstExprConstantNil:
		_, ok := b.(*AstExprConstantNil)
		return ok
	case AstExprConstantBool:
		bv, ok := b.(AstExprConstantBool)
		return ok && av.Value == bv.Value
	case *AstExprConstantBool:
		bv, ok := b.(*AstExprConstantBool)
		return ok && av.Value == bv.Value
	case AstExprConstantNumber:
		bv, ok := b.(AstExprConstantNumber)
		return ok && av.Value == bv.Value
	case *AstExprConstantNumber:
		bv, ok := b.(*AstExprConstantNumber)
		return ok && av.Value == bv.Value
	case AstExprConstantInteger:
		bv, ok := b.(AstExprConstantInteger)
		return ok && av.Value == bv.Value
	case *AstExprConstantInteger:
		bv, ok := b.(*AstExprConstantInteger)
		return ok && av.Value == bv.Value
	case AstExprConstantString:
		bv, ok := b.(AstExprConstantString)
		return ok && av.Value == bv.Value
	case *AstExprConstantString:
		bv, ok := b.(*AstExprConstantString)
		return ok && av.Value == bv.Value
	case AstExprIndexName:
		bv, ok := b.(AstExprIndexName)
		return ok && av.Index == bv.Index && av.Op == bv.Op && sourceSameExpr(av.Expr, bv.Expr)
	case *AstExprIndexName:
		bv, ok := b.(*AstExprIndexName)
		return ok && av.Index == bv.Index && av.Op == bv.Op && sourceSameExpr(av.Expr, bv.Expr)
	case AstExprIndexExpr:
		bv, ok := b.(AstExprIndexExpr)
		return ok && sourceSameExpr(av.Expr, bv.Expr) && sourceSameExpr(av.Index, bv.Index)
	case *AstExprIndexExpr:
		bv, ok := b.(*AstExprIndexExpr)
		return ok && sourceSameExpr(av.Expr, bv.Expr) && sourceSameExpr(av.Index, bv.Index)
	}
	return false
}

// sourceSafeCompoundTarget reports whether a target can be used in a compound assignment without changing behaviour. `target = target op value` evaluates the target twice, while `target op= value` evaluates it once, so targets whose evaluation has observable effects must keep the plain form.
func sourceSafeCompoundTarget(target AstExpr) bool {
	switch t := sourceUnwrapGroup(target).(type) {
	case AstExprLocal, *AstExprLocal, AstExprGlobal, *AstExprGlobal:
		return true
	case AstExprIndexName:
		return sourcePureExpr(t.Expr)
	case *AstExprIndexName:
		return sourcePureExpr(t.Expr)
	case AstExprIndexExpr:
		return sourcePureExpr(t.Expr) && sourcePureExpr(t.Index)
	case *AstExprIndexExpr:
		return sourcePureExpr(t.Expr) && sourcePureExpr(t.Index)
	}
	return false
}

// sourceCompoundAssign converts `target = target op value` into a compound assignment when doing so cannot change behaviour.
func sourceCompoundAssign(target AstExpr, value AstExpr) (AstStatCompoundAssign, bool) {
	var bin AstExprBinary
	switch v := sourceUnwrapGroup(value).(type) {
	case AstExprBinary:
		bin = v
	case *AstExprBinary:
		bin = *v
	default:
		return AstStatCompoundAssign{}, false
	}

	op, ok := sourceCompoundOp(BinaryOp(bin.Op))
	if !ok || !sourceSameExpr(target, bin.Left) || !sourceSafeCompoundTarget(target) {
		return AstStatCompoundAssign{}, false
	}

	return AstStatCompoundAssign{Var: target, Op: op, Value: bin.Right}, true
}

// writeIfBranch writes an if-expression branch value, putting a multi-line value on its own indented line.
func writeIfBranch(b *strings.Builder, expr AstExpr) {
	if expr == nil {
		b.WriteString(" nil")
		return
	}

	s := sourceExprPrec(expr, 0)
	if strings.Contains(s, "\n") {
		b.WriteByte('\n')
		b.WriteString(sourceIndent(s, 1))
		return
	}

	b.WriteByte(' ')
	b.WriteString(s)
}

// writeAssignedValue writes ` = values`, putting if-expressions on their own indented lines.
func writeAssignedValue(b *strings.Builder, exprs []AstExpr) {
	value := sourceExprList(exprs)
	if sourceHasIfExpr(exprs) {
		b.WriteString(" =\n")
		b.WriteString(sourceIndent(value, 1))
		return
	}

	b.WriteString(" = ")
	b.WriteString(value)
}

// sourceVarargAnnotation renders a vararg annotation. The parser wraps parameter vararg annotations in a variadic type pack, but the `...` is already written as part of the parameter list.
func sourceVarargAnnotation(pack AstTypePack) string {
	switch p := pack.(type) {
	case AstTypePackVariadic:
		return p.VariadicType.Source()
	case *AstTypePackVariadic:
		return p.VariadicType.Source()
	}
	return pack.Source()
}

func sourceDeclareParams(params AstTypeList, names []AstArgumentName, vararg bool) string {
	parts := make([]string, 0, len(params.Types)+1)

	for i, param := range params.Types {
		p := param.Source()
		if i < len(names) && names[i].Name != "" {
			p = names[i].Name + ": " + p
		}
		parts = append(parts, p)
	}

	if vararg {
		p := "..."
		if params.TailType != nil {
			p += ": " + sourceVarargAnnotation(*params.TailType)
		}
		parts = append(parts, p)
	} else if params.TailType != nil {
		parts = append(parts, (*params.TailType).Source())
	}

	return strings.Join(parts, ", ")
}

// sourceTypeStartsInline reports whether a multi-line type should stay on the current line, e.g. a table type's opening brace.
func sourceTypeStartsInline(s string) bool {
	return strings.HasPrefix(s, "{")
}

// writeAnnotation writes `: annotation`, putting a multi-line annotation on its own indented lines, unless it opens a block such as a table type.
func writeAnnotation(b *strings.Builder, annotation string) {
	if strings.Contains(annotation, "\n") && !sourceTypeStartsInline(annotation) {
		b.WriteString(":\n")
		b.WriteString(sourceIndent(annotation, 1))
		return
	}

	b.WriteString(": ")
	b.WriteString(annotation)
}

// writeTypeAnnotation writes `: type`, putting a multi-line type on its own indented lines.
func writeTypeAnnotation(b *strings.Builder, annotation AstType) {
	if annotation == nil {
		return
	}

	writeAnnotation(b, annotation.Source())
}

// sourceIsIdentifier reports whether s is a valid Luau identifier.
func sourceIsIdentifier(s string) bool {
	if s == "" {
		return false
	}

	for i := range s {
		ch := s[i]
		switch {
		case ch == '_' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z':
		case i > 0 && ch >= '0' && ch <= '9':
		default:
			return false
		}
	}

	return true
}

// sourceRecordKey returns the bare record-key form of a string table key, if the string is a valid identifier that isn't a reserved word.
func sourceRecordKey(expr AstExpr) (string, bool) {
	var value string
	switch key := sourceUnwrapGroup(expr).(type) {
	case AstExprConstantString: // can't collapse these cuz typing
		value = key.Value
	case *AstExprConstantString:
		value = key.Value
	default:
		return "", false
	}

	if !sourceIsIdentifier(value) || lex.IsReserved(value) {
		return "", false
	}

	return value, true
}

// sourceTableItem renders a table item, allowing the value of a trailing list item to expand into multiple values.
func sourceTableItem(item AstExprTableItem, expand bool) string {
	if item.Key == nil {
		return sourceExpr(item.Value, 0, expand)
	}

	value := sourceExpr(item.Value, 0, expand)

	switch item.Kind {
	case Record:
		if key, ok := (*item.Key).(AstExprConstantString); ok {
			return key.Value + " = " + value
		}
		return sourceExprPrec(*item.Key, 0) + " = " + value
	case General:
		// a string key that's a valid identifier can use the record form
		if key, ok := sourceRecordKey(*item.Key); ok {
			return key + " = " + value
		}
		return "[" + sourceExprPrec(*item.Key, 0) + "] = " + value
	}
	return value
}

// sourceUnionIntersection renders union and intersection types. Types are always split across multiple lines, one per line, each prefixed by op.
func sourceUnionIntersection(types []AstType, op string) string {
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
		parts[i] = typ.Source()
	}

	if len(nonOptional) == 1 {
		if seenOptional {
			return parts[0] + "?"
		}
		return parts[0]
	}

	var b strings.Builder
	for i, part := range parts {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(op)
		b.WriteByte(' ')
		b.WriteString(part)
	}

	if seenOptional {
		return "(\n" + sourceIndent(b.String(), 1) + "\n)?"
	}
	return b.String()
}

func asIfStat(stat AstStat) (*AstStatIf, bool) {
	if s, ok := stat.(*AstStatIf); ok {
		return s, true
	}
	return nil, false
}

func isIfStat(stat AstStat) bool {
	_, ok := asIfStat(stat)
	return ok
}

func asBlock(stat AstStat) (*AstStatBlock, bool) {
	if s, ok := stat.(*AstStatBlock); ok {
		return s, true
	}
	return nil, false
}

// -------------------------------------------------------------------------------- -- EXPRESSIONS --------------------------------------------------------------------------------

func (n AstExprBinary) Source() string {
	op := BinaryOp(n.Op)
	priorities := BinaryPriority[op]

	leftMin := priorities[0]
	rightMin := priorities[1]

	if priorities[1] < priorities[0] {
		// right-associative: only the left operand can repeat the operator
		leftMin++
	} else {
		// left-associative: only the right operand can repeat the operator
		rightMin++
	}

	left := sourceExpr(n.Left, leftMin, false)
	right := sourceExpr(n.Right, rightMin, false)

	return left + " " + sourceBinaryOp(op) + " " + right
}

func (n AstExprCall) Source() string {
	var b strings.Builder

	b.WriteString(sourceExprPostfix(n.Func))

	if n.TypeArguments != nil && len(*n.TypeArguments) > 0 {
		b.WriteString("<<")
		b.WriteString(sourceTypeOrPackList(*n.TypeArguments))
		b.WriteString(">>")
	}

	// a call with a single string or table argument can omit its parentheses
	if len(n.Args) == 1 && sourceCallSugar(n.Args[0]) {
		b.WriteByte(' ')
		b.WriteString(sourceExpr(n.Args[0], 0, true))
		return b.String()
	}

	b.WriteString("(")
	b.WriteString(sourceExprList(n.Args))
	b.WriteString(")")

	return b.String()
}

func (n AstExprConstantBool) Source() string {
	if n.Value {
		return "true"
	}
	return "false"
}

func (n AstExprConstantNil) Source() string {
	return "nil"
}

func (n AstExprConstantNumber) Source() string {
	return sourceNumber(n.Value)
}

func (n AstExprConstantInteger) Source() string {
	return fmt.Sprintf("%di", n.Value)
}

func (n AstExprConstantString) Source() string {
	return sourceString(n.Value, n.QuoteStyle)
}

func (n AstExprError) Source() string {
	if len(n.Expressions) > 0 {
		return sourceExprPrec(n.Expressions[0], 0)
	}
	return "nil"
}

func (n AstExprFunction) Source() string {
	var b strings.Builder

	for _, attr := range n.Attributes {
		b.WriteString(attr.Source())
		b.WriteByte(' ')
	}

	b.WriteString("function")
	b.WriteString(n.sourceRest())

	return b.String()
}

// sourceRest renders a function body from the generics onwards, i.e. without the leading `function` keyword. Named function statements use this so the name can be spliced in between `function` and the parameter list.
func (n AstExprFunction) sourceRest() string {
	var b strings.Builder

	if generics := sourceGenerics(n.Generics, n.GenericPacks); generics != "" {
		b.WriteString("<")
		b.WriteString(generics)
		b.WriteString(">")
	}

	b.WriteString("(")
	args := make([]string, 0, len(n.Args)+1)
	for _, arg := range n.Args {
		args = append(args, arg.Source())
	}
	if n.Vararg {
		if n.VarargAnnotation != nil {
			args = append(args, "...: "+sourceVarargAnnotation(*n.VarargAnnotation))
		} else {
			args = append(args, "...")
		}
	}
	b.WriteString(strings.Join(args, ", "))
	b.WriteString(")")

	if n.ReturnAnnotation != nil {
		writeAnnotation(&b, (*n.ReturnAnnotation).Source())
	}

	b.WriteString(appendEnd(sourceBlockBody(n.Body)))

	return b.String()
}

func (n AstExprGlobal) Source() string {
	return n.Name
}

func (n AstExprGroup) Source() string {
	return "(" + n.Expr.Source() + ")"
}

func (n AstExprIfElse) Source() string {
	return n.sourceIf("if")
}

// sourceIf renders an if-then-else expression with the given leading keyword, so nested else-if expressions can be rendered as `elseif` chains. Each if/elseif/else section is placed on its own line.
func (n AstExprIfElse) sourceIf(keyword string) string {
	var b strings.Builder

	b.WriteString(keyword)
	b.WriteByte(' ')

	if n.ConditionLocal != nil {
		if n.ConditionIsConst {
			b.WriteString("const ")
		} else {
			b.WriteString("local ")
		}
		b.WriteString(n.ConditionLocal.Name)
		b.WriteString(" = ")
	}
	b.WriteString(sourceExprPrec(n.Condition, 0))
	b.WriteString(" then")
	writeIfBranch(&b, n.TrueExpr)

	if falseExpr, ok := asIfElseExpr(sourceUnwrapGroup(n.FalseExpr)); ok {
		// `else` + `if ...` renders as an `elseif ...` chain
		b.WriteByte('\n')
		b.WriteString(falseExpr.sourceIf("elseif"))
	} else {
		b.WriteByte('\n')
		b.WriteString("else")
		writeIfBranch(&b, n.FalseExpr)
	}

	return b.String()
}

func (n AstExprIndexExpr) Source() string {
	return sourceExprPostfix(n.Expr) + "[" + sourceExprPrec(n.Index, 0) + "]"
}

func (n AstExprIndexName) Source() string {
	return sourceExprPostfix(n.Expr) + string(n.Op) + n.Index
}

func (n AstExprInterpString) Source() string {
	var b strings.Builder

	b.WriteByte('`')
	for i, str := range n.Strings {
		b.WriteString(sourceInterpStringPart(str))
		if i < len(n.Expressions) {
			b.WriteByte('{')
			b.WriteString(sourceExprPrec(n.Expressions[i], 0))
			b.WriteByte('}')
		}
	}
	b.WriteByte('`')

	return b.String()
}

func (n AstExprInstantiate) Source() string {
	return sourceExprPostfix(n.Expr) + "<<" + sourceTypeOrPackList(n.TypeArguments) + ">>"
}

func (n AstExprLocal) Source() string {
	return n.Local.Name
}

func (n AstExprTable) Source() string {
	if len(n.Items) == 0 {
		return "{}"
	}

	parts := make([]string, len(n.Items))
	for i, item := range n.Items {
		expand := i == len(n.Items)-1 && item.Kind == List
		parts[i] = sourceTableItem(item, expand)
	}

	// tables are always split across multiple lines, one item per line
	return "{\n" + sourceIndent(strings.Join(parts, ",\n"), 1) + ",\n}"
}

func (n AstExprTableItem) Source() string {
	return sourceTableItem(n, false)
}

func (n AstExprTypeAssertion) Source() string {
	return sourceExpr(n.Expr, precPrimary, false) + " :: " + n.Annotation.Source()
}

func (n AstExprVarargs) Source() string {
	return "..."
}

func (n AstExprUnary) Source() string {
	op := sourceUnaryOp(n.Op)
	expr := sourceExpr(n.Expr, precUnaryOperand, false)

	// avoid producing a `--` comment from `- -x`
	if op == "-" && strings.HasPrefix(expr, "-") {
		return op + " " + expr
	}

	return op + expr
}

// -------------------------------------------------------------------------------- -- LOCALS, GENERICS AND ATTRIBUTES --------------------------------------------------------------------------------

func (n AstAttr) Source() string {
	name := n.Type
	if n.Name != nil {
		name = *n.Name
	}

	if len(n.Args) > 0 {
		return "@[" + name + "(" + sourceExprList(n.Args) + ")]"
	}

	return "@" + name
}

func (n AstArgumentName) Source() string {
	return n.Name
}

func (n AstGenericType) Source() string {
	s := n.Name
	if n.DefaultValue != nil {
		s += " = " + (*n.DefaultValue).Source()
	}
	return s
}

func (n AstGenericTypePack) Source() string {
	s := n.Name + "..."
	if n.DefaultValue != nil {
		s += " = " + (*n.DefaultValue).Source()
	}
	return s
}

// Source renders a binding, i.e. `name` or `name: type`.
func (n AstLocal) Source() string {
	var b strings.Builder
	b.WriteString(n.Name)
	writeTypeAnnotation(&b, n.Annotation)
	return b.String()
}

// -------------------------------------------------------------------------------- -- STATEMENTS --------------------------------------------------------------------------------

func (n AstStatAssign) Source() string {
	// `target = target op value` can be written as `target op= value` when the target is safe to evaluate once
	if len(n.Vars) == 1 && len(n.Values) == 1 {
		if compound, ok := sourceCompoundAssign(n.Vars[0], n.Values[0]); ok {
			return compound.Source()
		}
	}

	var b strings.Builder
	b.WriteString(sourceExprList(n.Vars))
	writeAssignedValue(&b, n.Values)
	return b.String()
}

func (n AstStatBlock) Source() string {
	if !n.HasEnd {
		return sourceStatList(n.Body, n.Comments)
	}

	return "do" + appendEnd(sourceBlockBody(n))
}

func (n AstStatBreak) Source() string {
	return "break"
}

func (n AstStatCompoundAssign) Source() string {
	var b strings.Builder

	b.WriteString(sourceExprPrec(n.Var, 0))
	b.WriteByte(' ')
	b.WriteString(sourceBinaryOp(n.Op))
	b.WriteString("=")

	value := sourceExprPrec(n.Value, 0)
	if _, ok := asIfElseExpr(sourceUnwrapGroup(n.Value)); ok {
		b.WriteByte('\n')
		b.WriteString(sourceIndent(value, 1))
	} else {
		b.WriteByte(' ')
		b.WriteString(value)
	}

	return b.String()
}

func (n AstStatContinue) Source() string {
	return "continue"
}

func (n AstStatDeclareFunction) Source() string {
	var b strings.Builder

	b.WriteString(sourceAttrs(n.Attributes))
	b.WriteString("declare function ")
	b.WriteString(n.Name)

	if generics := sourceGenerics(n.Generics, n.GenericPacks); generics != "" {
		b.WriteString("<")
		b.WriteString(generics)
		b.WriteString(">")
	}

	b.WriteString("(")
	b.WriteString(sourceDeclareParams(n.Params, n.ParamNames, n.Vararg))
	b.WriteString(")")

	if n.RetTypes != nil {
		writeAnnotation(&b, n.RetTypes.Source())
	}

	return b.String()
}

func (n AstStatDeclareGlobal) Source() string {
	return "declare " + n.Name + ": " + n.Type.Source()
}

func (n AstStatDeclareExternType) Source() string {
	var b strings.Builder

	b.WriteString("declare extern type ")
	b.WriteString(n.Name)
	if n.SuperName != nil {
		b.WriteString(" extends ")
		b.WriteString(*n.SuperName)
	}
	b.WriteString(" with")

	for _, prop := range n.Props {
		b.WriteByte('\n')
		b.WriteString(sourceIndent(prop.Source(), 1))
	}
	if n.Indexer != nil {
		b.WriteByte('\n')
		b.WriteString(sourceIndent(n.Indexer.Source(), 1))
	}

	b.WriteString("\nend")

	return b.String()
}

func (n AstDeclaredExternTypeProperty) Source() string {
	name := n.Name.Value

	if !n.IsMethod {
		return name + ": " + n.Ty.Source()
	}

	if fn, ok := n.Ty.(AstTypeFunction); ok {
		params, ret := fn.sourceParamsAndReturn()
		// declared methods require `self` as the unannotated first parameter
		if params == "" {
			params = "self"
		} else {
			params = "self, " + params
		}

		var b strings.Builder
		b.WriteString("function ")
		b.WriteString(name)
		b.WriteString("(")
		b.WriteString(params)
		b.WriteString(")")
		writeAnnotation(&b, ret)
		return b.String()
	}

	return "function " + name + ": " + n.Ty.Source()
}

func (n AstStatError) Source() string {
	if len(n.Statements) > 0 {
		return sourceStatList(n.Statements, nil)
	}
	if len(n.Expressions) > 0 {
		return sourceExprList(n.Expressions)
	}
	return "--[[error]]"
}

func (n AstStatExpr) Source() string {
	return sourceExprPrec(n.Expr, 0)
}

func (n AstStatFor) Source() string {
	var b strings.Builder

	b.WriteString("for ")
	if n.Var != nil {
		b.WriteString(n.Var.Source())
	}
	b.WriteString(" = ")
	b.WriteString(sourceExprPrec(n.From, 0))
	b.WriteString(", ")
	b.WriteString(sourceExprPrec(n.To, 0))
	if n.Step != nil {
		b.WriteString(", ")
		b.WriteString(sourceExprPrec(n.Step, 0))
	}
	b.WriteString(" do")

	var body string
	if n.Body != nil {
		body = sourceBlockBody(*n.Body)
	}
	b.WriteString(appendEnd(body))

	return b.String()
}

func (n AstStatForIn) Source() string {
	var b strings.Builder

	b.WriteString("for ")
	vars := make([]string, len(n.Vars))
	for i, v := range n.Vars {
		if v != nil {
			vars[i] = v.Source()
		}
	}
	b.WriteString(strings.Join(vars, ", "))
	b.WriteString(" in ")
	b.WriteString(sourceExprList(n.Values))
	b.WriteString(" do")

	var body string
	if n.Body != nil {
		body = sourceBlockBody(*n.Body)
	}
	b.WriteString(appendEnd(body))

	return b.String()
}

func (n AstStatFunction) Source() string {
	var b strings.Builder

	b.WriteString(sourceAttrs(n.Func.Attributes))
	b.WriteString("function ")
	b.WriteString(sourceExprPrec(n.Name, precPrimary))
	b.WriteString(n.Func.sourceRest())

	return b.String()
}

func (n AstStatIf) Source() string {
	return n.sourceIf("if")
}

// sourceIf renders an if statement with the given leading keyword, so nested else-if statements can be rendered as `elseif` chains.
func (n AstStatIf) sourceIf(keyword string) string {
	var b strings.Builder

	b.WriteString(keyword)
	b.WriteByte(' ')

	if n.ConditionLocal != nil {
		if n.ConditionIsConst {
			b.WriteString("const ")
		} else {
			b.WriteString("local ")
		}
		b.WriteString(n.ConditionLocal.Name)
		b.WriteString(" = ")
	}

	b.WriteString(sourceExprPrec(n.Condition, 0))
	b.WriteString(" then")

	thenBody := sourceBlockBody(n.ThenBody)

	switch {
	case n.ElseBody == nil:
		b.WriteString(appendEnd(thenBody))
	case isIfStat(n.ElseBody):
		b.WriteString(thenBody)
		elseIf, _ := asIfStat(n.ElseBody)
		b.WriteByte('\n')
		b.WriteString(elseIf.sourceIf("elseif"))
	default:
		b.WriteString(thenBody)
		b.WriteString("\nelse")

		var elseBody string
		if elseBlock, ok := asBlock(n.ElseBody); ok {
			elseBody = sourceBlockBody(*elseBlock)
		} else {
			elseBody = "\n" + sourceIndent(sourceStat(n.ElseBody), 1)
		}
		b.WriteString(appendEnd(elseBody))
	}

	return b.String()
}

func (n AstStatLocal) Source() string {
	var b strings.Builder

	switch {
	case n.IsExported && n.IsConst:
		b.WriteString("export const ")
	case n.IsExported:
		b.WriteString("export local ")
	case n.IsConst:
		b.WriteString("const ")
	default:
		b.WriteString("local ")
	}

	vars := make([]string, len(n.Vars))
	for i := range n.Vars {
		vars[i] = n.Vars[i].Source()
	}
	b.WriteString(strings.Join(vars, ", "))

	if len(n.Values) > 0 {
		writeAssignedValue(&b, n.Values)
	}

	return b.String()
}

func (n AstStatLocalFunction) Source() string {
	var b strings.Builder

	b.WriteString(sourceAttrs(n.Func.Attributes))

	switch {
	case n.Name.IsExported:
		b.WriteString("export function ")
	case n.IsConst:
		b.WriteString("const function ")
	default:
		b.WriteString("local function ")
	}

	b.WriteString(n.Name.Name)
	b.WriteString(n.Func.sourceRest())

	return b.String()
}

func (n AstStatRepeat) Source() string {
	var b strings.Builder

	b.WriteString("repeat")
	if n.Body != nil {
		b.WriteString(sourceBlockBody(*n.Body))
	}
	b.WriteString("\nuntil ")
	b.WriteString(sourceExprPrec(n.Condition, 0))

	return b.String()
}

func (n AstStatReturn) Source() string {
	if len(n.List) == 0 {
		return "return" // a "return" return
	}

	list := sourceExprList(n.List)
	if sourceHasIfExpr(n.List) {
		return "return\n" + sourceIndent(list, 1)
	}
	return "return " + list
}

func (n AstStatTypeAlias) Source() string {
	var b strings.Builder

	if n.Exported {
		b.WriteString("export ")
	}
	b.WriteString("type ")
	b.WriteString(n.Name)

	if generics := sourceGenerics(n.Generics, n.GenericPacks); generics != "" {
		b.WriteString("<")
		b.WriteString(generics)
		b.WriteString(">")
	}

	// multi-line types start on their own line, unless they open a block such as a table type
	typeSource := n.Type.Source()
	b.WriteString(" =")
	if strings.Contains(typeSource, "\n") && !sourceTypeStartsInline(typeSource) {
		b.WriteByte('\n')
		b.WriteString(sourceIndent(typeSource, 1))
	} else {
		b.WriteByte(' ')
		b.WriteString(typeSource)
	}

	return b.String()
}

func (n AstStatTypeFunction) Source() string {
	prefix := ""
	if n.Exported {
		prefix = "export "
	}
	return prefix + "type function " + n.Name + n.Body.sourceRest()
}

func (n AstStatWhile) Source() string {
	var b strings.Builder

	b.WriteString("while ")
	b.WriteString(sourceExprPrec(n.Condition, 0))
	b.WriteString(" do")

	var body string
	if n.Body != nil {
		body = sourceBlockBody(*n.Body)
	}
	b.WriteString(appendEnd(body))

	return b.String()
}

// -------------------------------------------------------------------------------- -- TYPES --------------------------------------------------------------------------------

func (n AstTableIndexer) Source() string {
	var b strings.Builder

	switch n.Access {
	case "Read":
		b.WriteString("read ")
	case "Write":
		b.WriteString("write ")
	}

	b.WriteString("[")
	b.WriteString(n.IndexType.Source())
	b.WriteString("]")
	writeTypeAnnotation(&b, n.ResultType)

	return b.String()
}

func (n AstTableProp) Source() string {
	var b strings.Builder

	switch n.Access {
	case "Read":
		b.WriteString("read ")
	case "Write":
		b.WriteString("write ")
	}

	b.WriteString(n.Name.Value)
	writeTypeAnnotation(&b, n.Type)

	return b.String()
}

func (n AstTypeError) Source() string {
	if len(n.Types) > 0 {
		return n.Types[0].Source()
	}
	return "any"
}

func (n AstTypeFunction) Source() string {
	params, ret := n.sourceParamsAndReturn()

	var b strings.Builder
	if generics := sourceGenerics(n.Generics, n.GenericPacks); generics != "" {
		b.WriteString("<")
		b.WriteString(generics)
		b.WriteString(">")
	}
	b.WriteString("(")
	b.WriteString(params)
	b.WriteString(")")

	if strings.Contains(ret, "\n") && !sourceTypeStartsInline(ret) {
		b.WriteString(" ->\n")
		b.WriteString(sourceIndent(ret, 1))
	} else {
		b.WriteString(" -> ")
		b.WriteString(ret)
	}

	return b.String()
}

// sourceParamsAndReturn renders the parameter list and return type of a function type, so callers (e.g. declared methods) can splice a name in.
func (n AstTypeFunction) sourceParamsAndReturn() (params string, ret string) {
	parts := make([]string, 0, len(n.ArgTypes.Types)+1)

	for i, argType := range n.ArgTypes.Types {
		p := argType.Source()
		if i < len(n.ArgNames) && n.ArgNames[i] != nil && n.ArgNames[i].Name != "" {
			p = n.ArgNames[i].Name + ": " + p
		}
		parts = append(parts, p)
	}

	if n.ArgTypes.TailType != nil {
		parts = append(parts, (*n.ArgTypes.TailType).Source())
	}

	return strings.Join(parts, ", "), n.ReturnTypes.Source()
}

func (n AstTypeGroup) Source() string {
	inner := n.Type.Source()
	if strings.Contains(inner, "\n") {
		return "(\n" + sourceIndent(inner, 1) + "\n)"
	}
	return "(" + inner + ")"
}

func (n AstTypeIntersection) Source() string {
	return sourceUnionIntersection(n.Types, "&")
}

func (n AstTypeList) Source() string {
	parts := make([]string, 0, len(n.Types)+1)
	for _, typ := range n.Types {
		parts = append(parts, typ.Source())
	}
	if n.TailType != nil {
		parts = append(parts, (*n.TailType).Source())
	}
	return strings.Join(parts, ", ")
}

func (n AstTypeOptional) Source() string {
	return "?"
}

func (n AstTypeOrPack) Source() string {
	if n.Type != nil {
		return (*n.Type).Source()
	}
	if n.Pack != nil {
		return (*n.Pack).Source()
	}
	return ""
}

func (n AstTypePackExplicit) Source() string {
	list := n.Types.Source()
	if n.TailType != nil {
		if list != "" {
			list += ", "
		}
		list += (*n.TailType).Source()
	}

	single := len(n.Types.Types) == 1 && n.Types.TailType == nil && n.TailType == nil
	if single {
		return list
	}
	return "(" + list + ")"
}

func (n AstTypePackGeneric) Source() string {
	return n.GenericName + "..."
}

func (n AstTypePackVariadic) Source() string {
	return "..." + n.VariadicType.Source()
}

func (n AstTypeReference) Source() string {
	var b strings.Builder

	if n.Prefix != nil {
		b.WriteString(*n.Prefix)
		b.WriteByte('.')
	}
	b.WriteString(n.Name)

	if n.HasParameterList || len(n.Parameters) > 0 {
		b.WriteString("<")
		b.WriteString(sourceTypeOrPackList(n.Parameters))
		b.WriteString(">")
	}

	return b.String()
}

func (n AstTypeSingletonBool) Source() string {
	if n.Value {
		return "true"
	}
	return "false"
}

func (n AstTypeSingletonString) Source() string {
	return sourceString(n.Value, QuoteStyle_QuotedSimple)
}

func (n AstTypeTable) Source() string {
	// array-style tables, e.g. { number }
	if len(n.Props) == 0 && n.Indexer != nil {
		if ref, ok := n.Indexer.IndexType.(AstTypeReference); ok && ref.Prefix == nil && ref.Name == "number" && len(ref.Parameters) == 0 {
			return "{ " + n.Indexer.ResultType.Source() + " }"
		}
	}

	parts := make([]string, 0, len(n.Props)+1)
	if n.Indexer != nil {
		parts = append(parts, n.Indexer.Source())
	}
	for _, prop := range n.Props {
		parts = append(parts, prop.Source())
	}

	if len(parts) == 0 {
		return "{}"
	}

	// tables are always split across multiple lines, one item per line
	return "{\n" + sourceIndent(strings.Join(parts, ",\n"), 1) + ",\n}"
}

func (n AstTypeTypeof) Source() string {
	return "typeof(" + sourceExprPrec(n.Expr, 0) + ")"
}

func (n AstTypeUnion) Source() string {
	return sourceUnionIntersection(n.Types, "|")
}

// -------------------------------------------------------------------------------- -- HELPERS FOR VALUE/POINTER UNION CASES --------------------------------------------------------------------------------
