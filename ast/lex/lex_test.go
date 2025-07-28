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
	lexer := Lexer{buffer: []byte(testinput)}
	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, BrokenString)
	CHECK_EQ(t, lexeme.Location, Location{Position{0, 0}, Position{0, 2}})
}

func TestBrokenComment(t *testing.T) {
	const testinput = "--[[  "
	lexer := Lexer{buffer: []byte(testinput)}
	lexeme := lexer.next0()
	CHECK_EQ(t, lexeme.Type, BrokenComment)
	CHECK_EQ(t, lexeme.Location, Location{Position{0, 0}, Position{0, 6}})
}

func TestBrokenCommentKept(t *testing.T) {
	const testinput = "--[[  "
	lexer := Lexer{buffer: []byte(testinput), skipComments: true}
	CHECK_EQ(t, lexer.next0().Type, BrokenComment)
}

func TestCommentSkipped(t *testing.T) {
	const testinput = "--  "
	lexer := Lexer{buffer: []byte(testinput), skipComments: true}
	CHECK_EQ(t, lexer.next0().Type, Eof)
}

func TestMultilineCommentWithLexemeInAndAfter(t *testing.T) {
	const testinput = "--[[ function \n]] end"
	lexer := Lexer{buffer: []byte(testinput)}
	comment := lexer.next0()
	end := lexer.next0()

	CHECK_EQ(t, comment.Type, BlockComment)
	CHECK_EQ(t, comment.Location, Location{Position{0, 0}, Position{1, 2}})
	CHECK_EQ(t, end.Type, ReservedEnd)
	CHECK_EQ(t, end.Location, Location{Position{1, 3}, Position{1, 6}})
}
