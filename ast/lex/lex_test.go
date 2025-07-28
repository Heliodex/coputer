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
