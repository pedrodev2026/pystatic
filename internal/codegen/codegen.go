package codegen

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/pedrodev2026/pystatic/internal/ast"
)

type Scope struct {
	parent *Scope
	vars   map[string]bool
}

func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent: parent,
		vars:   make(map[string]bool),
	}
}

func (s *Scope) Define(name string) {
	s.vars[name] = true
}

func (s *Scope) IsDefined(name string) bool {
	curr := s
	for curr != nil {
		if curr.vars[name] {
			return true
		}
		curr = curr.parent
	}
	return false
}

type Generator struct {
	prog          *ast.Program
	buf           bytes.Buffer
	indent        int
	scope         *Scope
	inHTTPHandler bool
	imports       map[string]bool
}

func New(prog *ast.Program) *Generator {
	return &Generator{
		prog:    prog,
		scope:   NewScope(nil),
		imports: make(map[string]bool),
	}
}

func (g *Generator) getIndent() string {
	return strings.Repeat("    ", g.indent)
}

func (g *Generator) write(s string) {
	g.buf.WriteString(s)
}

func (g *Generator) writeln(s string) {
	g.buf.WriteString(s + "\n")
}

func (g *Generator) Generate() string {
	// First pass: generate the body to see what imports/runtime we need
	var bodyBuf bytes.Buffer
	oldBuf := g.buf
	g.buf = bodyBuf
	
	for _, stmt := range g.prog.Statements {
		g.write(g.getIndent())
		g.transpileStatement(stmt)
		g.writeln("")
	}
	
	bodyCode := g.buf.String()
	g.buf = oldBuf

	// Header
	g.writeln("package main")
	g.writeln("")
	g.writeln("import (")
	// Always include runtime import for builtins
	g.writeln("    rth \"github.com/pedrodev2026/pystatic/pkg/runtime\"")
	if g.imports["fmt"] {
		g.writeln(`    "fmt"`)
	}
	g.writeln(")")
	g.writeln("")
	g.write(bodyCode)

	return g.buf.String()
}

func (g *Generator) transpileStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ImportStmt:
		// PyStatic built-in imports are handled automatically, so we skip generating them in Go
		return
	case *ast.DefStmt:
		g.transpileDef(s)
	case *ast.ReturnStmt:
		g.transpileReturn(s)
	case *ast.AssignStmt:
		g.transpileAssign(s)
	case *ast.IfStmt:
		g.transpileIf(s)
	case *ast.ForStmt:
		g.transpileFor(s)
	case *ast.WhileStmt:
		g.transpileWhile(s)
	case *ast.ExprStmt:
		g.write(g.transpileExpression(s.Expression))
	}
}

func (g *Generator) transpileDef(s *ast.DefStmt) {
	isMain := s.Name.Value == "main"
	
	// Check if this is an HTTP handler
	isHTTPHandler := false
	if len(s.Parameters) == 1 && s.Parameters[0].Type == "http.Request" && s.ReturnType == "http.Response" {
		isHTTPHandler = true
	}

	// Nested function inside another function -> Closure
	isNested := g.indent > 0

	if isNested {
		g.write(s.Name.Value + " := func(")
	} else {
		g.write("func " + s.Name.Value + "(")
	}

	// Parameters
	g.scope = NewScope(g.scope) // New scope for parameters and body
	
	if isHTTPHandler {
		g.write("w rth.ResponseWriter, r *rth.Request")
		g.scope.Define("w")
		g.scope.Define("r")
	} else {
		var params []string
		for _, p := range s.Parameters {
			g.scope.Define(p.Name.Value)
			goType := g.mapType(p.Type)
			params = append(params, p.Name.Value+" "+goType)
		}
		g.write(strings.Join(params, ", "))
	}
	g.write(")")

	// Return type
	if !isMain && !isHTTPHandler && s.ReturnType != "" {
		g.write(" " + g.mapType(s.ReturnType))
	}

	g.writeln(" {")
	
	// Save previous handler state
	oldHandler := g.inHTTPHandler
	g.inHTTPHandler = isHTTPHandler

	g.indent++
	for _, child := range s.Body.Statements {
		g.write(g.getIndent())
		g.transpileStatement(child)
		g.writeln("")
	}
	g.indent--
	
	g.inHTTPHandler = oldHandler
	g.scope = g.scope.parent // pop scope
	
	g.write(g.getIndent() + "}")
}

func (g *Generator) transpileReturn(s *ast.ReturnStmt) {
	if s.Value == nil {
		g.write("return")
		return
	}

	// If inside HTTP handler, return http.Response is special
	if g.inHTTPHandler {
		if call, ok := s.Value.(*ast.CallExpr); ok {
			if dot, ok := call.Function.(*ast.DotExpr); ok {
				if dot.Left.String() == "http" && dot.Member.Value == "Response" {
					g.write(g.transpileHTTPResponse(call))
					return
				}
			}
		}
	}

	g.write("return " + g.transpileExpression(s.Value))
}

func (g *Generator) transpileHTTPResponse(call *ast.CallExpr) string {
	status := "200"
	body := `""`
	var headers []string

	for _, kw := range call.Keywords {
		switch kw.Name.Value {
		case "status":
			status = g.transpileExpression(kw.Value)
		case "body":
			body = g.transpileExpression(kw.Value)
		case "headers":
			if dict, ok := kw.Value.(*ast.DictLiteral); ok {
				for i := range dict.Keys {
					k := g.transpileExpression(dict.Keys[i])
					v := g.transpileExpression(dict.Value[i])
					headers = append(headers, fmt.Sprintf("w.Header().Set(%s, %s)", k, v))
				}
			}
		}
	}

	var sb strings.Builder
	for _, h := range headers {
		sb.WriteString(h + "\n" + g.getIndent())
	}
	sb.WriteString(fmt.Sprintf("w.WriteHeader(%s)\n", status))
	sb.WriteString(g.getIndent() + fmt.Sprintf("w.Write([]byte(%s))\n", body))
	sb.WriteString(g.getIndent() + "return")
	return sb.String()
}

func (g *Generator) transpileAssign(s *ast.AssignStmt) {
	left := g.transpileExpression(s.Left)
	val := g.transpileExpression(s.Value)

	// Determine if we need := or =
	op := " = "
	if ident, ok := s.Left.(*ast.Identifier); ok {
		if !g.scope.IsDefined(ident.Value) {
			g.scope.Define(ident.Value)
			op = " := "
		}
	}
	g.write(left + op + val)
}

func (g *Generator) transpileIf(s *ast.IfStmt) {
	g.write("if " + g.transpileExpression(s.Condition) + " {\n")
	g.indent++
	for _, child := range s.Consequence.Statements {
		g.write(g.getIndent())
		g.transpileStatement(child)
		g.writeln("")
	}
	g.indent--
	g.write(g.getIndent() + "}")

	for _, elif := range s.Elifs {
		g.write(" else if " + g.transpileExpression(elif.Condition) + " {\n")
		g.indent++
		for _, child := range elif.Consequence.Statements {
			g.write(g.getIndent())
			g.transpileStatement(child)
			g.writeln("")
		}
		g.indent--
		g.write(g.getIndent() + "}")
	}

	if s.Alternative != nil {
		g.write(" else {\n")
		g.indent++
		for _, child := range s.Alternative.Statements {
			g.write(g.getIndent())
			g.transpileStatement(child)
			g.writeln("")
		}
		g.indent--
		g.write(g.getIndent() + "}")
	}
}

func (g *Generator) transpileFor(s *ast.ForStmt) {
	var vars []string
	g.scope = NewScope(g.scope)
	for _, v := range s.Vars {
		g.scope.Define(v.Value)
		vars = append(vars, v.Value)
	}

	g.write("for " + strings.Join(vars, ", ") + " := range " + g.transpileExpression(s.Iterable) + " {\n")
	g.indent++
	for _, child := range s.Body.Statements {
		g.write(g.getIndent())
		g.transpileStatement(child)
		g.writeln("")
	}
	g.indent--
	g.scope = g.scope.parent
	g.write(g.getIndent() + "}")
}

func (g *Generator) transpileWhile(s *ast.WhileStmt) {
	g.write("for " + g.transpileExpression(s.Condition) + " {\n")
	g.indent++
	for _, child := range s.Body.Statements {
		g.write(g.getIndent())
		g.transpileStatement(child)
		g.writeln("")
	}
	g.indent--
	g.write(g.getIndent() + "}")
}

func (g *Generator) transpileExpression(expr ast.Expression) string {
	if expr == nil {
		return ""
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		return e.Value
	case *ast.NumberLiteral:
		return e.Value
	case *ast.StringLiteral:
		return fmt.Sprintf("%q", e.Value)
	case *ast.FStringLiteral:
		g.imports["fmt"] = true
		return g.transpileFString(e.Value)
	case *ast.BooleanLiteral:
		return fmt.Sprintf("%t", e.Value)
	case *ast.NoneLiteral:
		return "nil"
	case *ast.PrefixExpr:
		op := e.Operator
		if op == "not" {
			op = "!"
		}
		return op + g.transpileExpression(e.Right)
	case *ast.InfixExpr:
		op := e.Operator
		if op == "and" {
			op = "&&"
		} else if op == "or" {
			op = "||"
		}
		return fmt.Sprintf("(%s %s %s)", g.transpileExpression(e.Left), op, g.transpileExpression(e.Right))
	case *ast.DotExpr:
		left := g.transpileExpression(e.Left)
		member := toPascalCase(e.Member.Value)
		if left == "http" && member == "Server" {
			return "rth.Server"
		}
		if left == "http" {
			return "rth." + member
		}
		return left + "." + member
	case *ast.CallExpr:
		return g.transpileCall(e)
	case *ast.ListLiteral:
		var elems []string
		for _, el := range e.Elements {
			elems = append(elems, g.transpileExpression(el))
		}
		return "[]any{" + strings.Join(elems, ", ") + "}"
	case *ast.DictLiteral:
		var pairs []string
		for i := range e.Keys {
			pairs = append(pairs, g.transpileExpression(e.Keys[i])+": "+g.transpileExpression(e.Value[i]))
		}
		return "map[any]any{" + strings.Join(pairs, ", ") + "}"
	case *ast.IndexExpr:
		return fmt.Sprintf("%s[%s]", g.transpileExpression(e.Left), g.transpileExpression(e.Index))
	}
	return ""
}

func (g *Generator) transpileCall(e *ast.CallExpr) string {
	funStr := g.transpileExpression(e.Function)
	
	if funStr == "print" {
		var args []string
		for _, arg := range e.Arguments {
			args = append(args, g.transpileExpression(arg))
		}
		return "rth.Println(" + strings.Join(args, ", ") + ")"
	}

	if dot, ok := e.Function.(*ast.DotExpr); ok {
		if dot.Left.String() == "http" && dot.Member.Value == "Server" {
			addr := `":8080"`
			for _, kw := range e.Keywords {
				if kw.Name.Value == "addr" {
					addr = g.transpileExpression(kw.Value)
				}
			}
			return "rth.NewServer(" + addr + ")"
		}
	}

	var args []string
	for _, arg := range e.Arguments {
		args = append(args, g.transpileExpression(arg))
	}
	
	return funStr + "(" + strings.Join(args, ", ") + ")"
}

func (g *Generator) transpileFString(lit string) string {
	var format strings.Builder
	var args []string
	i := 0
	for i < len(lit) {
		if lit[i] == '{' {
			i++
			start := i
			for i < len(lit) && lit[i] != '}' {
				i++
			}
			expr := lit[start:i]
			args = append(args, expr)
			format.WriteString("%v")
			i++
		} else {
			format.WriteByte(lit[i])
			i++
		}
	}
	if len(args) == 0 {
		return fmt.Sprintf("%q", format.String())
	}
	return fmt.Sprintf("fmt.Sprintf(%q, %s)", format.String(), strings.Join(args, ", "))
}

func (g *Generator) mapType(pyType string) string {
	switch pyType {
	case "int":
		return "int"
	case "float":
		return "float64"
	case "str":
		return "string"
	case "bool":
		return "bool"
	case "None":
		return "any"
	case "http.Request":
		return "*rth.Request"
	case "http.Response":
		return "rth.Response"
	default:
		if strings.Contains(pyType, ".") {
			parts := strings.Split(pyType, ".")
			return parts[0] + "." + toPascalCase(parts[1])
		}
		return pyType
	}
}

func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	parts := strings.Split(s, "_")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}
