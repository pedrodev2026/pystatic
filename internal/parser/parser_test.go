package parser

import (
	"testing"

	"github.com/pedrodev2026/pystatic/internal/ast"
	"github.com/pedrodev2026/pystatic/internal/lexer"
)

func TestDefStatement(t *testing.T) {
	input := `def add(x: int, y: int) -> int:
    return x + y`

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()
	checkParserErrors(t, p)

	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}

	defStmt, ok := prog.Statements[0].(*ast.DefStmt)
	if !ok {
		t.Fatalf("expected DefStmt, got %T", prog.Statements[0])
	}

	if defStmt.Name.Value != "add" {
		t.Errorf("expected function name 'add', got '%s'", defStmt.Name.Value)
	}

	if len(defStmt.Parameters) != 2 {
		t.Fatalf("expected 2 parameters, got %d", len(defStmt.Parameters))
	}

	if defStmt.Parameters[0].Name.Value != "x" || defStmt.Parameters[0].Type != "int" {
		t.Errorf("expected parameter x: int, got %s: %s", defStmt.Parameters[0].Name.Value, defStmt.Parameters[0].Type)
	}

	if defStmt.ReturnType != "int" {
		t.Errorf("expected return type 'int', got '%s'", defStmt.ReturnType)
	}
}

func TestIfStatement(t *testing.T) {
	input := `if x == 5:
    return 1
elif x == 6:
    return 2
else:
    return 3`

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()
	checkParserErrors(t, p)

	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}

	ifStmt, ok := prog.Statements[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("expected IfStmt, got %T", prog.Statements[0])
	}

	// Verify condition
	infix, ok := ifStmt.Condition.(*ast.InfixExpr)
	if !ok {
		t.Fatalf("expected infix expression for if condition, got %T", ifStmt.Condition)
	}
	if infix.Operator != "==" {
		t.Errorf("expected operator '==', got '%s'", infix.Operator)
	}

	if len(ifStmt.Elifs) != 1 {
		t.Fatalf("expected 1 elif clause, got %d", len(ifStmt.Elifs))
	}

	if ifStmt.Alternative == nil {
		t.Fatalf("expected else clause to be present")
	}
}

func TestCallExpression(t *testing.T) {
	input := `func(1, 2, status=200, body=data)`

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()
	checkParserErrors(t, p)

	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}

	exprStmt, ok := prog.Statements[0].(*ast.ExprStmt)
	if !ok {
		t.Fatalf("expected ExprStmt, got %T", prog.Statements[0])
	}

	call, ok := exprStmt.Expression.(*ast.CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", exprStmt.Expression)
	}

	if len(call.Arguments) != 2 {
		t.Errorf("expected 2 positional arguments, got %d", len(call.Arguments))
	}

	if len(call.Keywords) != 2 {
		t.Errorf("expected 2 keyword arguments, got %d", len(call.Keywords))
	}

	if call.Keywords[0].Name.Value != "status" {
		t.Errorf("expected keyword arg 'status', got '%s'", call.Keywords[0].Name.Value)
	}
}

func TestForAndWhile(t *testing.T) {
	input := `for x, y in pairs:
    while x < y:
        x = x + 1`

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()
	checkParserErrors(t, p)

	if len(prog.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(prog.Statements))
	}

	forStmt, ok := prog.Statements[0].(*ast.ForStmt)
	if !ok {
		t.Fatalf("expected ForStmt, got %T", prog.Statements[0])
	}

	if len(forStmt.Vars) != 2 {
		t.Fatalf("expected 2 loop variables, got %d", len(forStmt.Vars))
	}

	if forStmt.Vars[0].Value != "x" || forStmt.Vars[1].Value != "y" {
		t.Errorf("incorrect loop variables: %v", forStmt.Vars)
	}

	if len(forStmt.Body.Statements) != 1 {
		t.Fatalf("expected 1 statement in for body, got %d", len(forStmt.Body.Statements))
	}

	whileStmt, ok := forStmt.Body.Statements[0].(*ast.WhileStmt)
	if !ok {
		t.Fatalf("expected WhileStmt, got %T", forStmt.Body.Statements[0])
	}

	if whileStmt.Condition.String() != "(x < y)" {
		t.Errorf("incorrect while condition: %s", whileStmt.Condition.String())
	}
}

func TestDictAndList(t *testing.T) {
	input := `x = [1, 2, 3]
y = {"a": 1, "b": 2}`

	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()
	checkParserErrors(t, p)

	if len(prog.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(prog.Statements))
	}

	// Verify list assignment
	assignList, ok := prog.Statements[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", prog.Statements[0])
	}

	listLit, ok := assignList.Value.(*ast.ListLiteral)
	if !ok {
		t.Fatalf("expected ListLiteral, got %T", assignList.Value)
	}

	if len(listLit.Elements) != 3 {
		t.Errorf("expected 3 elements, got %d", len(listLit.Elements))
	}

	// Verify dict assignment
	assignDict, ok := prog.Statements[1].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected AssignStmt, got %T", prog.Statements[1])
	}

	dictLit, ok := assignDict.Value.(*ast.DictLiteral)
	if !ok {
		t.Fatalf("expected DictLiteral, got %T", assignDict.Value)
	}

	if len(dictLit.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(dictLit.Keys))
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
