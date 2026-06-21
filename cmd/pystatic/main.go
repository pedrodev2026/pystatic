package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pedrodev2026/pystatic/internal/codegen"
	"github.com/pedrodev2026/pystatic/internal/lexer"
	"github.com/pedrodev2026/pystatic/internal/parser"
)

type PackageConfig struct {
	Name    string   `json:"name"`
	Version string   `json:"version"`
	Entry   string   `json:"entry"`
	Modules []string `json:"modules"`
}

const Version = "0.1"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle version flag early
	if command == "--version" || command == "version" || command == "-v" {
		fmt.Println("PyStatic", Version)
		return
	}

	switch command {
	case "run":
		runCmd := flag.NewFlagSet("run", flag.ExitOnError)
		showTokens := runCmd.Bool("tokens", false, "Print lexed tokens and exit")
		showAST := runCmd.Bool("ast", false, "Print AST and exit")

		if len(os.Args) < 3 {
			fmt.Println("Error: No file specified for 'run' command.")
			fmt.Println("Usage: pystatic run <file.pstic> [--tokens] [--ast]")
			os.Exit(1)
		}

		filePath := os.Args[2]
		
		// Parse flags from the rest of the arguments (e.g. command-line flags after file path)
		// We support both `pystatic run file.pstic --tokens` and `pystatic run --tokens file.pstic`
		// To handle this robustly, we check if the first arg after 'run' starts with '-'
		var parsedFlags []string
		if len(os.Args) >= 3 {
			for i := 2; i < len(os.Args); i++ {
				if os.Args[i][0] == '-' {
					parsedFlags = append(parsedFlags, os.Args[i])
				} else {
					filePath = os.Args[i]
				}
			}
		}

		runCmd.Parse(parsedFlags)

		if filePath == "" || filePath[0] == '-' {
			fmt.Println("Error: No file specified for 'run' command.")
			fmt.Println("Usage: pystatic run <file.pstic> [--tokens] [--ast]")
			os.Exit(1)
		}

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", filePath, err)
			os.Exit(1)
		}

		if *showTokens {
			l := lexer.New(string(content))
			for {
				tok := l.NextToken()
				fmt.Printf("[%d:%d] Type: %-10s Literal: %q\n", tok.Line, tok.Column, tok.Type, tok.Literal)
				if tok.Type == lexer.EOF {
					break
				}
			}
			return
		}

		if *showAST {
			l := lexer.New(string(content))
			p := parser.New(l)
			prog := p.ParseProgram()
			if len(p.Errors()) > 0 {
				fmt.Println("Parser errors:")
				for _, err := range p.Errors() {
					fmt.Println("  ", err)
				}
				os.Exit(1)
			}
			fmt.Print(prog.String())
			return
		}

		// Transpile and run
		l := lexer.New(string(content))
		p := parser.New(l)
		prog := p.ParseProgram()
		if len(p.Errors()) > 0 {
			fmt.Println("Parser errors:")
			for _, err := range p.Errors() {
				fmt.Println("  ", err)
			}
			os.Exit(1)
		}

		gen := codegen.New(prog)
		goCode := gen.Generate()

		// Create a temporary folder inside workspace
		tmpDir := "./tmp_run"
		err = os.MkdirAll(tmpDir, 0755)
		if err != nil {
			fmt.Printf("Error creating temp dir: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(tmpDir)

		tmpFile := tmpDir + "/main.go"
		err = ioutil.WriteFile(tmpFile, []byte(goCode), 0644)
		if err != nil {
			fmt.Printf("Error writing transpiled file: %v\n", err)
			os.Exit(1)
		}

		// Run go run
		cmd := exec.Command("go", "run", tmpFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		err = cmd.Run()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			os.Exit(1)
		}
		return

	case "build":
		buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
		outputFile := buildCmd.String("o", "", "Output binary path")
		buildCmd.Parse(os.Args[2:])

		entryFile := ""
		projectName := "app"

		if buildCmd.NArg() > 0 {
			entryFile = buildCmd.Arg(0)
		}

		if entryFile == "" {
			content, err := ioutil.ReadFile("pstic-pkg.json")
			if err != nil {
				fmt.Printf("Error: No input file specified and pstic-pkg.json is missing: %v\n", err)
				os.Exit(1)
			}

			var config PackageConfig
			err = json.Unmarshal(content, &config)
			if err != nil {
				fmt.Printf("Error parsing pstic-pkg.json: %v\n", err)
				os.Exit(1)
			}
			entryFile = config.Entry
			projectName = config.Name
		} else {
			projectName = entryFile
			if idx := strings.LastIndex(projectName, "/"); idx != -1 {
				projectName = projectName[idx+1:]
			}
			if idx := strings.Index(projectName, "."); idx != -1 {
				projectName = projectName[:idx]
			}
		}

		if entryFile == "" {
			entryFile = "main.pstic"
		}

		entryContent, err := ioutil.ReadFile(entryFile)
		if err != nil {
			fmt.Printf("Error reading entry file %s: %v\n", entryFile, err)
			os.Exit(1)
		}

		l := lexer.New(string(entryContent))
		p := parser.New(l)
		prog := p.ParseProgram()
		if len(p.Errors()) > 0 {
			fmt.Println("Parser errors:")
			for _, err := range p.Errors() {
				fmt.Println("  ", err)
			}
			os.Exit(1)
		}

		gen := codegen.New(prog)
		goCode := gen.Generate()

		tmpDir := "./tmp_build"
		err = os.MkdirAll(tmpDir, 0755)
		if err != nil {
			fmt.Printf("Error creating temp build dir: %v\n", err)
			os.Exit(1)
		}
		defer os.RemoveAll(tmpDir)

		tmpFile := tmpDir + "/main.go"
		err = ioutil.WriteFile(tmpFile, []byte(goCode), 0644)
		if err != nil {
			fmt.Printf("Error writing transpiled file: %v\n", err)
			os.Exit(1)
		}

		outPath := *outputFile
		if outPath == "" {
			err = os.MkdirAll("build", 0755)
			if err != nil {
				fmt.Printf("Error creating build directory: %v\n", err)
				os.Exit(1)
			}
			outPath = "build/" + projectName
		} else {
			if idx := strings.LastIndex(outPath, "/"); idx != -1 {
				err = os.MkdirAll(outPath[:idx], 0755)
				if err != nil {
					fmt.Printf("Error creating custom build directory: %v\n", err)
					os.Exit(1)
				}
			}
		}

		fmt.Printf("Building %s...\n", outPath)
		cmd := exec.Command("go", "build", "-o", outPath, tmpFile)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			fmt.Printf("Build failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully built %s\n", outPath)
		return

	case "new":
		if len(os.Args) < 3 {
			fmt.Println("Error: No project name specified.")
			fmt.Println("Usage: pystatic new <project-name>")
			os.Exit(1)
		}
		projectName := os.Args[2]

		err := os.MkdirAll(projectName, 0755)
		if err != nil {
			fmt.Printf("Error creating project directory: %v\n", err)
			os.Exit(1)
		}

		pkgJSON := fmt.Sprintf(`{
  "name": "%s",
  "version": "0.1.0",
  "entry": "main.pstic",
  "modules": []
}
`, projectName)

		err = ioutil.WriteFile(projectName+"/pstic-pkg.json", []byte(pkgJSON), 0644)
		if err != nil {
			fmt.Printf("Error creating pstic-pkg.json: %v\n", err)
			os.Exit(1)
		}

		mainPstic := `def main():
    print("Hello from PyStatic!")
`
		err = ioutil.WriteFile(projectName+"/main.pstic", []byte(mainPstic), 0644)
		if err != nil {
			fmt.Printf("Error creating main.pstic: %v\n", err)
			os.Exit(1)
		}

		gitignore := `/build/
/tmp_run/
`
		err = ioutil.WriteFile(projectName+"/.gitignore", []byte(gitignore), 0644)
		if err != nil {
			fmt.Printf("Error creating .gitignore: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Created new PyStatic project: %s\n", projectName)
		return

	case "add":
		if len(os.Args) < 4 || os.Args[2] != "mod" {
			fmt.Println("Error: Invalid syntax.")
			fmt.Println("Usage: pystatic add mod <module-name>")
			os.Exit(1)
		}
		modName := os.Args[3]

		content, err := ioutil.ReadFile("pstic-pkg.json")
		if err != nil {
			fmt.Printf("Error reading pstic-pkg.json: %v. Are you in a PyStatic project directory?\n", err)
			os.Exit(1)
		}

		var config PackageConfig
		err = json.Unmarshal(content, &config)
		if err != nil {
			fmt.Printf("Error parsing pstic-pkg.json: %v\n", err)
			os.Exit(1)
		}

		alreadyExists := false
		for _, m := range config.Modules {
			if m == modName {
				alreadyExists = true
				break
			}
		}

		if alreadyExists {
			fmt.Printf("Module %s is already in dependencies.\n", modName)
			return
		}

		config.Modules = append(config.Modules, modName)

		updatedContent, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			fmt.Printf("Error encoding updated config: %v\n", err)
			os.Exit(1)
		}

		err = ioutil.WriteFile("pstic-pkg.json", updatedContent, 0644)
		if err != nil {
			fmt.Printf("Error writing to pstic-pkg.json: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Added module %s to pstic-pkg.json\n", modName)
		return

	case "test":
		// Discover all *.test.pstic files in current directory
		pattern := "*.test.pstic"
		if len(os.Args) > 2 {
			pattern = os.Args[2]
		}
		matches, err := filepath.Glob(pattern)
		if err != nil || len(matches) == 0 {
			// Also try tests/ directory
			matches, err = filepath.Glob("tests/" + pattern)
		}
		if err != nil {
			fmt.Printf("Error finding test files: %v\n", err)
			os.Exit(1)
		}
		if len(matches) == 0 {
			fmt.Println("No test files found (*.test.pstic)")
			os.Exit(0)
		}

		passed := 0
		failed := 0
		for _, testFile := range matches {
			fmt.Printf("=== RUN   %s\n", testFile)
			content, err := ioutil.ReadFile(testFile)
			if err != nil {
				fmt.Printf("--- FAIL  %s (cannot read file: %v)\n", testFile, err)
				failed++
				continue
			}
			l := lexer.New(string(content))
			p := parser.New(l)
			prog := p.ParseProgram()
			if len(p.Errors()) > 0 {
				fmt.Printf("--- FAIL  %s (parse errors)\n", testFile)
				for _, e := range p.Errors() {
					fmt.Println("  ", e)
				}
				failed++
				continue
			}
			gen := codegen.New(prog)
			goCode := gen.Generate()

			tmpDir, err := os.MkdirTemp("", "pstic-test-*")
			if err != nil {
				fmt.Printf("--- FAIL  %s (temp dir: %v)\n", testFile, err)
				failed++
				continue
			}
			tmpFile := filepath.Join(tmpDir, "main.go")
			if err := ioutil.WriteFile(tmpFile, []byte(goCode), 0644); err != nil {
				os.RemoveAll(tmpDir)
				fmt.Printf("--- FAIL  %s (write: %v)\n", testFile, err)
				failed++
				continue
			}
			cmd := exec.Command("go", "run", tmpFile)
			out, runErr := cmd.CombinedOutput()
			os.RemoveAll(tmpDir)
			if runErr != nil {
				fmt.Printf("--- FAIL  %s\n", testFile)
				fmt.Printf("%s", out)
				failed++
			} else {
				fmt.Printf("--- PASS  %s\n", testFile)
				if len(out) > 0 {
					fmt.Printf("%s", out)
				}
				passed++
			}
		}
		fmt.Printf("\nResults: %d passed, %d failed\n", passed, failed)
		if failed > 0 {
			os.Exit(1)
		}
		return

	default:
		fmt.Printf("Error: Unknown command %q\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("PyStatic compiler tool")
	fmt.Println()
	fmt.Printf("PyStatic %s — Python-style language that compiles to Go\n", Version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  pystatic <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  run <file.pstic> [--tokens] [--ast]   Transpile and run a PyStatic file")
	fmt.Println("  build [file.pstic] [-o output]        Build a static binary")
	fmt.Println("  test [pattern]                        Run *.test.pstic test files")
	fmt.Println("  new <project-name>                    Create a new project scaffold")
	fmt.Println("  add mod <module-name>                 Add a module dependency")
	fmt.Println("  version                               Print the PyStatic version")
}
