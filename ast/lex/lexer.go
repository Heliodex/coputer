package lex

import (
	"slices"
)

func LUAU_ASSERT(condition bool) {
	if !condition {
		panic("Assertion failed")
	}
}

type AstName struct {
	Value string
}

type NameTypePair struct {
	Name AstName
	Type LexemeType
}

type Position struct {
	Line, Column uint32
}

type Location struct {
	Begin, End Position
}

func LocationLen(start Position, l uint32) Location {
	return Location{
		Begin: start,
		End:   Position{Line: start.Line, Column: start.Column + l},
	}
}

type AstNameTable struct {
	data map[string]NameTypePair
}

func MakeAstNameTable() AstNameTable {
	t := AstNameTable{make(map[string]NameTypePair)}

	for i := Reserved_BEGIN; i < Reserved_END; i++ {
		t.addStatic(kReserved[i-Reserved_BEGIN], LexemeType(i))
	}
	return t
}

func (t *AstNameTable) addStatic(name string, lt LexemeType) AstName {
	entry := NameTypePair{
		Name: AstName{Value: name},
		Type: lt,
	}

	_, ok := t.data[name]
	LUAU_ASSERT(!ok)
	t.data[name] = entry

	return entry.Name
}

func (t *AstNameTable) getOrAddWithType(name []byte) (r NameTypePair) {
	sn := string(name)
	if entry, ok := t.data[sn]; ok {
		return entry
	}

	var lt LexemeType
	if name[0] == '@' {
		lt = Attribute
	} else {
		lt = Name
	}

	r = NameTypePair{
		Name: AstName{Value: sn},
		Type: lt,
	}
	t.data[sn] = r

	return
}

func (t *AstNameTable) getWithType(name []byte) NameTypePair {
	if entry, ok := t.data[string(name)]; ok {
		return entry
	}
	return NameTypePair{AstName{}, Name}
}

type Lexer struct {
	buffer []byte

	offset, line, lineOffset uint32
	lexeme                   Lexeme
	prevLocation             Location

	names AstNameTable

	skipComments bool
	readNames    bool

	braceStack []BraceType
}

func (l *Lexer) Current() Lexeme {
	return l.lexeme
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

func (l *Lexer) Lookahead() Lexeme {
	currentOffset := l.offset
	currentLine := l.line
	currentLineOffset := l.lineOffset
	currentLexeme := l.lexeme
	currentPrevLocation := l.prevLocation // lel
	currentBraceStackSize := len(l.braceStack)
	var currentBraceType BraceType
	if currentBraceStackSize == 0 {
		currentBraceType = Normal
	} else {
		currentBraceType = l.braceStack[currentBraceStackSize-1]
	}

	result := l.next0()

	l.offset = currentOffset
	l.line = currentLine
	l.lineOffset = currentLineOffset
	l.lexeme = currentLexeme
	l.prevLocation = currentPrevLocation

	if len(l.braceStack) < currentBraceStackSize {
		l.braceStack = append(l.braceStack, currentBraceType)
	} else if len(l.braceStack) > currentBraceStackSize {
		l.braceStack = l.braceStack[:len(l.braceStack)-1]
	}

	return result
}

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
		Data:     l.buffer[startOffset:l.offset],
		rest:     l.buffer[l.offset:],
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
			if l.skipLongSeparator() == sep {
				LUAU_ASSERT(l.peekch0() == ']')
				l.consume() // skip (second) ]

				endOffset := l.offset - uint32(sep) - 2
				LUAU_ASSERT(endOffset >= startOffset)

				return Lexeme{
					Location: Location{start, l.position()},
					Type:     ok,
					Data:     l.buffer[startOffset:endOffset],
					rest:     l.buffer[endOffset:],
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

func (l *Lexer) readQuotedString() Lexeme {
	start := l.position()

	delimiter := l.peekch0()
	LUAU_ASSERT(l.peekch0() == '\'' || l.peekch0() == '"')
	l.consume()

	startOffset := l.offset

	for l.peekch0() != delimiter {
		switch l.peekch0() {
		case 0, '\r', '\n':
			return Lexeme{
				Location: Location{start, l.position()},
				Type:     BrokenString,
			}

		case '\\':
			l.readBackslashInString()

		default:
			l.consume()
		}
	}

	l.consume()

	return Lexeme{
		Location: Location{start, l.position()},
		Type:     QuotedString,
		Data:     l.buffer[startOffset : l.offset-1],
		rest:     l.buffer[l.offset-1:],
	}
}

func (l *Lexer) readInterpolatedStringBegin() Lexeme {
	LUAU_ASSERT(l.peekch0() == '`')

	start := l.position()
	l.consume()

	return l.readInterpolatedStringSection(start, InterpStringBegin, InterpStringSimple)
}

func (l *Lexer) readInterpolatedStringSection(start Position, formatType, endType LexemeType) Lexeme {
	startOffset := l.offset

	for l.peekch0() != '`' {
		switch l.peekch0() {
		case 0, '\r', '\n':
			return Lexeme{
				Location: Location{start, l.position()},
				Type:     BrokenString,
			}

		case '\\':
			// Allow for \u{}, which would otherwise be consumed by looking for {
			if l.peekch(1) == 'u' && l.peekch(2) == '{' {
				l.consume() // backslash
				l.consume() // u
				l.consume() // {
				break
			}

			l.readBackslashInString()

		case '{':
			l.braceStack = append(l.braceStack, InterpolatedString)

			if l.peekch(1) == '{' {
				brokenDoubleBrace := Lexeme{
					Location: Location{start, l.position()},
					Type:     BrokenInterpDoubleBrace,
					Data:     l.buffer[startOffset:l.offset],
					rest:     l.buffer[l.offset:],
				}
				l.consume()
				l.consume()
				return brokenDoubleBrace
			}

			l.consume()
			return Lexeme{
				Location: Location{start, l.position()},
				Type:     formatType,
				Data:     l.buffer[startOffset : l.offset-1],
				rest:     l.buffer[l.offset-1:],
			}

		default:
			l.consume()
		}
	}

	l.consume()

	return Lexeme{
		Location: Location{start, l.position()},
		Type:     endType,
		Data:     l.buffer[startOffset : l.offset-1],
		rest:     l.buffer[l.offset-1:],
	}
}

func (l *Lexer) readNumber(start Position, startOffset uint32) Lexeme {
	LUAU_ASSERT(isDigit(l.peekch0()))

	// This function does not do the number parsing - it only skips a number-like pattern.
	// It uses the same logic as Lua stock lexer; the resulting string is later converted
	// to a number with proper verification.
	for {
		l.consume()
		if !(isDigit(l.peekch0()) || l.peekch0() == '.' || l.peekch0() == '_') {
			break
		}
	}

	if l.peekch0() == 'e' || l.peekch0() == 'E' {
		l.consume()

		if l.peekch0() == '+' || l.peekch0() == '-' {
			l.consume()
		}
	}

	for isAlpha(l.peekch0()) || isDigit(l.peekch0()) || l.peekch0() == '_' {
		l.consume()
	}

	return Lexeme{
		Location: Location{start, l.position()},
		Type:     Number,
		Data:     l.buffer[startOffset:l.offset],
		rest:     l.buffer[l.offset:],
	}
}

func (l *Lexer) readName() NameTypePair {
	LUAU_ASSERT(isAlpha(l.peekch0()) || l.peekch0() == '_' || l.peekch0() == '@')

	startOffset := l.offset

	for {
		l.consume()
		if !(isAlpha(l.peekch0()) || isDigit(l.peekch0()) || l.peekch0() == '_') {
			break
		}
	}

	if l.readNames {
		return l.names.getOrAddWithType(l.buffer[startOffset:][:l.offset-startOffset])
	}
	return l.names.getWithType(l.buffer[startOffset:][:l.offset-startOffset])
}

func (l *Lexer) readNext() Lexeme {
	start := l.position()

	switch l.peekch0() {
	case 0:
		return Lexeme{
			Location: LocationLen(start, 0),
			Type:     Eof,
		}

	case '-':
		if l.peekch(1) == '>' {
			l.consume()
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 2),
				Type:     SkinnyArrow,
			}
		} else if l.peekch(1) == '=' {
			l.consume()
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 2),
				Type:     SubAssign,
			}
		} else if l.peekch(1) == '-' {
			return l.readCommentBody()
		} else {
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '-',
			}
		}

	case '[':
		sep := l.skipLongSeparator()

		if sep >= 0 {
			return l.readLongString(start, sep, RawString, BrokenString)
		} else if sep == -1 {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '[',
			}
		} else {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     BrokenString,
			}
		}

	case '{':
		l.consume()

		if len(l.braceStack) != 0 {
			l.braceStack = append(l.braceStack, Normal)
		}

		return Lexeme{
			Location: LocationLen(start, 1),
			Type:     '{',
		}

	case '}':
		l.consume()

		if len(l.braceStack) == 0 {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '}',
			}
		}

		braceStackTop := l.braceStack[len(l.braceStack)-1]
		l.braceStack = l.braceStack[:len(l.braceStack)-1]

		if braceStackTop != InterpolatedString {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '}',
			}
		}

		return l.readInterpolatedStringSection(start, InterpStringMid, InterpStringEnd)

	case '=':
		l.consume()

		if l.peekch0() == '=' {
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 2),
				Type:     Equal,
			}
		} else {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '=',
			}
		}

	case '<':
		l.consume()

		if l.peekch0() == '=' {
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 2),
				Type:     LessEqual,
			}
		} else {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '<',
			}
		}

	case '>':
		l.consume()

		if l.peekch0() == '=' {
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 2),
				Type:     GreaterEqual,
			}
		} else {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '>',
			}
		}

	case '~':
		l.consume()

		if l.peekch0() == '=' {
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 2),
				Type:     NotEqual,
			}
		} else {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '~', // is this even a valid luau token
			}
		}

	case '"', '\'':
		return l.readQuotedString()

	case '`':
		return l.readInterpolatedStringBegin()

	case '.':
		l.consume()

		if l.peekch0() == '.' {
			l.consume()

			if l.peekch0() == '.' {
				l.consume()

				return Lexeme{
					Location: LocationLen(start, 3),
					Type:     Dot3,
				}
			} else if l.peekch0() == '=' {
				l.consume()

				return Lexeme{
					Location: LocationLen(start, 3),
					Type:     ConcatAssign,
				}
			} else {
				return Lexeme{
					Location: LocationLen(start, 2),
					Type:     Dot2,
				}
			}
		} else {
			if isDigit(l.peekch0()) {
				return l.readNumber(start, l.offset-1)
			} else {
				return Lexeme{
					Location: LocationLen(start, 1),
					Type:     '.',
				}
			}
		}

	case '+':
		l.consume()

		if l.peekch0() == '=' {
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 2),
				Type:     AddAssign,
			}
		} else {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '+',
			}
		}

	case '/':
		l.consume()

		ch := l.peekch0()

		if ch == '=' {
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 2),
				Type:     DivAssign,
			}
		} else if ch == '/' {
			l.consume()

			if l.peekch0() == '=' {
				l.consume()
				return Lexeme{
					Location: LocationLen(start, 3),
					Type:     FloorDivAssign,
				}
			} else {
				return Lexeme{
					Location: LocationLen(start, 2),
					Type:     FloorDiv,
				}
			}
		} else {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '/',
			}
		}

	case '*':
		l.consume()

		if l.peekch0() == '=' {
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 2),
				Type:     MulAssign,
			}
		} else {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '*',
			}
		}

	case '%':
		l.consume()

		if l.peekch0() == '=' {
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 2),
				Type:     ModAssign,
			}
		} else {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '%',
			}
		}

	case '^':
		l.consume()

		if l.peekch0() == '=' {
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 2),
				Type:     PowAssign,
			}
		} else {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     '^',
			}
		}

	case ':':
		l.consume()
		if l.peekch0() == ':' {
			l.consume()
			return Lexeme{
				Location: LocationLen(start, 2),
				Type:     DoubleColon,
			}
		} else {
			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     ':',
			}
		}

	case '(', ')', ']', ';', ',', '#', '?', '&', '|':
		ch := l.peekch0()
		l.consume()

		return Lexeme{
			Location: LocationLen(start, 1),
			Type:     LexemeType(ch),
		}

	case '@':
		attribute := l.readName()
		return Lexeme{
			Location: Location{start, l.position()},
			Type:     Attribute,
			name:     &attribute.Name.Value,
		}

	default:
		if isDigit(l.peekch0()) {
			return l.readNumber(start, l.offset)
		} else if isAlpha(l.peekch0()) || l.peekch0() == '_' {
			name := l.readName()

			return Lexeme{
				Location: Location{start, l.position()},
				Type:     name.Type,
				name:     &name.Name.Value,
			}
		} else if (l.peekch0() & 0x80) != 0 {
			return l.readUtf8Error()
		} else {
			ch := l.peekch0()
			l.consume()

			return Lexeme{
				Location: LocationLen(start, 1),
				Type:     LexemeType(ch),
			}
		}
	}
}

func (l *Lexer) readUtf8Error() Lexeme {
	start := l.position()
	var codepoint uint32
	var size int

	if (l.peekch0() & 0b10000000) == 0b00000000 {
		size = 1
		codepoint = uint32(l.peekch0() & 0x7F)
	} else if (l.peekch0() & 0b11100000) == 0b11000000 {
		size = 2
		codepoint = uint32(l.peekch0() & 0b11111)
	} else if (l.peekch0() & 0b11110000) == 0b11100000 {
		size = 3
		codepoint = uint32(l.peekch0() & 0b1111)
	} else if (l.peekch0() & 0b11111000) == 0b11110000 {
		size = 4
		codepoint = uint32(l.peekch0() & 0b111)
	} else {
		l.consume()
		return Lexeme{
			Location: Location{start, l.position()},
			Type:     BrokenUnicode,
		}
	}

	l.consume()

	for range size {
		if (l.peekch0() & 0b11000000) != 0b10000000 {
			return Lexeme{
				Location: Location{start, l.position()},
				Type:     BrokenUnicode,
			}
		}

		codepoint <<= 6
		codepoint |= uint32(l.peekch0() & 0b00111111)
		l.consume()
	}

	// ?
	return Lexeme{
		Location:  Location{start, l.position()},
		Type:      BrokenUnicode,
		Codepoint: &codepoint,
	}
}

func toUtf8(buf []byte, codepoint uint32) uint {
	if codepoint < 0x80 { // U+0000..U+007F
		buf[0] = byte(codepoint)
		return 1
	} else if codepoint < 0x800 { // U+0080..U+07FF
		buf[0] = byte(0xC0 | (codepoint >> 6))
		buf[1] = byte(0x80 | (codepoint & 0x3F))
		return 2
	} else if codepoint < 0x10000 { // U+0800..U+FFFF
		buf[0] = byte(0xE0 | (codepoint >> 12))
		buf[1] = byte(0x80 | ((codepoint >> 6) & 0x3F))
		buf[2] = byte(0x80 | (codepoint & 0x3F))
		return 3
	} else if codepoint < 0x110000 { // U+10000..U+10FFFF
		buf[0] = byte(0xF0 | (codepoint >> 18))
		buf[1] = byte(0x80 | ((codepoint >> 12) & 0x3F))
		buf[2] = byte(0x80 | ((codepoint >> 6) & 0x3F))
		buf[3] = byte(0x80 | (codepoint & 0x3F))
		return 4
	} else {
		return 0 // invalid code point
	}
}

func (l *Lexer) fixupQuotedString(data *[]byte) bool {
	if len(*data) == 0 || !slices.Contains(*data, '\\') {
		return true
	}

	size := uint(len(*data))
	var write uint

	for i := uint(0); i < size; {
		if (*data)[i] == '\\' {
			(*data)[write] = (*data)[i]
			write++
			i++
			continue
		}

		if i+1 == size {
			return false
		}

		escape := (*data)[i+1]
		i += 2 // skip \e

		switch escape {
		case '\n':
			(*data)[write] = '\n'
			write++

		case '\r':
			(*data)[write] = '\n'
			if i < size && (*data)[i] == '\n' {
				i++
			}

		case 0:
			return false

		case 'x':
			// hex escape codes are exactly 2 hex digits long
			if i+2 > size {
				return false
			}

			var code uint32

			for j := range uint(2) {
				ch := (*data)[i+j]
				if !isHexDigit(ch) {
					return false
				}

				// use or trick to convert to lower case
				var v byte
				if isDigit(ch) {
					v = ch - '0'
				} else {
					v = (ch | ' ') - 'a' + 10
				}
				code = 16*code + uint32(v)
			}

			(*data)[write] = byte(code)
			write++

		case 'z':
			for i < size && isSpace((*data)[i]) {
				i++
			}

		case 'u':
			// unicode escape codes are at least 3 characters including braces
			if i+3 > size {
				return false
			}

			if (*data)[i] != '{' {
				return false
			}
			i++

			if (*data)[i] == '}' {
				return false
			}

			var code uint32

			for range 16 {
				if i == size {
					return false
				}

				ch := (*data)[i]

				if ch == '}' {
					break
				}

				if !isHexDigit(ch) {
					return false
				}

				// use or trick to convert to lower case
				var v byte
				if isDigit(ch) {
					v = ch - '0'
				} else {
					v = (ch | ' ') - 'a' + 10
				}
				code = 16*code + uint32(v)
				i++
			}

			if i == size || (*data)[i] != '}' {
				return false
			}

			i++

			utf8 := toUtf8((*data)[write:][:4], code)
			if utf8 == 0 {
				return false
			}

			write += utf8

		default:
			if isDigit(escape) {
				code := uint32(escape - '0')

				for range 2 {
					if i == size || !isDigit((*data)[i]) {
						break
					}

					code = 10*code + uint32((*data)[i]-'0')
					i++
				}

				if code > 0xFF {
					return false
				}

				(*data)[write] = byte(code)
				write++
			} else {
				(*data)[write] = unescape(escape)
				write++
			}
		}
	}

	LUAU_ASSERT(write <= size)
	*data = (*data)[:write]

	return true
}

func (l *Lexer) fixupMultilineString(data *[]byte) {
	if len(*data) == 0 {
		return
	}

	// Lua rules for multiline strings are as follows:
	// - standalone \r, \r\n, \n\r and \n are all considered newlines
	// - first newline in the multiline string is skipped
	// - all other newlines are normalized to \n

	// Since our lexer just treats \n as newlines, we apply a simplified set of rules that is sufficient to get normalized newlines for Windows/Unix:
	// - \r\n and \n are considered newlines
	// - first newline is skipped
	// - newlines are normalized to \n

	// This makes the string parsing behavior consistent with general lexing behavior - a standalone \r isn't considered a new line from the line
	// tracking perspective

	src := *data
	dst := 0

	// skip leading newline
	if src[0] == '\r' && src[1] == '\n' {
		src = src[2:]
	} else if src[0] == '\n' {
		src = src[1:]
	}

	// parse the rest of the string, converting newlines as we go
	for len(src) > 0 {
		if src[0] == '\r' && src[1] == '\n' {
			(*data)[dst] = '\n'
			dst++
			src = src[2:]
		} else if src[0] == '\n' {
			(*data)[dst] = src[0]
			dst++
			src = src[1:]
		}
	}

	*data = (*data)[:dst-int((*data)[0])]
}
