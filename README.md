# PyStatic

A Python-style language that compiles to Go. Write Python, ship Go binaries.

[![License: BSD-3-Clause](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](LICENSE)
[![Support the project](https://img.shields.io/badge/Support-the%20project-ff69b4?logo=github-sponsors&logoColor=white)](https://apoie.pedrodev.top)

---

## Why PyStatic?

- **Python syntax** you already know — `def`, `for`, `if`, f-strings, type hints.
- **Go performance** — PyStatic transpiles to idiomatic Go and compiles to a single static binary.
- **Zero runtime overhead** — no interpreter, no GC pauses from a scripting runtime.
- **Built-in modules** — `http`, `json`, `csv`, `os`, `time`, `math` all map to the Go standard library.

---

## Installation

```bash
git clone https://github.com/pedrodev2026/pystatic
cd pystatic
go build -o pystatic cmd/pystatic/main.go
sudo mv pystatic /usr/local/bin/
```

---

## Quick Start

**`hello.pstic`**
```python
def main():
    print("Hello from PyStatic!")
    name = f"World"
    print(f"Hello, {name}!")
```

```bash
pystatic run hello.pstic
# → Hello from PyStatic!
# → Hello, World!
```

---

## Language Tour

### Variables & Types

```python
x = 42
pi = 3.14
name = "Alice"
active = True
nothing = None
```

### Functions & Type Hints

```python
def greet(name: str) -> str:
    return f"Hello, {name}!"

def main():
    msg = greet("World")
    print(msg)
```

### Control Flow

```python
def main():
    x = 10
    if x > 5:
        print("big")
    elif x == 5:
        print("five")
    else:
        print("small")
```

### Loops

```python
def main():
    items = [1, 2, 3]
    for i, v in items:
        print(f"index={i} value={v}")

    n = 0
    while n < 5:
        print(n)
        n = n + 1
```

### Collections

```python
def main():
    fruits = ["apple", "banana", "cherry"]
    config = {"host": "localhost", "port": 8080}
    print(config["host"])
```

### HTTP Server

```python
import http

def main():
    server = http.Server(addr=":3000")

    def hello(req: http.Request) -> http.Response:
        return http.Response(status=200, body="Hello, World!")

    server.handle_func("/hello", hello)
    print("Listening on :3000")
    server.listen_and_serve()
```

### JSON

```python
import json

def main():
    data = json.encode({"message": "ok", "code": 200})
    print(data)
```

### F-Strings

F-strings are compiled to `fmt.Sprintf` automatically:

```python
name = "PyStatic"
version = "1.0"
msg = f"Welcome to {name} v{version}"
```

---

## CLI Commands

| Command | Description |
|---------|-------------|
| `pystatic run <file.pstic>` | Transpile + run immediately |
| `pystatic run <file.pstic> --tokens` | Print lexer tokens (debug) |
| `pystatic run <file.pstic> --ast` | Print parsed AST (debug) |
| `pystatic build` | Build binary from `pstic-pkg.json` |
| `pystatic build -o mybin file.pstic` | Build a single file to a binary |
| `pystatic new <project-name>` | Create a new project scaffold |
| `pystatic add mod <module>` | Add a module to `pstic-pkg.json` |

---

## Project Layout

```
pystatic/
├── cmd/pystatic/main.go        # CLI entrypoint
├── internal/
│   ├── lexer/                  # Tokenizer
│   ├── ast/                    # AST node types
│   ├── parser/                 # Pratt recursive descent parser
│   └── codegen/                # Go code generator
├── pkg/runtime/                # Built-in module runtime wrappers
│   ├── runtime.go              # print()
│   ├── http.go                 # http.Server, Request, Response
│   ├── json.go                 # json.encode / json.decode
│   ├── csv.go                  # CSV reader/writer
│   ├── os.go                   # os.args, os.getenv, os.exit
│   ├── time.go                 # time.now, time.sleep
│   └── math.go                 # math.sqrt, math.abs, math.pow
├── examples/
│   ├── hello.pstic
│   └── http_server.pstic
└── pstic-pkg.json              # Project configuration
```

---

## `pstic-pkg.json`

```json
{
  "name": "myapp",
  "version": "0.1.0",
  "entry": "main.pstic",
  "modules": ["http", "json"]
}
```

---

## Compilation Pipeline

```
.pstic → [Lexer] → Tokens → [Parser] → AST → [Codegen] → .go → go build → binary
```

---

## Type Mapping

| PyStatic | Go |
|----------|----|
| `int` | `int` |
| `float` | `float64` |
| `str` | `string` |
| `bool` | `bool` |
| `None` | `nil` |
| `list` | `[]any` |
| `dict` | `map[any]any` |
| `-> T` | Go return type |

---

## Built-in Modules

### `http`
- `http.Server(addr=":8080")` — creates a new HTTP server
- `def handler(req: http.Request) -> http.Response` — HTTP handler function signature
- `http.Response(status=200, body="ok", headers={...})` — response constructor
- `server.handle_func("/path", handler)` — register a route
- `server.listen_and_serve()` — start the server

### `json`
- `json.encode(value)` — serialize to JSON string
- `json.decode(data, &target)` — deserialize JSON string

### `os`
- `os.args()` — command-line arguments
- `os.getenv("KEY")` — environment variable
- `os.exit(code)` — exit with code

### `time`
- `time.now()` — current time
- `time.sleep(seconds)` — sleep

### `math`
- `math.sqrt(x)`, `math.abs(x)`, `math.pow(x, y)`

---

## Running Tests

```bash
go test ./...
```

---

## License

BSD-3-Clause. See [LICENSE](LICENSE).
