package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pedrodev2026/pystatic/internal/lexer"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// Program Node
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// ----------------------------------------------------
// Statements
// ----------------------------------------------------

// DefStmt (Function Definition)
type Parameter struct {
	Name *Identifier
	Type string // e.g. "http.Request" or "int"
}

type DefStmt struct {
	Token      lexer.Token // 'def' token
	Name       *Identifier
	Parameters []Parameter
	ReturnType string // e.g. "http.Response" or empty
	Body       *BlockStmt
}

func (ds *DefStmt) statementNode()       {}
func (ds *DefStmt) TokenLiteral() string { return ds.Token.Literal }
func (ds *DefStmt) String() string {
	var out bytes.Buffer
	out.WriteString("def " + ds.Name.String() + "(")
	var params []string
	for _, p := range ds.Parameters {
		if p.Type != "" {
			params = append(params, p.Name.String()+": "+p.Type)
		} else {
			params = append(params, p.Name.String())
		}
	}
	out.WriteString(strings.Join(params, ", ") + ")")
	if ds.ReturnType != "" {
		out.WriteString(" -> " + ds.ReturnType)
	}
	out.WriteString(":\n" + ds.Body.String())
	return out.String()
}

// BlockStmt
type BlockStmt struct {
	Token      lexer.Token // INDENT token
	Statements []Statement
}

func (bs *BlockStmt) statementNode()       {}
func (bs *BlockStmt) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStmt) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString("    " + s.String() + "\n")
	}
	return out.String()
}

// IfStmt
type ElifClause struct {
	Condition   Expression
	Consequence *BlockStmt
}

type IfStmt struct {
	Token       lexer.Token // 'if' token
	Condition   Expression
	Consequence *BlockStmt
	Elifs       []ElifClause
	Alternative *BlockStmt
}

func (is *IfStmt) statementNode()       {}
func (is *IfStmt) TokenLiteral() string { return is.Token.Literal }
func (is *IfStmt) String() string {
	var out bytes.Buffer
	out.WriteString("if " + is.Condition.String() + ":\n" + is.Consequence.String())
	for _, elif := range is.Elifs {
		out.WriteString("elif " + elif.Condition.String() + ":\n" + elif.Consequence.String())
	}
	if is.Alternative != nil {
		out.WriteString("else:\n" + is.Alternative.String())
	}
	return out.String()
}

// ForStmt
type ForStmt struct {
	Token    lexer.Token // 'for' token
	Vars     []*Identifier
	Iterable Expression
	Body     *BlockStmt
}

func (fs *ForStmt) statementNode()       {}
func (fs *ForStmt) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStmt) String() string {
	var out bytes.Buffer
	var vars []string
	for _, v := range fs.Vars {
		vars = append(vars, v.String())
	}
	out.WriteString("for " + strings.Join(vars, ", ") + " in " + fs.Iterable.String() + ":\n" + fs.Body.String())
	return out.String()
}

// WhileStmt
type WhileStmt struct {
	Token     lexer.Token // 'while' token
	Condition Expression
	Body      *BlockStmt
}

func (ws *WhileStmt) statementNode()       {}
func (ws *WhileStmt) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStmt) String() string {
	var out bytes.Buffer
	out.WriteString("while " + ws.Condition.String() + ":\n" + ws.Body.String())
	return out.String()
}

// ReturnStmt
type ReturnStmt struct {
	Token lexer.Token // 'return' token
	Value Expression  // can be nil
}

func (rs *ReturnStmt) statementNode()       {}
func (rs *ReturnStmt) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStmt) String() string {
	var out bytes.Buffer
	out.WriteString("return")
	if rs.Value != nil {
		out.WriteString(" " + rs.Value.String())
	}
	return out.String()
}

// AssignStmt
type AssignStmt struct {
	Token lexer.Token // '=' token
	Left  Expression
	Value Expression
}

func (as *AssignStmt) statementNode()       {}
func (as *AssignStmt) TokenLiteral() string { return as.Token.Literal }
func (as *AssignStmt) String() string {
	return as.Left.String() + " = " + as.Value.String()
}

// ImportStmt
type ImportStmt struct {
	Token lexer.Token // 'import' token
	Name  string      // e.g. "http"
}

func (is *ImportStmt) statementNode()       {}
func (is *ImportStmt) TokenLiteral() string { return is.Token.Literal }
func (is *ImportStmt) String() string {
	return "import " + is.Name
}

// ExprStmt
type ExprStmt struct {
	Token      lexer.Token
	Expression Expression
}

func (es *ExprStmt) statementNode()       {}
func (es *ExprStmt) TokenLiteral() string { return es.Token.Literal }
func (es *ExprStmt) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// ClassStmt
type ClassStmt struct {
	Token lexer.Token // 'class' token
	Name  *Identifier
	Body  *BlockStmt
}

func (cs *ClassStmt) statementNode()       {}
func (cs *ClassStmt) TokenLiteral() string { return cs.Token.Literal }
func (cs *ClassStmt) String() string {
	return "class " + cs.Name.String() + ":\n" + cs.Body.String()
}

// MatchStmt
type MatchCase struct {
	Pattern Expression
	Body    *BlockStmt
}

type MatchStmt struct {
	Token   lexer.Token // 'match' token
	Subject Expression
	Cases   []MatchCase
}

func (ms *MatchStmt) statementNode()       {}
func (ms *MatchStmt) TokenLiteral() string { return ms.Token.Literal }
func (ms *MatchStmt) String() string {
	var out bytes.Buffer
	out.WriteString("match " + ms.Subject.String() + ":\n")
	for _, c := range ms.Cases {
		out.WriteString("    case " + c.Pattern.String() + ":\n" + c.Body.String())
	}
	return out.String()
}

// ----------------------------------------------------
// Expressions
// ----------------------------------------------------

// Identifier
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// NumberLiteral
type NumberLiteral struct {
	Token lexer.Token
	Value string
}

func (nl *NumberLiteral) expressionNode()      {}
func (nl *NumberLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NumberLiteral) String() string       { return nl.Value }

// StringLiteral
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return fmt.Sprintf("%q", sl.Value) }

// FStringLiteral
type FStringLiteral struct {
	Token lexer.Token
	Value string
}

func (fsl *FStringLiteral) expressionNode()      {}
func (fsl *FStringLiteral) TokenLiteral() string { return fsl.Token.Literal }
func (fsl *FStringLiteral) String() string       { return "f" + fmt.Sprintf("%q", fsl.Value) }

// BooleanLiteral
type BooleanLiteral struct {
	Token lexer.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode()      {}
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return fmt.Sprintf("%t", bl.Value) }

// NoneLiteral
type NoneLiteral struct {
	Token lexer.Token
}

func (nl *NoneLiteral) expressionNode()      {}
func (nl *NoneLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NoneLiteral) String() string       { return "None" }

// ListLiteral
type ListLiteral struct {
	Token    lexer.Token // '['
	Elements []Expression
}

func (ll *ListLiteral) expressionNode()      {}
func (ll *ListLiteral) TokenLiteral() string { return ll.Token.Literal }
func (ll *ListLiteral) String() string {
	var elems []string
	for _, el := range ll.Elements {
		elems = append(elems, el.String())
	}
	return "[" + strings.Join(elems, ", ") + "]"
}

// DictLiteral
type DictLiteral struct {
	Token lexer.Token // '{'
	Keys  []Expression
	Value []Expression
}

func (dl *DictLiteral) expressionNode()      {}
func (dl *DictLiteral) TokenLiteral() string { return dl.Token.Literal }
func (dl *DictLiteral) String() string {
	var pairs []string
	for i := range dl.Keys {
		pairs = append(pairs, dl.Keys[i].String()+": "+dl.Value[i].String())
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

// PrefixExpr
type PrefixExpr struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpr) expressionNode()      {}
func (pe *PrefixExpr) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpr) String() string       { return "(" + pe.Operator + pe.Right.String() + ")" }

// InfixExpr
type InfixExpr struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpr) expressionNode()      {}
func (ie *InfixExpr) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpr) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

// CallExpr
type KeywordArg struct {
	Name  *Identifier
	Value Expression
}

type CallExpr struct {
	Token     lexer.Token // '('
	Function  Expression
	Arguments []Expression
	Keywords  []KeywordArg
}

func (ce *CallExpr) expressionNode()      {}
func (ce *CallExpr) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpr) String() string {
	var args []string
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	for _, kw := range ce.Keywords {
		args = append(args, kw.Name.String()+"="+kw.Value.String())
	}
	return ce.Function.String() + "(" + strings.Join(args, ", ") + ")"
}

// DotExpr
type DotExpr struct {
	Token  lexer.Token // '.'
	Left   Expression
	Member *Identifier
}

func (de *DotExpr) expressionNode()      {}
func (de *DotExpr) TokenLiteral() string { return de.Token.Literal }
func (de *DotExpr) String() string {
	return de.Left.String() + "." + de.Member.String()
}

// IndexExpr
type IndexExpr struct {
	Token lexer.Token // '['
	Left  Expression
	Index Expression
}

func (ie *IndexExpr) expressionNode()      {}
func (ie *IndexExpr) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpr) String() string {
	return ie.Left.String() + "[" + ie.Index.String() + "]"
}
