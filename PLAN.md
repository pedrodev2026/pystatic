# PyStatic — Plano de Implementação

Linguagem Python-style que transpila para Go. BSD-3-Clause.

## Decisões de Design

| Item | Escolha |
|------|---------|
| Palavra-chave função | `def` |
| Indentação | 4 espaços (PEP 8) |
| Variáveis | Inferência estilo Python (`x = 5`) |
| Módulos built-in | `http`, `json`, `csv`, `os`, `time`, `math` |
| Extensão | `.pstic` |
| Config de pacote | `pstic-pkg.json` |
| Entrega | Fase por fase |

## Estrutura do Projeto

```
/home/pedro/pystatic/
├── .gitignore
├── LICENSE
├── PLAN.md
├── README.md
├── go.mod
├── go.sum
├── pstic-pkg.json
├── cmd/
│   └── pystatic/
│       └── main.go
├── internal/
│   ├── ast/
│   │   └── ast.go
│   ├── lexer/
│   │   ├── token.go
│   │   ├── lexer.go
│   │   └── lexer_test.go
│   ├── parser/
│   │   ├── parser.go
│   │   └── parser_test.go
│   ├── codegen/
│   │   ├── codegen.go
│   │   ├── imports.go
│   │   └── codegen_test.go
│   ├── module/
│   │   ├── manager.go
│   │   └── manager_test.go
│   └── builtins/
│       ├── http_builtin.go
│       ├── json_builtin.go
│       ├── csv_builtin.go
│       ├── os_builtin.go
│       ├── time_builtin.go
│       └── math_builtin.go
├── pkg/
│   └── runtime/
│       ├── runtime.go
│       ├── http.go
│       ├── json.go
│       ├── csv.go
│       ├── os.go
│       ├── time.go
│       └── math.go
├── examples/
│   ├── hello.pstic
│   ├── http_server.pstic
│   ├── json_demo.pstic
│   ├── csv_read.pstic
│   ├── os_demo.pstic
│   ├── time_demo.pstic
│   └── math_demo.pstic
└── tests/
    ├── lexer_test.go
    ├── parser_test.go
    ├── codegen_test.go
    └── integration_test.go
```

## Pipeline de Compilação

```
.pstic → [Lexer] → tokens → [Parser] → AST → [Codegen] → .go → [go build] → binário
```

## Mapeamento de Tipos

| PyStatic | Go |
|----------|-----|
| `int` | `int` |
| `float` | `float64` |
| `str` | `string` |
| `bool` | `bool` |
| `None` | `nil` |
| `list[T]` | `[]T` |
| `dict[K,V]` | `map[K]V` |
| `-> T` | `T` no retorno |

## Sintaxe

```python
import http
import json

def handler(req: http.Request) -> http.Response:
    data = json.encode({"message": "PyStatic!"})
    return http.Response(status=200, body=data, headers={"Content-Type": "application/json"})

def main():
    server = http.Server(addr=":3000")
    server.handle_func("/api/hello", handler)
    print("Servidor rodando em :3000")
    server.listen_and_serve()
```

## Fases

### Fase 1 — Fundação (Lexer + CLI)
- `go mod init`, estrutura de diretórios
- CLI com `flag` nativo
- Lexer com todos os tokens
- Tokens: `IDENT`, `NUMBER`, `STRING`, `FSTRING`, `INDENT`, `DEDENT`, `NEWLINE`, `EOF`
- Operadores: `+`, `-`, `*`, `/`, `==`, `!=`, `<`, `>`, `<=`, `>=`, `=`, `:`, `->`, `,`, `.`, `(`, `)`, `[`, `]`, `{`, `}`
- Keywords: `def`, `if`, `elif`, `else`, `while`, `for`, `in`, `return`, `import`, `class`, `match`, `case`, `break`, `continue`, `None`, `True`, `False`, `and`, `or`, `not`
- Comando: `pystatic run exemplo.pstic --tokens`

### Fase 2 — Parser (AST)
- AST nodes: `Program`, `DefStmt`, `IfStmt`, `ForStmt`, `WhileStmt`, `ReturnStmt`, `ExprStmt`, `ImportStmt`, `AssignStmt`, `ClassStmt`, `MatchStmt`
- Expressões: `BinaryExpr`, `UnaryExpr`, `CallExpr`, `Ident`, `Number`, `String`, `FString`, `List`, `Dict`
- Parser recursive descent com controle de indentação via stack
- Comando: `pystatic run exemplo.pstic --ast`

### Fase 3 — Code Generator (Transpiler)
- AST → código Go válido
- F-strings → `fmt.Sprintf`
- Geração de arquivo `.go` temporário
- `go run`/`go build` automático
- Comando: `pystatic run exemplo.pstic`

### Fase 4 — Runtime + Módulos Built-in
- `http`: wrapper `net/http` (Server, Request, Response, HandleFunc, ListenAndServe)
- `json`: wrapper `encoding/json` (encode, decode)
- `csv`: wrapper `encoding/csv` (Reader, Writer)
- `os`: wrapper `os` (args, env, file, exit)
- `time`: wrapper `time` (now, sleep, duration, format)
- `math`: wrapper `math` (abs, sin, cos, sqrt, pow, max, min)

### Fase 5 — Sistema de Pacotes
- Leitura/validação de `pstic-pkg.json`
- `pystatic build` com build multi-arquivo
- `pystatic new nome-projeto` (scaffold)
- `pystatic add mod http`

### Fase 6 — Polimento
- `pystatic test` (testes de integração)
- CI (GitHub Actions: lint + testes + build)
- README.md completo
- `go vet` + formatação
- Mensagens de erro com linha/coluna

## Comandos CLI

```
pystatic run arquivo.pstic            # transpila + go run
pystatic build                        # lê pstic-pkg.json, compila tudo
pystatic build -o saida arquivo.pstic
pystatic new nome-projeto             # scaffold
pystatic add mod http                 # adiciona módulo
pystatic run exemplo.pstic --tokens   # debug: mostra tokens
pystatic run exemplo.pstic --ast      # debug: mostra AST
```

## Exemplo de Transpilação

**Entrada (`.pstic`):**
```python
import http
import json

def main():
    server = http.Server(addr=":3000")

    def hello(req: http.Request) -> http.Response:
        data = json.encode({"message": "Hello, World!"})
        return http.Response(
            status=200,
            body=data,
            headers={"Content-Type": "application/json"}
        )

    server.handle_func("/api/hello", hello)
    print("API rodando em :3000")
    server.listen_and_serve()
```

**Saída gerada (`.go`):**
```go
package main

import (
    rth "github.com/pedrodev2026/pystatic/pkg/runtime"
)

func main() {
    server := rth.NewServer(":3000")
    server.HandleFunc("/api/hello", func(w rth.ResponseWriter, r *rth.Request) {
        data := `{"message": "Hello, World!"}`
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(200)
        w.Write([]byte(data))
    })
    rth.Println("API rodando em :3000")
    rth.ListenAndServe(":3000", nil)
}
```
