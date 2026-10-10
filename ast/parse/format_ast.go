package main

import (
	"sort"
)

// This file implements formatAST, the AST-to-AST transformation step of formatting. Parsing produces statements and comments in source order; formatAST reorders the top-level chunk (comment directives, `game:GetService` declarations and `require` declarations are each sorted) so that rendering only lays out already-ordered code. Run it between Parse and Source; Minify intentionally skips it and keeps source order.

// formatAST returns result with its top-level block transformed, mirroring what renderRoot used to do while rendering: the header directives are pulled together and sorted, the leading runs of `game:GetService` and `require` declarations are sorted, and the header section is split off. Plain function assignments anywhere in the tree are rewritten as function declarations first. The statements and comments keep travelling together as units, so every comment stays attached to its statement.
//
// The transformation mutates the result's statements in place, so treat the input as consumed. The reordered statements and comments are given fresh line numbers (columns are preserved) so comment attachment re-derives identically when rendering. Statement locations are retargeted in place; comment locations are replaced, leaving Result.CommentLocations reporting the original positions.
func formatAST(result Result) Result {
	root := result.Root
	formatBlockFunctionDecls(&root)

	units := sourceStatUnits(root.Body, root.Comments)
	sortHeaderDirectives(units)

	serviceEnd := serviceSectionEnd(units)

	requireEnd := serviceEnd
	for requireEnd < len(units) {
		if _, _, ok := requireDeclaration(units[requireEnd].stat); !ok {
			break
		}
		requireEnd++
	}

	// the leading comments form their own section when they hold a directive, or when a sorted section follows them, so they aren't glued to the first declaration
	var header []Comment
	if len(units) > 0 && len(units[0].leading) > 0 &&
		(hasDirective(units[0].leading) || serviceEnd > 0 || requireEnd > 0) {
		header = units[0].leading
		units[0].leading = nil
	}

	if serviceEnd >= 2 {
		run := units[:serviceEnd]
		sort.SliceStable(run, func(i, j int) bool {
			a, _ := serviceDeclarationName(run[i].stat)
			b, _ := serviceDeclarationName(run[j].stat)
			return a < b
		})
	}

	if requireEnd-serviceEnd >= 2 {
		run := units[serviceEnd:requireEnd]
		sort.SliceStable(run, func(i, j int) bool {
			ak, as, _ := requireDeclaration(run[i].stat)
			bk, bs, _ := requireDeclaration(run[j].stat)
			if as != bs {
				return as
			}
			return ak < bk
		})
	}

	// the first declaration after sorting may carry its own comments; fold them into the header so reformatting the result doesn't reclassify them
	if header != nil && len(units) > 0 && len(units[0].leading) > 0 {
		header = append(header, units[0].leading...)
		units[0].leading = nil
	}

	root.Body, root.Comments = flattenStatUnits(header, units)
	result.Root = root
	return result
}

// flattenStatUnits lays units back out as a statement list and a comment list, assigning fresh line numbers in order so the units re-derive identically when rendering. Columns are preserved. Statements are retargeted in place; each comment gets a fresh location, leaving any shared locations untouched.
func flattenStatUnits(header []Comment, units []sourceStatUnit) ([]AstStat, []Comment) {
	var body []AstStat
	var comments []Comment

	line := uint32(1)
	nextLine := func() uint32 {
		l := line
		line++
		return l
	}

	placeComment := func(c *Comment, l uint32) {
		loc := c.GetLocation()
		loc.Begin.Line = l
		loc.End.Line = l
		c.NodeLoc = &NodeLoc{Location: loc}
		comments = append(comments, *c)
	}

	placeLeading := func(leading []Comment) {
		for i := range leading {
			placeComment(&leading[i], nextLine())
		}
	}

	for i := range header {
		placeComment(&header[i], nextLine())
	}

	for i := range units {
		placeLeading(units[i].leading)

		if units[i].stat != nil {
			statLine := nextLine()
			loc := units[i].stat.GetLocation()
			loc.Begin.Line = statLine
			loc.End.Line = statLine
			units[i].stat.SetLocation(loc)
			body = append(body, units[i].stat)

			for j := range units[i].trailing {
				placeComment(&units[i].trailing[j], statLine)
			}
		}
	}

	return body, comments
}

// asFuncValue unwraps parentheses around expr and reports whether it is an anonymous function value. Type assertions are not unwrapped: an annotated function stays an assignment.
func asFuncValue(expr AstExpr) (AstExprFunction, bool) {
	switch e := sourceUnwrapGroup(expr).(type) {
	case AstExprFunction:
		return e, true
	case *AstExprFunction:
		return *e, true
	}
	return AstExprFunction{}, false
}

// asLocalFunction rewrites `local x = function ... end` (and the `const` and `export` equivalents) as a local function declaration, reporting whether it applies. Only a single unannotated binding with a single function value qualifies: an annotation has nowhere to go on a declaration, and `export const function` isn't valid syntax so the constness couldn't survive the rewrite.
func asLocalFunction(n *AstStatLocal) (*AstStatLocalFunction, bool) {
	if n == nil || len(n.Vars) != 1 || len(n.Values) != 1 {
		return nil, false
	}
	if n.IsExported && n.IsConst {
		return nil, false
	}
	if n.Vars[0].Annotation != nil {
		return nil, false
	}
	fn, ok := asFuncValue(n.Values[0])
	if !ok || fn.Self != nil {
		return nil, false
	}

	return &AstStatLocalFunction{
		NodeLoc:      n.NodeLoc,
		Name:         n.Vars[0],
		Func:         fn,
		IsConst:      n.IsConst,
		HasSemicolon: n.HasSemicolon,
	}, true
}

// isFuncNameTarget reports whether target can head a `function` declaration: a global or a dotted path off one. A local would redirect to a global, `a[b]` has no declaration form, and `a:b` would gain an implicit self parameter.
func isFuncNameTarget(target AstExpr) bool {
	switch t := sourceUnwrapGroup(target).(type) {
	case AstExprGlobal, *AstExprGlobal:
		return true
	case AstExprIndexName:
		return t.Op == '.' && isFuncNameTarget(t.Expr)
	case *AstExprIndexName:
		return t.Op == '.' && isFuncNameTarget(t.Expr)
	}
	return false
}

// asStatFunction rewrites `x = function ... end` as a function statement, reporting whether it applies. Only a single suitable target with a single function value qualifies; see isFuncNameTarget for what a suitable target is.
func asStatFunction(n *AstStatAssign) (*AstStatFunction, bool) {
	if n == nil || len(n.Vars) != 1 || len(n.Values) != 1 {
		return nil, false
	}
	if !isFuncNameTarget(n.Vars[0]) {
		return nil, false
	}
	fn, ok := asFuncValue(n.Values[0])
	if !ok || fn.Self != nil {
		return nil, false
	}

	return &AstStatFunction{
		NodeLoc:      n.NodeLoc,
		Name:         n.Vars[0],
		Func:         fn,
		HasSemicolon: n.HasSemicolon,
	}, true
}

// formatBlockFunctionDecls rewrites plain function assignments as function declarations throughout block, recursing into nested blocks and function bodies. Converted statements reuse their old locations so comments stay attached.
func formatBlockFunctionDecls(block *AstStatBlock) {
	if block == nil {
		return
	}
	for i, stat := range block.Body {
		block.Body[i] = formatStatFunctionDecl(stat)
	}
}

// formatStatFunctionDecl rewrites a single statement, returning the replacement. Nested blocks and function bodies are transformed first, so a moved function value arrives already transformed.
func formatStatFunctionDecl(stat AstStat) AstStat {
	if stat == nil {
		return nil
	}

	switch n := stat.(type) {
	case *AstStatLocal:
		for _, value := range n.Values {
			formatExprFunctionDecls(value)
		}
		if conv, ok := asLocalFunction(n); ok {
			return conv
		}
	case *AstStatAssign:
		for _, target := range n.Vars {
			formatExprFunctionDecls(target)
		}
		for _, value := range n.Values {
			formatExprFunctionDecls(value)
		}
		if conv, ok := asStatFunction(n); ok {
			return conv
		}
	case *AstStatBlock:
		formatBlockFunctionDecls(n)
	case *AstStatIf:
		formatExprFunctionDecls(n.Condition)
		formatBlockFunctionDecls(&n.ThenBody)
		n.ElseBody = formatStatFunctionDecl(n.ElseBody)
	case *AstStatFor:
		formatExprFunctionDecls(n.From)
		formatExprFunctionDecls(n.To)
		formatExprFunctionDecls(n.Step)
		formatBlockFunctionDecls(n.Body)
	case *AstStatForIn:
		for _, value := range n.Values {
			formatExprFunctionDecls(value)
		}
		formatBlockFunctionDecls(n.Body)
	case *AstStatWhile:
		formatExprFunctionDecls(n.Condition)
		formatBlockFunctionDecls(n.Body)
	case *AstStatRepeat:
		formatExprFunctionDecls(n.Condition)
		formatBlockFunctionDecls(n.Body)
	case *AstStatFunction:
		formatExprFunctionDecls(n.Name)
		formatBlockFunctionDecls(&n.Func.Body)
	case *AstStatLocalFunction:
		formatBlockFunctionDecls(&n.Func.Body)
	case *AstStatExpr:
		formatExprFunctionDecls(n.Expr)
	case *AstStatReturn:
		for _, value := range n.List {
			formatExprFunctionDecls(value)
		}
	case *AstStatCompoundAssign:
		formatExprFunctionDecls(n.Var)
		formatExprFunctionDecls(n.Value)
	case *AstStatError:
		for _, value := range n.Expressions {
			formatExprFunctionDecls(value)
		}
		for i, s := range n.Statements {
			n.Statements[i] = formatStatFunctionDecl(s)
		}
	case *AstStatTypeFunction:
		formatBlockFunctionDecls(&n.Body.Body)
	}

	return stat
}

// formatExprFunctionDecls walks expr looking for nested function bodies to transform. Expressions themselves are never replaced, only the statement lists inside function bodies.
func formatExprFunctionDecls(expr AstExpr) {
	if expr == nil {
		return
	}

	switch n := expr.(type) {
	case *AstExprFunction:
		formatFuncValueBodies(&n.Attributes, &n.Body)
	case AstExprFunction:
		fn := n
		formatFuncValueBodies(&fn.Attributes, &fn.Body)
	case *AstExprCall:
		formatExprFunctionDecls(n.Func)
		for _, arg := range n.Args {
			formatExprFunctionDecls(arg)
		}
	case AstExprCall:
		formatExprFunctionDecls(n.Func)
		for _, arg := range n.Args {
			formatExprFunctionDecls(arg)
		}
	case *AstExprBinary:
		formatExprFunctionDecls(n.Left)
		formatExprFunctionDecls(n.Right)
	case AstExprBinary:
		formatExprFunctionDecls(n.Left)
		formatExprFunctionDecls(n.Right)
	case *AstExprUnary:
		formatExprFunctionDecls(n.Expr)
	case AstExprUnary:
		formatExprFunctionDecls(n.Expr)
	case *AstExprGroup:
		formatExprFunctionDecls(n.Expr)
	case AstExprGroup:
		formatExprFunctionDecls(n.Expr)
	case *AstExprTypeAssertion:
		formatExprFunctionDecls(n.Expr)
	case AstExprTypeAssertion:
		formatExprFunctionDecls(n.Expr)
	case *AstExprIndexExpr:
		formatExprFunctionDecls(n.Expr)
		formatExprFunctionDecls(n.Index)
	case AstExprIndexExpr:
		formatExprFunctionDecls(n.Expr)
		formatExprFunctionDecls(n.Index)
	case *AstExprIndexName:
		formatExprFunctionDecls(n.Expr)
	case AstExprIndexName:
		formatExprFunctionDecls(n.Expr)
	case *AstExprInterpString:
		for _, value := range n.Expressions {
			formatExprFunctionDecls(value)
		}
	case AstExprInterpString:
		for _, value := range n.Expressions {
			formatExprFunctionDecls(value)
		}
	case *AstExprIfElse:
		formatExprFunctionDecls(n.Condition)
		formatExprFunctionDecls(n.TrueExpr)
		formatExprFunctionDecls(n.FalseExpr)
	case AstExprIfElse:
		formatExprFunctionDecls(n.Condition)
		formatExprFunctionDecls(n.TrueExpr)
		formatExprFunctionDecls(n.FalseExpr)
	case *AstExprInstantiate:
		formatExprFunctionDecls(n.Expr)
	case AstExprInstantiate:
		formatExprFunctionDecls(n.Expr)
	case *AstExprTable:
		for i := range n.Items {
			if n.Items[i].Key != nil {
				formatExprFunctionDecls(*n.Items[i].Key)
			}
			formatExprFunctionDecls(n.Items[i].Value)
		}
	case AstExprTable:
		for i := range n.Items {
			if n.Items[i].Key != nil {
				formatExprFunctionDecls(*n.Items[i].Key)
			}
			formatExprFunctionDecls(n.Items[i].Value)
		}
	case *AstExprError:
		for _, value := range n.Expressions {
			formatExprFunctionDecls(value)
		}
	case AstExprError:
		for _, value := range n.Expressions {
			formatExprFunctionDecls(value)
		}
	}
}

// formatFuncValueBodies transforms the bodies nested in a function value: attribute arguments (which can hold functions) and the function body itself.
func formatFuncValueBodies(attrs *[]AstAttr, body *AstStatBlock) {
	for _, attr := range *attrs {
		for _, arg := range attr.Args {
			formatExprFunctionDecls(arg)
		}
	}
	formatBlockFunctionDecls(body)
}
