package main

func LUAU_ASSERT(condition bool) {
	if !condition {
		panic("Assertion failed")
	}
}

type Position struct {
	Line   uint32
	Column uint32
}

type Location struct {
	Start Position
	End   Position
}

type Lexer struct {
	buffer []byte

	offset, line, lineOffset uint32
	lexeme                   Lexeme
	prevLocation             Location

	skipComments bool
	readNames    bool

	braceStack []BraceType
}

func (l *Lexer) next0() Lexeme {
	return l.next(l.skipComments, true)
}

func (l *Lexer) next(skipComments, updatePrevLocation bool) Lexeme {
	// in skipComments mode we reject valid comments
	for {
		for isSpace(l.peekch0()) {
			l.consumeAny()
		}

		if updatePrevLocation {
			l.prevLocation = l.lexeme.Location
		}

		l.lexeme = l.readNext()
		updatePrevLocation = false

		if !(skipComments && (l.lexeme.Type == Comment || l.lexeme.Type == BlockComment)) {
			break
		}
	}

	return l.lexeme
}

func (l *Lexer) nextline() {
	for l.peekch0() != 0 && l.peekch0() != '\r' && !isNewline(l.peekch0()) {
		l.consume()
	}

	l.next0()
}

func (l *Lexer) lookahead() Lexeme {}

func (l *Lexer) isReserved(word string) bool {
	for i := Reserved_BEGIN; i < Reserved_END; i++ {
		if word == kReserved[i-Reserved_BEGIN] {
			return true
		}
	}

	return false
}

func (l *Lexer) peekch0() byte {
	if l.offset < uint32(len(l.buffer)) {
		return l.buffer[l.offset]
	}
	return 0
}

func (l *Lexer) peekch(lookahead uint32) byte {
	if (l.offset + lookahead) < uint32(len(l.buffer)) {
		return l.buffer[l.offset+lookahead]
	}
	return 0
}

func (l *Lexer) position() Position {
	return Position{l.line, l.offset - l.lineOffset}
}

func (l *Lexer) consume() {
	// consume() assumes current character is known to not be a newline; use consumeAny if this is not guaranteed
	LUAU_ASSERT(!isNewline(l.buffer[l.offset]))

	l.offset++
}

func (l *Lexer) consumeAny() {
	if isNewline(l.buffer[l.offset]) {
		l.line++
		l.lineOffset = l.offset + 1
	}

	l.offset++
}

func (l *Lexer) readCommentBody() Lexeme {
	start := l.position()

	LUAU_ASSERT(l.peekch(0) == '-' && l.peekch(1) == '-')
	l.consume()
	l.consume()

	startOffset := l.offset

	if l.peekch0() == '[' {
		sep := l.skipLongSeparator()

		if sep >= 0 {
			return l.readLongString(start, sep, BlockComment, BrokenComment)
		}
	}

	// fall back to single-line comment
	for l.peekch0() != 0 && l.peekch0() != '\r' && !isNewline(l.peekch0()) {
		l.consume()
	}

	return Lexeme{
		Location: Location{start, l.position()},
		Type:     Comment,
		data:     l.buffer[startOffset:l.offset],
	}
}

// Given a sequence [===[ or ]===], returns:
// 1. number of equal signs (or 0 if none present) between the brackets
// 2. -1 if this is not a long comment/string separator
// 3. -N if this is a malformed separator
// Does *not* consume the closing brace.
func (l *Lexer) skipLongSeparator() int {
	start := l.peekch0()

	LUAU_ASSERT(start == '[' || start == ']')
	l.consume()

	var count int

	for l.peekch0() == '=' {
		l.consume()
		count++
	}

	if start == l.peekch0() {
		return count
	}
	return -count - 1
}

func (l *Lexer) readLongString(start Position, sep int, ok, broken LexemeType) Lexeme {
	// skip (second) [
	LUAU_ASSERT(l.peekch0() == '[')
	l.consume()

	startOffset := l.offset

	for l.peekch0() != 0 {
		if l.peekch0() == ']' {
			if l.skipLongSeparator() == ']' {
				LUAU_ASSERT(l.peekch0() == ']')
				l.consume() // skip (second) ]

				endOffset := l.offset - uint32(sep) - 2
				LUAU_ASSERT(endOffset >= startOffset)

				return Lexeme{
					Location: Location{start, l.position()},
					Type:     ok,
					data:     l.buffer[startOffset:endOffset],
				}
			}
		} else {
			l.consumeAny()
		}
	}

	return Lexeme{
		Location: Location{start, l.position()},
		Type:     broken,
	}
}

func (l *Lexer) readBackslashInString() {
	LUAU_ASSERT(l.peekch0() == '\\')
	l.consume()
	switch l.peekch0() {
	case '\r':
		l.consume()
		if l.peekch0() == '\n' {
			l.consumeAny()
		}

	case 0:

	case 'z':
		l.consume()
		for isSpace(l.peekch0()) {
			l.consumeAny()
		}

	default:
		l.consumeAny()
	}
}
