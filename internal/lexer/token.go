package lexer

type TokenType string

const (
	// Special tokens
	EOF     TokenType = "EOF"
	ILLEGAL TokenType = "ILLEGAL"

	// Formatting
	NEWLINE TokenType = "NEWLINE"
	INDENT  TokenType = "INDENT"
	DEDENT  TokenType = "DEDENT"

	// Identifiers & Literals
	IDENT   TokenType = "IDENT"
	NUMBER  TokenType = "NUMBER"
	STRING  TokenType = "STRING"
	FSTRING TokenType = "FSTRING"

	// Operators and Delimiters
	PLUS     TokenType = "+"
	MINUS    TokenType = "-"
	ASTERISK TokenType = "*"
	SLASH    TokenType = "/"
	PERCENT  TokenType = "%"
	EQ       TokenType = "=="
	NOT_EQ   TokenType = "!="
	LT       TokenType = "<"
	GT       TokenType = ">"
	LTE      TokenType = "<="
	GTE      TokenType = ">="
	ASSIGN   TokenType = "="
	COLON    TokenType = ":"
	ARROW    TokenType = "->"
	COMMA    TokenType = ","
	DOT      TokenType = "."
	LPAREN   TokenType = "("
	RPAREN   TokenType = ")"
	LBRACKET TokenType = "["
	RBRACKET TokenType = "]"
	LBRACE   TokenType = "{"
	RBRACE   TokenType = "}"

	// Keywords
	DEF      TokenType = "def"
	IF       TokenType = "if"
	ELIF     TokenType = "elif"
	ELSE     TokenType = "else"
	WHILE    TokenType = "while"
	FOR      TokenType = "for"
	IN       TokenType = "in"
	RETURN   TokenType = "return"
	IMPORT   TokenType = "import"
	CLASS    TokenType = "class"
	MATCH    TokenType = "match"
	CASE     TokenType = "case"
	BREAK    TokenType = "break"
	CONTINUE TokenType = "continue"
	NONE     TokenType = "None"
	TRUE     TokenType = "True"
	FALSE    TokenType = "False"
	AND      TokenType = "and"
	OR       TokenType = "or"
	NOT      TokenType = "not"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

var keywords = map[string]TokenType{
	"def":      DEF,
	"if":       IF,
	"elif":     ELIF,
	"else":     ELSE,
	"while":    WHILE,
	"for":      FOR,
	"in":       IN,
	"return":   RETURN,
	"import":   IMPORT,
	"class":    CLASS,
	"match":    MATCH,
	"case":     CASE,
	"break":    BREAK,
	"continue": CONTINUE,
	"None":     NONE,
	"True":     TRUE,
	"False":    FALSE,
	"and":      AND,
	"or":       OR,
	"not":      NOT,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
