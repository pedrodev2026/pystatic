package lexer

import (
	"fmt"
	"strings"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int
	column       int

	indentStack []int
	parenDepth  int // tracks nesting depth of (), [], {}
	atLineStart bool

	tokenBuffer []Token
}

func New(input string) *Lexer {
	l := &Lexer{
		input:       input,
		line:        1,
		column:      0,
		indentStack: []int{0},
		atLineStart: true,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}

	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.column++
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) skipComment() {
	if l.ch == '#' {
		for l.ch != '\n' && l.ch != 0 {
			l.readChar()
		}
	}
}

func (l *Lexer) handleLineStart() {
	for {
		indentPos := l.position
		indentVal := 0

		// Count spaces/tabs on this line
		for indentPos < len(l.input) {
			c := l.input[indentPos]
			if c == ' ' {
				indentVal++
				indentPos++
			} else if c == '\t' {
				indentVal += 4
				indentPos++
			} else {
				break
			}
		}

		// Check what the first non-whitespace character is
		var firstNonWS byte = 0
		if indentPos < len(l.input) {
			firstNonWS = l.input[indentPos]
		}

		// If it's an empty line, a carriage return, or a comment, skip it
		if firstNonWS == '\n' || firstNonWS == '\r' || firstNonWS == '#' || firstNonWS == 0 {
			for l.position < indentPos {
				l.readChar()
			}
			if l.ch == '#' {
				for l.ch != '\n' && l.ch != 0 {
					l.readChar()
				}
			}
			if l.ch == '\r' {
				l.readChar()
			}
			if l.ch == '\n' {
				l.readChar()
			}
			if l.ch == 0 {
				break
			}
			continue
		}

		// Consume the spaces/tabs that make up the indentation
		for l.position < indentPos {
			l.readChar()
		}

		if l.parenDepth == 0 {
			top := l.indentStack[len(l.indentStack)-1]
			if indentVal > top {
				l.indentStack = append(l.indentStack, indentVal)
				l.tokenBuffer = append(l.tokenBuffer, Token{
					Type:    INDENT,
					Literal: strings.Repeat(" ", indentVal-top),
					Line:    l.line,
					Column:  1,
				})
			} else if indentVal < top {
				// Pop until match
				for len(l.indentStack) > 1 && l.indentStack[len(l.indentStack)-1] > indentVal {
					l.indentStack = l.indentStack[:len(l.indentStack)-1]
					l.tokenBuffer = append(l.tokenBuffer, Token{
						Type:    DEDENT,
						Literal: "",
						Line:    l.line,
						Column:  1,
					})
				}
				newTop := l.indentStack[len(l.indentStack)-1]
				if newTop != indentVal {
					l.tokenBuffer = []Token{{
						Type:    ILLEGAL,
						Literal: fmt.Sprintf("inconsistent indentation (expected %d, got %d)", newTop, indentVal),
						Line:    l.line,
						Column:  l.column,
					}}
				}
			}
		}

		l.atLineStart = false
		break
	}
}

func (l *Lexer) NextToken() Token {
	if len(l.tokenBuffer) > 0 {
		tok := l.tokenBuffer[0]
		l.tokenBuffer = l.tokenBuffer[1:]
		return tok
	}

	if l.atLineStart {
		l.handleLineStart()
		if len(l.tokenBuffer) > 0 {
			tok := l.tokenBuffer[0]
			l.tokenBuffer = l.tokenBuffer[1:]
			return tok
		}
	}

	l.skipWhitespace()

	if l.ch == '#' {
		l.skipComment()
	}

	var tok Token
	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case 0:
		if !l.atLineStart {
			l.atLineStart = true
			tok.Type = NEWLINE
			tok.Literal = "\n"
			return tok
		}
		if len(l.indentStack) > 1 {
			l.indentStack = l.indentStack[:len(l.indentStack)-1]
			tok.Type = DEDENT
			tok.Literal = ""
			return tok
		}
		tok.Type = EOF
		tok.Literal = ""
	case '\n':
		if l.parenDepth > 0 {
			l.readChar()
			return l.NextToken()
		}
		tok.Type = NEWLINE
		tok.Literal = "\n"
		l.atLineStart = true
		l.readChar()
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: EQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: ASSIGN, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
		l.readChar()
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NOT_EQ, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
			l.readChar()
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
			l.readChar()
		}
	case '<':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: LTE, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: LT, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
		l.readChar()
	case '>':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: GTE, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: GT, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
		l.readChar()
	case '-':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: ARROW, Literal: string(ch) + string(l.ch), Line: tok.Line, Column: tok.Column}
		} else {
			tok = Token{Type: MINUS, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		}
		l.readChar()
	case '+':
		tok = Token{Type: PLUS, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case '*':
		tok = Token{Type: ASTERISK, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case '/':
		tok = Token{Type: SLASH, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case '%':
		tok = Token{Type: PERCENT, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case ':':
		tok = Token{Type: COLON, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case ',':
		tok = Token{Type: COMMA, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case '.':
		tok = Token{Type: DOT, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case '(':
		l.parenDepth++
		tok = Token{Type: LPAREN, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case ')':
		if l.parenDepth > 0 {
			l.parenDepth--
		}
		tok = Token{Type: RPAREN, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case '[':
		l.parenDepth++
		tok = Token{Type: LBRACKET, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case ']':
		if l.parenDepth > 0 {
			l.parenDepth--
		}
		tok = Token{Type: RBRACKET, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case '{':
		l.parenDepth++
		tok = Token{Type: LBRACE, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case '}':
		if l.parenDepth > 0 {
			l.parenDepth--
		}
		tok = Token{Type: RBRACE, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
		l.readChar()
	case '"', '\'':
		return l.readString()
	default:
		if l.ch == 'f' || l.ch == 'F' {
			peek := l.peekChar()
			if peek == '"' || peek == '\'' {
				return l.readString()
			}
		}

		if isLetter(l.ch) {
			ident := l.readIdentifier()
			tok.Type = LookupIdent(ident)
			tok.Literal = ident
			return tok
		} else if isDigit(l.ch) {
			return l.readNumber()
		} else {
			tok = Token{Type: ILLEGAL, Literal: string(l.ch), Line: tok.Line, Column: tok.Column}
			l.readChar()
		}
	}

	return tok
}

func (l *Lexer) readString() Token {
	line := l.line
	column := l.column

	isFString := false
	if l.ch == 'f' || l.ch == 'F' {
		isFString = true
		l.readChar()
	}

	quote := l.ch
	l.readChar()

	isTriple := false
	if l.ch == quote && l.peekChar() == quote {
		isTriple = true
		l.readChar()
		l.readChar()
	}

	var sb strings.Builder
	for {
		if l.ch == 0 {
			return Token{Type: ILLEGAL, Literal: "unterminated string literal", Line: line, Column: column}
		}

		if isTriple {
			if l.ch == quote && l.peekChar() == quote {
				if l.readPosition < len(l.input) && l.input[l.readPosition] == quote {
					l.readChar()
					l.readChar()
					l.readChar()
					break
				}
			}
		} else {
			if l.ch == quote {
				l.readChar()
				break
			}
			if l.ch == '\n' {
				return Token{Type: ILLEGAL, Literal: "newline in single-quoted string literal", Line: line, Column: column}
			}
		}

		if l.ch == '\\' {
			sb.WriteByte(l.ch)
			l.readChar()
			if l.ch != 0 {
				sb.WriteByte(l.ch)
				l.readChar()
			}
		} else {
			sb.WriteByte(l.ch)
			l.readChar()
		}
	}

	tokType := STRING
	if isFString {
		tokType = FSTRING
	}

	return Token{Type: tokType, Literal: sb.String(), Line: line, Column: column}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() Token {
	line := l.line
	column := l.column
	position := l.position

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return Token{
		Type:    NUMBER,
		Literal: l.input[position:l.position],
		Line:    line,
		Column:  column,
	}
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
