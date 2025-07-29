package main

import (
	"fmt"
	"testing"
)

func CHECK_EQ[T comparable](t *testing.T, a, b T) {
	if a != b {
		panic(fmt.Sprintf("Expected %v, got %v", b, a))
	}
}

func TestBrokenStringWorks(t *testing.T) {
	const testinput = "[["
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}
	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, BrokenString)
	CHECK_EQ(t, lexeme.Location, Location{Position{0, 0}, Position{0, 2}})
}

func TestBrokenComment(t *testing.T) {
	const testinput = "--[[  "
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}
	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, BrokenComment)
	CHECK_EQ(t, lexeme.Location, Location{Position{0, 0}, Position{0, 6}})
}

func TestBrokenCommentKept(t *testing.T) {
	const testinput = "--[[  "
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, skipComments: true, readNames: true}
	CHECK_EQ(t, lexer.next0().Type, BrokenComment)
}

func TestCommentSkipped(t *testing.T) {
	const testinput = "--  "
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, skipComments: true, readNames: true}
	CHECK_EQ(t, lexer.next0().Type, Eof)
}

func TestMultilineCommentWithLexemeInAndAfter(t *testing.T) {
	const testinput = "--[[ function \n]] end"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}
	comment := lexer.next0()
	end := lexer.next0()

	CHECK_EQ(t, comment.Type, BlockComment)
	CHECK_EQ(t, comment.Location, Location{Position{0, 0}, Position{1, 2}})
	CHECK_EQ(t, end.Type, ReservedEnd)
	CHECK_EQ(t, end.Location, Location{Position{1, 3}, Position{1, 6}})
}

func TestBrokenEscapeTolerant(t *testing.T) {
	const testinput = "'\\3729472897292378'"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}
	lexeme := lexer.next0()

	CHECK_EQ(t, lexeme.Type, QuotedString)
	CHECK_EQ(t, lexeme.Location, Location{Position{0, 0}, Position{0, uint32(len(testinput))}})
}

func TestBigDelimiters(t *testing.T) {
	const testinput = "--[===[\n\n\n\n]===]"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}
	item := lexer.next0()

	CHECK_EQ(t, item.Type, BlockComment)
	CHECK_EQ(t, item.Location, Location{Position{0, 0}, Position{4, 5}})
}

func TestLookahead(t *testing.T) {
	const testinput = "foo --[[ comment ]] bar : nil end"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, skipComments: true, readNames: true}
	lexer.next0() // must call next() before reading data from lexer at least once

	CHECK_EQ(t, lexer.current().Type, Name)
	CHECK_EQ(t, *(lexer.current().name), "foo")
	CHECK_EQ(t, lexer.lookahead().Type, Name)
	CHECK_EQ(t, *(lexer.lookahead().name), "bar")

	lexer.next0()

	CHECK_EQ(t, lexer.current().Type, Name)
	CHECK_EQ(t, *(lexer.current().name), "bar")
	CHECK_EQ(t, lexer.lookahead().Type, ':')

	lexer.next0()

	CHECK_EQ(t, lexer.current().Type, ':')
	CHECK_EQ(t, lexer.lookahead().Type, ReservedNil)

	lexer.next0()

	CHECK_EQ(t, lexer.current().Type, ReservedNil)
	CHECK_EQ(t, lexer.lookahead().Type, ReservedEnd)

	lexer.next0()

	CHECK_EQ(t, lexer.current().Type, ReservedEnd)
	CHECK_EQ(t, lexer.lookahead().Type, Eof)

	lexer.next0()

	CHECK_EQ(t, lexer.current().Type, Eof)
	CHECK_EQ(t, lexer.lookahead().Type, Eof)
}

func TestStringInterpolationBasic(t *testing.T) {
	const testinput = "`foo {\"bar\"}`"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	interpBegin := lexer.next0()
	CHECK_EQ(t, interpBegin.Type, InterpStringBegin)

	quote := lexer.next0()
	CHECK_EQ(t, quote.Type, QuotedString)

	interpEnd := lexer.next0()
	CHECK_EQ(t, interpEnd.Type, InterpStringEnd)
	// The InterpStringEnd should start with }, not `.
	CHECK_EQ(t, interpEnd.Location.Start.Column, 11)
}

func TestStringInterpolationFull(t *testing.T) {
	const testinput = "`foo {\"bar\"} {\"baz\"} end`"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	interpBegin := lexer.next0()
	CHECK_EQ(t, interpBegin.Type, InterpStringBegin)
	CHECK_EQ(t, interpBegin.String(), "`foo {")

	quote1 := lexer.next0()
	CHECK_EQ(t, quote1.Type, QuotedString)
	CHECK_EQ(t, quote1.String(), "\"bar\"")

	interpMid := lexer.next0()
	CHECK_EQ(t, interpMid.Type, InterpStringMid)
	CHECK_EQ(t, interpMid.String(), "} {")
	CHECK_EQ(t, interpMid.Location.Start.Column, 11)

	quote2 := lexer.next0()
	CHECK_EQ(t, quote2.Type, QuotedString)
	CHECK_EQ(t, quote2.String(), "\"baz\"")

	interpEnd := lexer.next0()
	CHECK_EQ(t, interpEnd.Type, InterpStringEnd)
	CHECK_EQ(t, interpEnd.String(), "} end`")
	CHECK_EQ(t, interpEnd.Location.Start.Column, 19)
}

func TestStringInterpolationDoubleBrace(t *testing.T) {
	const testinput = "`foo{{bad}}bar`"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	brokenInterpBegin := lexer.next0()
	CHECK_EQ(t, brokenInterpBegin.Type, BrokenInterpDoubleBrace)
	CHECK_EQ(t, string(brokenInterpBegin.data), string("foo"))

	CHECK_EQ(t, lexer.next0().Type, Name)

	interpEnd := lexer.next0()
	CHECK_EQ(t, interpEnd.Type, InterpStringEnd)
	CHECK_EQ(t, string(interpEnd.data), string("}bar"))
}

func TestStringInterpolationDoubleButUnmatchedBrace(t *testing.T) {
	const testinput = "`{{oops}`, 1"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	CHECK_EQ(t, lexer.next0().Type, BrokenInterpDoubleBrace)
	CHECK_EQ(t, lexer.next0().Type, Name)
	CHECK_EQ(t, lexer.next0().Type, InterpStringEnd)
	CHECK_EQ(t, lexer.next0().Type, ',')
	CHECK_EQ(t, lexer.next0().Type, Number)
}

func TestStringInterpolationUnmatchedBrace(t *testing.T) {
	const testinput = "{\n        `hello {\"world\"}\n    } -- this might be incorrectly parsed as a string"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	CHECK_EQ(t, lexer.next0().Type, '{')
	CHECK_EQ(t, lexer.next0().Type, InterpStringBegin)
	CHECK_EQ(t, lexer.next0().Type, QuotedString)
	CHECK_EQ(t, lexer.next0().Type, BrokenString)
	CHECK_EQ(t, lexer.next0().Type, '}')
}

func TestStringInterpolationWithUnicodeEscape(t *testing.T) {
	const testinput = "`\\u{1F41B}`"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	CHECK_EQ(t, lexer.next0().Type, InterpStringSimple)
	CHECK_EQ(t, lexer.next0().Type, Eof)
}

func TestSingleQuotedString(t *testing.T) {
	const testinput = "'test'"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, QuotedString)
	CHECK_EQ(t, lexeme.getQuoteStyle(), Single)
}

func TestDoubleQuotedString(t *testing.T) {
	const testinput = `"test"`
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, QuotedString)
	CHECK_EQ(t, lexeme.getQuoteStyle(), Double)
}

func TestLexerDeterminesStringBlockDepth0(t *testing.T) {
	const testinput = "[[ test ]]"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, RawString)
	CHECK_EQ(t, lexeme.getBlockDepth(), 0)
}

func TestLexerDeterminesStringBlockDepth0Multiline1(t *testing.T) {
	const testinput = `[[ test
    ]]`

	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, RawString)
	CHECK_EQ(t, lexeme.getBlockDepth(), 0)
}

func TestLexerDeterminesStringBlockDepth0Multiline2(t *testing.T) {
	const testinput = `[[
    test
    ]]`

	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, RawString)
	CHECK_EQ(t, lexeme.getBlockDepth(), 0)
}

func TestLexerDeterminesStringBlockDepth0Multiline3(t *testing.T) {
	const testinput = `[[
    test ]]`

	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, RawString)
	CHECK_EQ(t, lexeme.getBlockDepth(), 0)
}

func TestLexerDeterminesStringBlockDepth1(t *testing.T) {
	const testinput = "[=[[%s]]=]"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, RawString)
	CHECK_EQ(t, lexeme.getBlockDepth(), 1)
}

func TestLexerDeterminesStringBlockDepth2(t *testing.T) {
	const testinput = "[==[ test ]==]"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, RawString)
	CHECK_EQ(t, lexeme.getBlockDepth(), 2)
}

func TestLexerDeterminesStringBlockDepth2Multiline1(t *testing.T) {
	const testinput = `[==[ test
    ]==]`
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, RawString)
	CHECK_EQ(t, lexeme.getBlockDepth(), 2)
}

func TestLexerDeterminesStringBlockDepth2Multiline2(t *testing.T) {
	const testinput = `[==[
    test
    ]==]`
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, RawString)
	CHECK_EQ(t, lexeme.getBlockDepth(), 2)
}

func TestLexerDeterminesStringBlockDepth2Multiline3(t *testing.T) {
	const testinput = `[==[

    test ]==]`
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, RawString)
	CHECK_EQ(t, lexeme.getBlockDepth(), 2)
}

func TestLexerDeterminesCommentBlockDepth0(t *testing.T) {
	const testinput = "--[[ test ]]"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, BlockComment)
	CHECK_EQ(t, lexeme.getBlockDepth(), 0)
}

func TestLexerDeterminesStringBlockDepth21(t *testing.T) {
	const testinput = "--[=[ μέλλον ]=]"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, BlockComment)
	CHECK_EQ(t, lexeme.getBlockDepth(), 1)
}

func TestLexerDeterminesStringBlockDepth22(t *testing.T) {
	const testinput = "--[==[ test ]==]"
	table := NewAstNameTable()
	lexer := Lexer{buffer: []byte(testinput), names: table, readNames: true}

	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, BlockComment)
	CHECK_EQ(t, lexeme.getBlockDepth(), 2)
}
