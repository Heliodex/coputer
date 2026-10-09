package main

import (
	"sort"
)

// This file implements formatAST, the AST-to-AST transformation step of formatting. Parsing produces statements and comments in source order; formatAST reorders the top-level chunk (comment directives, `game:GetService` declarations and `require` declarations are each sorted) so that rendering only lays out already-ordered code. Run it between Parse and Source; Minify intentionally skips it and keeps source order.

// formatAST returns result with its top-level block transformed, mirroring what renderRoot used to do while rendering: the header directives are pulled together and sorted, the leading runs of `game:GetService` and `require` declarations are sorted, and the header section is split off. The statements and comments keep travelling together as units, so every comment stays attached to its statement.
//
// The reordered statements and comments are given fresh line numbers (columns are preserved) so comment attachment re-derives identically when rendering. Statement locations are retargeted in place; comment locations are replaced, leaving Result.CommentLocations reporting the original positions.
func formatAST(result Result) Result {
	root := result.Root

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
