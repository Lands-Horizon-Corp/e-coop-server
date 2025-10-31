package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func renameMain() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run function_renamer.go <directory>")
		os.Exit(1)
	}

	dir := os.Args[1]
	err := filepath.Walk(dir, processFile)
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		os.Exit(1)
	}
}

func processFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// Only process .go files
	if !strings.HasSuffix(path, ".go") {
		return nil
	}

	// Skip test files for now
	if strings.HasSuffix(path, "_test.go") {
		return nil
	}

	fmt.Printf("Processing: %s\n", path)

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		fmt.Printf("Error parsing %s: %v\n", path, err)
		return nil
	}

	// Track if we made any changes
	changed := false

	// Walk through the AST
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			// Check if function name starts with uppercase
			if x.Name != nil && x.Name.IsExported() && len(x.Name.Name) > 0 {
				oldName := x.Name.Name
				newName := strings.ToLower(string(oldName[0])) + oldName[1:]

				// Skip certain function names that should remain capitalized
				skipFunctions := []string{
					"NewController", "NewProvider", "NewValidator",
					"Newmodelcore", "Newmodellogs", "NewSeeder",
					"NewTransactionService", "NewUserToken",
					"NewUserOrganizationToken", "NewEvent",
					"GetExchangeRate", "GetCloudflareHeaders",
					"NewTypeScriptGenerator", "Start",
				}

				shouldSkip := false
				for _, skip := range skipFunctions {
					if oldName == skip {
						shouldSkip = true
						break
					}
				}

				if !shouldSkip && oldName != newName {
					fmt.Printf("  Renaming function: %s -> %s\n", oldName, newName)
					x.Name.Name = newName
					changed = true
				}
			}
		}
		return true
	})

	// If we made changes, write the file back
	if changed {
		file, err := os.Create(path)
		if err != nil {
			fmt.Printf("Error creating file %s: %v\n", path, err)
			return nil
		}
		defer file.Close()

		err = format.Node(file, fset, node)
		if err != nil {
			fmt.Printf("Error formatting file %s: %v\n", path, err)
			return nil
		}

		fmt.Printf("  Updated: %s\n", path)
	}

	return nil
}

func isUppercase(r rune) bool {
	return unicode.IsUpper(r)
}

func init() {
	renameMain()
}
