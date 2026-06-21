package parser

import (
	"fmt"

	"github.com/pedrodev2026/pystatic/internal/ast"
	"github.com/pedrodev2026/pystatic/internal/lexer"
)

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or < or >= or <=
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or not X
	CALL        // myFunction(X)
	INDEX       // array[index] or dot.member
)

var precedences = map[lexer.TokenType]int{
	lexer.EQ:       EQUALS,
	lexer.NOT_EQ:   EQUALS,
	lexer.LT:       LESSGREATER,
	lexer.GT:       LESSGREATER,
	lexer.LTE:      LESSGREATER,
	lexer.GTE:      LESSGREATER,
	lexer.PLUS:     SUM,
	lexer.MINUS:    SUM,
	lexer.SLASH:    PRODUCT,
	lexer.ASTERISK: PRODUCT,
	lexer.PERCENT:  PRODUCT,
	lexer.LPAREN:   CALL,
	lexer.LBRACKET: INDEX,
	lexer.DOT:      INDEX,
	lexer.AND:      EQUALS,
	lexer.OR:       EQUALS,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  lexer.Token
	peekToken lexer.Token

	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[lexer.TokenType]prefixParseFn)
	p.registerPrefix(lexer.IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.NUMBER, p.parseNumberLiteral)
	p.registerPrefix(lexer.STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.FSTRING, p.parseFStringLiteral)
	p.registerPrefix(lexer.TRUE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.FALSE, p.parseBooleanLiteral)
	p.registerPrefix(lexer.NONE, p.parseNoneLiteral)
	p.registerPrefix(lexer.MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(lexer.LBRACKET, p.parseListLiteral)
	p.registerPrefix(lexer.LBRACE, p.parseDictLiteral)

	p.infixParseFns = make(map[lexer.TokenType]infixParseFn)
	p.registerInfix(lexer.PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.ASTERISK, p.parseInfixExpression)
	p.registerInfix(lexer.PERCENT, p.parseInfixExpression)
	p.registerInfix(lexer.EQ, p.parseInfixExpression)
	p.registerInfix(lexer.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.LT, p.parseInfixExpression)
	p.registerInfix(lexer.GT, p.parseInfixExpression)
	p.registerInfix(lexer.LTE, p.parseInfixExpression)
	p.registerInfix(lexer.GTE, p.parseInfixExpression)
	p.registerInfix(lexer.AND, p.parseInfixExpression)
	p.registerInfix(lexer.OR, p.parseInfixExpression)

	p.registerInfix(lexer.LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.LBRACKET, p.parseIndexExpression)
	p.registerInfix(lexer.DOT, p.parseDotExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekError(t lexer.TokenType) {
	msg := fmt.Sprintf("[%d:%d] expected next token to be %s, got %s instead",
		p.peekToken.Line, p.peekToken.Column, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	msg := fmt.Sprintf("[%d:%d] no prefix parse function for %s found",
		p.curToken.Line, p.curToken.Column, t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	for !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		if isSimpleStatement(stmt) {
			p.nextToken()
		}
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	for p.curTokenIs(lexer.NEWLINE) {
		p.nextToken()
	}

	switch p.curToken.Type {
	case lexer.IMPORT:
		return p.parseImportStatement()
	case lexer.DEF:
		return p.parseDefStatement()
	case lexer.IF:
		return p.parseIfStatement()
	case lexer.WHILE:
		return p.parseWhileStatement()
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.RETURN:
		return p.parseReturnStatement()
	case lexer.CLASS:
		return p.parseClassStatement()
	case lexer.MATCH:
		return p.parseMatchStatement()
	default:
		return p.parseExpressionOrAssignmentStatement()
	}
}

func (p *Parser) parseImportStatement() *ast.ImportStmt {
	stmt := &ast.ImportStmt{Token: p.curToken}
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = p.curToken.Literal

	for p.peekTokenIs(lexer.DOT) {
		p.nextToken() // consume '.'
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		stmt.Name += "." + p.curToken.Literal
	}

	return stmt
}

func (p *Parser) parseDefStatement() *ast.DefStmt {
	stmt := &ast.DefStmt{Token: p.curToken}
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseDefParameters()

	if p.peekTokenIs(lexer.ARROW) {
		p.nextToken() // consume '->'
		if !p.expectPeek(lexer.IDENT) {
			return nil
		}
		retType := p.curToken.Literal
		for p.peekTokenIs(lexer.DOT) {
			p.nextToken() // consume '.'
			if !p.expectPeek(lexer.IDENT) {
				return nil
			}
			retType += "." + p.curToken.Literal
		}
		stmt.ReturnType = retType
	}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}

	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}

	if !p.expectPeek(lexer.INDENT) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseDefParameters() []ast.Parameter {
	var params []ast.Parameter
	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()

	for {
		param := ast.Parameter{}
		if !p.curTokenIs(lexer.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf("[%d:%d] expected identifier in parameter definition, got %s", p.curToken.Line, p.curToken.Column, p.curToken.Type))
			return nil
		}
		param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		if p.peekTokenIs(lexer.COLON) {
			p.nextToken() // consume ':'
			if !p.expectPeek(lexer.IDENT) {
				return nil
			}
			paramType := p.curToken.Literal
			for p.peekTokenIs(lexer.DOT) {
				p.nextToken() // consume '.'
				if !p.expectPeek(lexer.IDENT) {
					return nil
				}
				paramType += "." + p.curToken.Literal
			}
			param.Type = paramType
		}

		params = append(params, param)

		if p.peekTokenIs(lexer.RPAREN) {
			p.nextToken()
			break
		}

		if !p.expectPeek(lexer.COMMA) {
			return nil
		}
		p.nextToken()
	}

	return params
}

func (p *Parser) parseBlockStatement() *ast.BlockStmt {
	block := &ast.BlockStmt{Token: p.curToken}
	block.Statements = []ast.Statement{}

	if p.curTokenIs(lexer.INDENT) {
		p.nextToken()
	}

	for !p.curTokenIs(lexer.DEDENT) && !p.curTokenIs(lexer.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		if isSimpleStatement(stmt) {
			p.nextToken()
		}
		for p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}
	}

	if p.curTokenIs(lexer.DEDENT) {
		p.nextToken()
	}

	return block
}

func isSimpleStatement(stmt ast.Statement) bool {
	if stmt == nil {
		return false
	}
	switch stmt.(type) {
	case *ast.DefStmt, *ast.IfStmt, *ast.ForStmt, *ast.WhileStmt, *ast.ClassStmt, *ast.MatchStmt:
		return false
	default:
		return true
	}
}

func (p *Parser) parseIfStatement() *ast.IfStmt {
	stmt := &ast.IfStmt{Token: p.curToken}
	p.nextToken() // consume 'if'
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.COLON) {
		return nil
	}
	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}
	if !p.expectPeek(lexer.INDENT) {
		return nil
	}
	stmt.Consequence = p.parseBlockStatement()

	for p.curTokenIs(lexer.ELIF) {
		elifClause := ast.ElifClause{}
		p.nextToken() // consume 'elif'
		elifClause.Condition = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		if !p.expectPeek(lexer.NEWLINE) {
			return nil
		}
		if !p.expectPeek(lexer.INDENT) {
			return nil
		}
		elifClause.Consequence = p.parseBlockStatement()
		stmt.Elifs = append(stmt.Elifs, elifClause)
	}

	if p.curTokenIs(lexer.ELSE) {
		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		if !p.expectPeek(lexer.NEWLINE) {
			return nil
		}
		if !p.expectPeek(lexer.INDENT) {
			return nil
		}
		stmt.Alternative = p.parseBlockStatement()
	}

	return stmt
}

func (p *Parser) parseWhileStatement() *ast.WhileStmt {
	stmt := &ast.WhileStmt{Token: p.curToken}
	p.nextToken() // consume 'while'
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.COLON) {
		return nil
	}
	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}
	if !p.expectPeek(lexer.INDENT) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStmt {
	stmt := &ast.ForStmt{Token: p.curToken}
	p.nextToken() // consume 'for'

	stmt.Vars = []*ast.Identifier{}
	for {
		if !p.curTokenIs(lexer.IDENT) {
			p.errors = append(p.errors, fmt.Sprintf("[%d:%d] expected identifier in for loop variables, got %s", p.curToken.Line, p.curToken.Column, p.curToken.Type))
			return nil
		}
		stmt.Vars = append(stmt.Vars, &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal})

		if p.peekTokenIs(lexer.COMMA) {
			p.nextToken() // consume ','
			p.nextToken() // move to next identifier
		} else {
			break
		}
	}

	if !p.expectPeek(lexer.IN) {
		return nil
	}
	p.nextToken() // consume 'in'
	stmt.Iterable = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.COLON) {
		return nil
	}
	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}
	if !p.expectPeek(lexer.INDENT) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStmt {
	stmt := &ast.ReturnStmt{Token: p.curToken}
	if p.peekTokenIs(lexer.NEWLINE) || p.peekTokenIs(lexer.EOF) {
		p.nextToken()
		return stmt
	}
	p.nextToken() // consume 'return'
	stmt.Value = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseClassStatement() *ast.ClassStmt {
	stmt := &ast.ClassStmt{Token: p.curToken}
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(lexer.COLON) {
		return nil
	}
	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}
	if !p.expectPeek(lexer.INDENT) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseMatchStatement() *ast.MatchStmt {
	stmt := &ast.MatchStmt{Token: p.curToken}
	p.nextToken() // consume 'match'
	stmt.Subject = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.COLON) {
		return nil
	}
	if !p.expectPeek(lexer.NEWLINE) {
		return nil
	}
	if !p.expectPeek(lexer.INDENT) {
		return nil
	}

	stmt.Cases = []ast.MatchCase{}
	for p.curTokenIs(lexer.CASE) {
		mc := ast.MatchCase{}
		p.nextToken() // consume 'case'
		mc.Pattern = p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		if !p.expectPeek(lexer.NEWLINE) {
			return nil
		}
		if !p.expectPeek(lexer.INDENT) {
			return nil
		}
		mc.Body = p.parseBlockStatement()
		stmt.Cases = append(stmt.Cases, mc)

		if p.curTokenIs(lexer.NEWLINE) {
			p.nextToken()
		}
	}

	if p.curTokenIs(lexer.DEDENT) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionOrAssignmentStatement() ast.Statement {
	startToken := p.curToken
	expr := p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.ASSIGN) {
		p.nextToken() // move to '='
		assignTok := p.curToken
		p.nextToken() // consume '='
		val := p.parseExpression(LOWEST)
		return &ast.AssignStmt{
			Token: assignTok,
			Left:  expr,
			Value: val,
		}
	}

	return &ast.ExprStmt{
		Token:      startToken,
		Expression: expr,
	}
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(lexer.NEWLINE) && !p.peekTokenIs(lexer.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseNumberLiteral() ast.Expression {
	return &ast.NumberLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseFStringLiteral() ast.Expression {
	return &ast.FStringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curTokenIs(lexer.TRUE)}
}

func (p *Parser) parseNoneLiteral() ast.Expression {
	return &ast.NoneLiteral{Token: p.curToken}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpr{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}
	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if !p.expectPeek(lexer.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseListLiteral() ast.Expression {
	lit := &ast.ListLiteral{Token: p.curToken}
	lit.Elements = p.parseExpressionList(lexer.RBRACKET)
	return lit
}

func (p *Parser) parseDictLiteral() ast.Expression {
	lit := &ast.DictLiteral{Token: p.curToken}
	lit.Keys = []ast.Expression{}
	lit.Value = []ast.Expression{}

	if p.peekTokenIs(lexer.RBRACE) {
		p.nextToken()
		return lit
	}

	p.nextToken() // consume '{'
	for {
		key := p.parseExpression(LOWEST)
		if !p.expectPeek(lexer.COLON) {
			return nil
		}
		p.nextToken() // consume ':'
		val := p.parseExpression(LOWEST)

		lit.Keys = append(lit.Keys, key)
		lit.Value = append(lit.Value, val)

		if p.peekTokenIs(lexer.RBRACE) {
			p.nextToken()
			break
		}
		if !p.expectPeek(lexer.COMMA) {
			return nil
		}
		p.nextToken() // consume ','
	}
	return lit
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpr{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	exp := &ast.CallExpr{Token: p.curToken, Function: left}
	exp.Arguments = []ast.Expression{}
	exp.Keywords = []ast.KeywordArg{}

	if p.peekTokenIs(lexer.RPAREN) {
		p.nextToken()
		return exp
	}

	p.nextToken()

	for {
		isKeyword := p.curTokenIs(lexer.IDENT) && p.peekTokenIs(lexer.ASSIGN)

		if isKeyword {
			ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
			p.nextToken() // move to '='
			p.nextToken() // consume '='
			val := p.parseExpression(LOWEST)
			exp.Keywords = append(exp.Keywords, ast.KeywordArg{Name: ident, Value: val})
		} else {
			arg := p.parseExpression(LOWEST)
			exp.Arguments = append(exp.Arguments, arg)
		}

		if p.peekTokenIs(lexer.RPAREN) {
			p.nextToken()
			break
		}
		if !p.expectPeek(lexer.COMMA) {
			return nil
		}
		p.nextToken()
	}
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpr{Token: p.curToken, Left: left}
	p.nextToken() // consume '['
	exp.Index = p.parseExpression(LOWEST)
	if !p.expectPeek(lexer.RBRACKET) {
		return nil
	}
	return exp
}

func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
	exp := &ast.DotExpr{Token: p.curToken, Left: left}
	if !p.expectPeek(lexer.IDENT) {
		return nil
	}
	exp.Member = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	return exp
}

func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := []ast.Expression{}
	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}
	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	for p.peekTokenIs(lexer.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(end) {
		return nil
	}
	return list
}
