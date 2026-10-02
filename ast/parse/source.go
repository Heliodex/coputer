package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/Heliodex/coputer/ast/lex"
)

// This file implements Source() for every AST node. Each Source() renders the
// node as formatted Luau code starting at column zero, delegating to the
// Source() methods of its children and re-indenting their output where needed.
//
// Statements/expressions that span multiple lines indent their own contents
// with tabs relative to their first line, so a parent can embed them by
// indenting every line once more with sourceIndent.

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

// sourceLongString renders a long (`[[...]]`) string with enough equals signs
// to avoid an accidental terminator.
func sourceLongString(value string) string {
	for eq := 0; ; eq++ {
		eqs := strings.Repeat("=", eq)
		strend := "]" + eqs + "]"
		if !strings.Contains(value, strend) {
			return "[" + eqs + "[" + value + strend
		}
	}
}

// normalizeCommentContent removes the common leading indentation from the
// continuation lines of a block comment. Without this, re-indenting a comment
// when rendering would keep adding whitespace every time the output is parsed
// and rendered again.
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

// sourceString renders a string literal in the requested quote style, falling
// back to a safer representation when needed.
func sourceString(value string, style QuoteStyle) string {
	// long strings are avoided when they contain newlines: a multi-line
	// literal can't be re-indented without changing its value
	if style == QuoteStyle_QuotedRaw && !strings.ContainsAny(value, "\r\n") {
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
	for i := range len(value) {
		ch := value[i]
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

	rep := fmt.Sprintf("%g", value)
	rep = strings.Replace(rep, "e+", "e", 1)

	for strings.Contains(rep, "e0") {
		rep = strings.Replace(rep, "e0", "e", 1)
	}
	for strings.Contains(rep, "e-0") {
		rep = strings.Replace(rep, "e-0", "e-", 1)
	}

	if strings.Contains(rep, "e-") || !strings.Contains(rep, "e") {
		return rep
	}

	// remove the exponent if it merely reflects the number of decimal places
	eSplit := strings.Split(rep, "e")
	if len(eSplit) == 1 {
		return rep
	}
	dotSplit := strings.Split(eSplit[0], ".")
	if len(dotSplit) == 1 {
		return rep
	}

	start, decimal, exponent := dotSplit[0], dotSplit[1], eSplit[1]
	if exponent == fmt.Sprintf("%d", len(decimal)) {
		return start + decimal
	}
	return rep
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

// sourceStat renders a statement, preserving a trailing semicolon when the
// source had one (semicolons can be semantically significant).
func sourceStat(stat AstStat) string {
	s := stat.Source()
	if sourceHasSemicolon(stat) {
		s += ";"
	}
	return s
}

// sourceStatList renders statements separated by newlines, interleaving the
// block's comments in their original positions: leading comments stay above the
// statement they precede and trailing comments stay on the same line.
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

// sourceBlockBody renders a block's statements indented one level, preceded by
// a newline. Empty blocks render as an empty string.
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

// appendEnd appends an `end` to a block body, collapsing it onto the previous
// line when that line already ends a nested block. The collapsed line is
// unindented one level, so a chain of ends lines up with the outermost block.
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

// sourceExprList renders comma-separated expressions.
func sourceExprList(exprs []AstExpr) string {
	parts := make([]string, len(exprs))
	for i, expr := range exprs {
		parts[i] = expr.Source()
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
	switch expr.(type) {
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

	for i := range s {
		ch := s[i]
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

// sourceHasIfExpr reports whether any of the expressions is an if-then-else
// expression.
func sourceHasIfExpr(exprs []AstExpr) bool {
	for _, expr := range exprs {
		if _, ok := asIfElseExpr(expr); ok {
			return true
		}
	}
	return false
}

// writeIfBranch writes an if-expression branch value, putting a multi-line
// value on its own indented line.
func writeIfBranch(b *strings.Builder, expr AstExpr) {
	if expr == nil {
		b.WriteString(" nil")
		return
	}

	s := expr.Source()
	if strings.Contains(s, "\n") {
		b.WriteByte('\n')
		b.WriteString(sourceIndent(s, 1))
		return
	}

	b.WriteByte(' ')
	b.WriteString(s)
}

// writeAssignedValue writes ` = values`, putting if-expressions on their own
// indented lines.
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

// sourceVarargAnnotation renders a vararg annotation. The parser wraps
// parameter vararg annotations in a variadic type pack, but the `...` is
// already written as part of the parameter list.
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

// writeAnnotation writes `: annotation`, putting a multi-line annotation on its
// own indented lines.
func writeAnnotation(b *strings.Builder, annotation string) {
	if strings.Contains(annotation, "\n") {
		b.WriteString(":\n")
		b.WriteString(sourceIndent(annotation, 1))
		return
	}

	b.WriteString(": ")
	b.WriteString(annotation)
}

// writeTypeAnnotation writes `: type`, putting a multi-line type on its own
// indented lines.
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
	switch key := expr.(type) {
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

// sourceUnionIntersection renders union and intersection types. Types are
// always split across multiple lines, one per line, each prefixed by op.
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

// --------------------------------------------------------------------------------
// -- EXPRESSIONS
// --------------------------------------------------------------------------------

func (n AstExprBinary) Source() string {
	return n.Left.Source() + " " + sourceBinaryOp(BinaryOp(n.Op)) + " " + n.Right.Source()
}

func (n AstExprCall) Source() string {
	var b strings.Builder

	b.WriteString(n.Func.Source())

	if n.TypeArguments != nil && len(*n.TypeArguments) > 0 {
		b.WriteString("<<")
		b.WriteString(sourceTypeOrPackList(*n.TypeArguments))
		b.WriteString(">>")
	}

	// a call with a single string or table argument can omit its parentheses
	if len(n.Args) == 1 && sourceCallSugar(n.Args[0]) {
		b.WriteByte(' ')
		b.WriteString(n.Args[0].Source())
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
		return n.Expressions[0].Source()
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

// sourceRest renders a function body from the generics onwards, i.e. without
// the leading `function` keyword. Named function statements use this so the
// name can be spliced in between `function` and the parameter list.
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

// sourceIf renders an if-then-else expression with the given leading keyword,
// so nested else-if expressions can be rendered as `elseif` chains. Each
// if/elseif/else section is placed on its own line.
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
	b.WriteString(n.Condition.Source())
	b.WriteString(" then")
	writeIfBranch(&b, n.TrueExpr)

	if falseExpr, ok := asIfElseExpr(n.FalseExpr); ok {
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
	return n.Expr.Source() + "[" + n.Index.Source() + "]"
}

func (n AstExprIndexName) Source() string {
	return n.Expr.Source() + string(n.Op) + n.Index
}

func (n AstExprInterpString) Source() string {
	var b strings.Builder

	b.WriteByte('`')
	for i, str := range n.Strings {
		b.WriteString(sourceInterpStringPart(str))
		if i < len(n.Expressions) {
			b.WriteByte('{')
			b.WriteString(n.Expressions[i].Source())
			b.WriteByte('}')
		}
	}
	b.WriteByte('`')

	return b.String()
}

func (n AstExprInstantiate) Source() string {
	return n.Expr.Source() + "<<" + sourceTypeOrPackList(n.TypeArguments) + ">>"
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
		parts[i] = item.Source()
	}

	// tables are always split across multiple lines, one item per line
	return "{\n" + sourceIndent(strings.Join(parts, ",\n"), 1) + ",\n}"
}

func (n AstExprTableItem) Source() string {
	if n.Key == nil {
		return n.Value.Source()
	}
	switch n.Kind {
	case Record:
		if key, ok := (*n.Key).(AstExprConstantString); ok {
			return key.Value + " = " + n.Value.Source()
		}
		return (*n.Key).Source() + " = " + n.Value.Source()
	case General:
		// a string key that's a valid identifier can use the record form
		if key, ok := sourceRecordKey(*n.Key); ok {
			return key + " = " + n.Value.Source()
		}
		return "[" + (*n.Key).Source() + "] = " + n.Value.Source()
	}
	return n.Value.Source()
}

func (n AstExprTypeAssertion) Source() string {
	return n.Expr.Source() + " :: " + n.Annotation.Source()
}

func (n AstExprVarargs) Source() string {
	return "..."
}

func (n AstExprUnary) Source() string {
	op := sourceUnaryOp(n.Op)
	expr := n.Expr.Source()

	// avoid producing a `--` comment from `- -x`
	if op == "-" && strings.HasPrefix(expr, "-") {
		return op + " " + expr
	}

	return op + expr
}

// --------------------------------------------------------------------------------
// -- LOCALS, GENERICS AND ATTRIBUTES
// --------------------------------------------------------------------------------

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

// --------------------------------------------------------------------------------
// -- STATEMENTS
// --------------------------------------------------------------------------------

func (n AstStatAssign) Source() string {
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

	b.WriteString(n.Var.Source())
	b.WriteByte(' ')
	b.WriteString(sourceBinaryOp(n.Op))
	b.WriteString("=")

	value := n.Value.Source()
	if _, ok := asIfElseExpr(n.Value); ok {
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
	return n.Expr.Source()
}

func (n AstStatFor) Source() string {
	var b strings.Builder

	b.WriteString("for ")
	if n.Var != nil {
		b.WriteString(n.Var.Source())
	}
	b.WriteString(" = ")
	b.WriteString(n.From.Source())
	b.WriteString(", ")
	b.WriteString(n.To.Source())
	if n.Step != nil {
		b.WriteString(", ")
		b.WriteString(n.Step.Source())
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
	b.WriteString(n.Name.Source())
	b.WriteString(n.Func.sourceRest())

	return b.String()
}

func (n AstStatIf) Source() string {
	return n.sourceIf("if")
}

// sourceIf renders an if statement with the given leading keyword, so nested
// else-if statements can be rendered as `elseif` chains.
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

	b.WriteString(n.Condition.Source())
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
	b.WriteString(n.Condition.Source())

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

	// types that span multiple lines start on their own line
	typeSource := n.Type.Source()
	b.WriteString(" =")
	if strings.Contains(typeSource, "\n") {
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
	b.WriteString(n.Condition.Source())
	b.WriteString(" do")

	var body string
	if n.Body != nil {
		body = sourceBlockBody(*n.Body)
	}
	b.WriteString(appendEnd(body))

	return b.String()
}

// --------------------------------------------------------------------------------
// -- TYPES
// --------------------------------------------------------------------------------

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

	if strings.Contains(ret, "\n") {
		b.WriteString(" ->\n")
		b.WriteString(sourceIndent(ret, 1))
	} else {
		b.WriteString(" -> ")
		b.WriteString(ret)
	}

	return b.String()
}

// sourceParamsAndReturn renders the parameter list and return type of a
// function type, so callers (e.g. declared methods) can splice a name in.
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

	return "{ " + strings.Join(parts, ", ") + " }"
}

func (n AstTypeTypeof) Source() string {
	return "typeof(" + n.Expr.Source() + ")"
}

func (n AstTypeUnion) Source() string {
	return sourceUnionIntersection(n.Types, "|")
}

// --------------------------------------------------------------------------------
// -- HELPERS FOR VALUE/POINTER UNION CASES
// --------------------------------------------------------------------------------
