package codegen

import (
	"strings"
	"testing"

	"github.com/pedrodev2026/pystatic/internal/lexer"
	"github.com/pedrodev2026/pystatic/internal/parser"
)

func TestTranspileSimple(t *testing.T) {
	input := `def main():
    print("hello")`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	gen := New(prog)
	out := gen.Generate()

	expected := `package main

import (
    rth "github.com/pedrodev2026/pystatic/pkg/runtime"
)

func main() {
    rth.Println("hello")
}`

	if normalizeWhitespace(out) != normalizeWhitespace(expected) {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, out)
	}
}

func TestTranspileHTTP(t *testing.T) {
	input := `import http

def main():
    server = http.Server(addr=":3000")

    def hello(req: http.Request) -> http.Response:
        return http.Response(status=200, body="ok")

    server.handle_func("/hello", hello)`

	l := lexer.New(input)
	p := parser.New(l)
	prog := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	gen := New(prog)
	out := gen.Generate()

	expected := `package main

import (
    rth "github.com/pedrodev2026/pystatic/pkg/runtime"
)

func main() {
    server := rth.NewServer(":3000")
    hello := func(w rth.ResponseWriter, r *rth.Request) {
        w.WriteHeader(200)
        w.Write([]byte("ok"))
        return
    }
    server.HandleFunc("/hello", hello)
}`

	if normalizeWhitespace(out) != normalizeWhitespace(expected) {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, out)
	}
}

func normalizeWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return strings.Join(result, "\n")
}
