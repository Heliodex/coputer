package lex

import "fmt"

type LexemeType uint16

const (
	Eof      LexemeType = 0
	Char_END LexemeType = iota + 255
	Equal

	LessEqual
	GreaterEqual
	NotEqual

	Dot2
	Dot3

	SkinnyArrow
	DoubleColon
	FloorDiv

	InterpStringBegin
	InterpStringMid
	InterpStringEnd
	// An interpolated string with no expressions (like `x`)
	InterpStringSimple

	AddAssign
	SubAssign
	MulAssign
	DivAssign
	FloorDivAssign
	ModAssign
	PowAssign
	ConcatAssign

	RawString
	QuotedString
	Number
	Name

	Comment
	BlockComment

	Attribute
	AttributeOpen

	BrokenString
	BrokenComment
	BrokenUnicode
	BrokenInterpDoubleBrace
	Error

	Reserved_BEGIN
)

const (
	ReservedAnd = Reserved_BEGIN + iota
	ReservedBreak
	ReservedDo
	ReservedElse
	ReservedElseif
	ReservedEnd
	ReservedFalse
	ReservedFor
	ReservedFunction
	ReservedIf
	ReservedIn
	ReservedLocal
	ReservedNil
	ReservedNot
	ReservedOr
	ReservedRepeat
	ReservedReturn
	ReservedThen
	ReservedTrue
	ReservedUntil
	ReservedWhile
	Reserved_END
)

type Lexeme struct {
	Type     LexemeType
	Location Location

	data, rest []byte
	name       *string
	codepoint  *uint32
}

func (l Lexeme) String() string {
	switch l.Type {
	case Eof:
		return "<eof>"

	case Equal:
		return "'=='"

	case LessEqual:
		return "'<='"

	case GreaterEqual:
		return "'>='"

	case NotEqual:
		return "'~='"

	case Dot2:
		return "'..'"

	case Dot3:
		return "'...'"

	case SkinnyArrow:
		return "'->'"

	case DoubleColon:
		return "'::'"

	case FloorDiv:
		return "'//'"

	case AddAssign:
		return "'+='"

	case SubAssign:
		return "'-='"

	case MulAssign:
		return "'*='"

	case DivAssign:
		return "'/='"

	case FloorDivAssign:
		return "'//='"

	case ModAssign:
		return "'%='"

	case PowAssign:
		return "'^='"

	case ConcatAssign:
		return "'..='"

	case RawString:
	case QuotedString:
		if l.data != nil {
			return fmt.Sprintf("\"%s\"", string(l.data))
		}
		return "string"

	case InterpStringBegin:
		if l.data != nil {
			return fmt.Sprintf("`%s{", string(l.data))
		}
		return "the beginning of an interpolated string"

	case InterpStringMid:
		if l.data != nil {
			return fmt.Sprintf("}%s{", string(l.data))
		}
		return "the middle of an interpolated string"

	case InterpStringEnd:
		if l.data != nil {
			return fmt.Sprintf("}%s`", string(l.data))
		}
		return "the end of an interpolated string"

	case InterpStringSimple:
		if l.data != nil {
			return fmt.Sprintf("`%s`", string(l.data))
		}
		return "interpolated string"

	case Number:
		if l.data != nil {
			return fmt.Sprintf("'%s'", string(l.data))
		}
		return "number"

	case Name:
		if l.name != nil {
			return fmt.Sprintf("'%s'", *l.name)
		}
		return "identifier"

	case Comment:
		return "comment"

	case Attribute:
		if l.name != nil {
			return fmt.Sprintf("'%s'", *l.name)
		}
		return "attribute"

	case AttributeOpen:
		return "'@["

	case BrokenString:
		return "malformed string"

	case BrokenComment:
		return "unfinished comment"

	case BrokenInterpDoubleBrace:
		return "'{{', which is invalid (did you mean '\\{'?)"

	case BrokenUnicode:
		if l.codepoint != nil {
			confusable := findConfusable(*l.codepoint)
			if confusable != nil {
				return fmt.Sprintf("Unicode character U+%x (did you mean '%s'?)", *l.codepoint, *confusable)
			}
			return fmt.Sprintf("Unicode character U+%x", *l.codepoint)
		}
		return "invalid UTF-8 sequence"
	}

	if l.Type < Char_END {
		return fmt.Sprintf("'%c'", l.Type)
	}
	if l.Type >= Reserved_BEGIN && l.Type < Reserved_END {
		return fmt.Sprintf("'%s'", kReserved[l.Type-Reserved_BEGIN])
	}
	return "<unknown>"
}

type QuoteStyle uint8

const (
	Single QuoteStyle = iota
	Double
)

type BraceType uint8

const (
	InterpolatedString BraceType = iota
	Normal
)

func (l Lexeme) getBlockDepth() uint32 {
	LUAU_ASSERT(l.Type == RawString || l.Type == BlockComment)

	// If we have a well-formed string, we are guaranteed to see 2 `]` characters after the end of the string contents
	LUAU_ASSERT(l.rest[0] == ']')
	var depth uint32
	for {
		depth++
		if l.rest[depth] == ']' {
			break
		}
	}

	return depth - 1
}

func (l Lexeme) getQuoteStyle() QuoteStyle {
	LUAU_ASSERT(l.Type == QuotedString)

	// If we have a well-formed string, we are guaranteed to see a closing delimiter after the string
	LUAU_ASSERT(len(l.data) > 0)

	quote := l.rest[0]
	switch quote {
	case '\'':
		return Single
	case '"':
		return Double
	}

	panic("unknown quote style")
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' || ch == '\v' || ch == '\f'
}

var kReserved = [...]string{"and", "break", "do", "else", "elseif", "end", "false", "for", "function", "if", "in", "local", "nil", "not", "or", "repeat", "return", "then", "true", "until", "while"}

func isAlpha(ch byte) bool {
	// use or trick to convert to lower case and unsigned comparison to do range check
	return uint8((ch|' ')-'a') < 26
}

func isDigit(ch byte) bool {
	return uint8(ch-'0') < 10
}

func isHexDigit(ch byte) bool {
	// use or trick to convert to lower case and unsigned comparison to do range check
	return (uint8(ch-'0') < 10) || (uint8((ch|' ')-'a') < 6)
}

func isNewline(ch byte) bool {
	return ch == '\n'
}

func unescape(ch byte) byte {
	switch ch {
	case 'a':
		return '\a'
	case 'b':
		return '\b'
	case 'f':
		return '\f'
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	case 'v':
		return '\v'
	default:
		return ch
	}
}
