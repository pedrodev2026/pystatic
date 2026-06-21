package lexer

import (
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `def hello():
    x = 5
    if x == 5:
        print(f"x is {x}")
    return x`

	expectedTokens := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{DEF, "def"},
		{IDENT, "hello"},
		{LPAREN, "("},
		{RPAREN, ")"},
		{COLON, ":"},
		{NEWLINE, "\n"},
		{INDENT, "    "},
		{IDENT, "x"},
		{ASSIGN, "="},
		{NUMBER, "5"},
		{NEWLINE, "\n"},
		{IF, "if"},
		{IDENT, "x"},
		{EQ, "=="},
		{NUMBER, "5"},
		{COLON, ":"},
		{NEWLINE, "\n"},
		{INDENT, "    "},
		{IDENT, "print"},
		{LPAREN, "("},
		{FSTRING, "x is {x}"},
		{RPAREN, ")"},
		{NEWLINE, "\n"},
		{DEDENT, ""},
		{RETURN, "return"},
		{IDENT, "x"},
		{NEWLINE, "\n"},
		{DEDENT, ""},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range expectedTokens {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestIndentationNesting(t *testing.T) {
	input := `if True:
    if True:
        x = 1
x = 2`

	expectedTokens := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{IF, "if"},
		{TRUE, "True"},
		{COLON, ":"},
		{NEWLINE, "\n"},
		{INDENT, "    "},
		{IF, "if"},
		{TRUE, "True"},
		{COLON, ":"},
		{NEWLINE, "\n"},
		{INDENT, "    "},
		{IDENT, "x"},
		{ASSIGN, "="},
		{NUMBER, "1"},
		{NEWLINE, "\n"},
		{DEDENT, ""},
		{DEDENT, ""},
		{IDENT, "x"},
		{ASSIGN, "="},
		{NUMBER, "2"},
		{NEWLINE, "\n"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range expectedTokens {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestParenNesting(t *testing.T) {
	input := `x = (
    1,
    2
)
y = 3`

	expectedTokens := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{IDENT, "x"},
		{ASSIGN, "="},
		{LPAREN, "("},
		{NUMBER, "1"},
		{COMMA, ","},
		{NUMBER, "2"},
		{RPAREN, ")"},
		{NEWLINE, "\n"},
		{IDENT, "y"},
		{ASSIGN, "="},
		{NUMBER, "3"},
		{NEWLINE, "\n"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range expectedTokens {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestTripleQuotes(t *testing.T) {
	input := `x = """hello
world"""
y = 1`

	expectedTokens := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{IDENT, "x"},
		{ASSIGN, "="},
		{STRING, "hello\nworld"},
		{NEWLINE, "\n"},
		{IDENT, "y"},
		{ASSIGN, "="},
		{NUMBER, "1"},
		{NEWLINE, "\n"},
		{EOF, ""},
	}

	l := New(input)

	for i, tt := range expectedTokens {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q (literal=%q)",
				i, tt.expectedType, tok.Type, tok.Literal)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestInconsistentIndentation(t *testing.T) {
	input := `if True:
    x = 1
  y = 2`

	l := New(input)

	// Skip: 'if', 'True', ':', '\n', 'INDENT', 'x', '=', '1', '\n'
	for i := 0; i < 9; i++ {
		l.NextToken()
	}

	tok := l.NextToken()
	if tok.Type != ILLEGAL {
		t.Fatalf("expected ILLEGAL token for inconsistent indentation, got %q (literal=%q)", tok.Type, tok.Literal)
	}
}

