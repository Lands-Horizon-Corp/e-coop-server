package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run update_controller_calls.go <file>")
		os.Exit(1)
	}

	file := os.Args[1]
	err := processFile(file)
	if err != nil {
		fmt.Printf("Error processing file: %v\n", err)
		os.Exit(1)
	}
}

func processFile(path string) error {
	fmt.Printf("Processing: %s\n", path)

	// Read file
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	changed := false

	// Regex to match method calls like "c.MethodName()"
	methodCallRegex := regexp.MustCompile(`^(\s*c\.)([A-Z][a-zA-Z0-9]*)(\(\))`)

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this line contains a method call with uppercase first letter
		if matches := methodCallRegex.FindStringSubmatch(line); matches != nil {
			methodName := matches[2]
			
			// Skip certain methods that should remain capitalized
			skipMethods := []string{
				"Start", // Keep the Start method capitalized
			}
			
			shouldSkip := false
			for _, skip := range skipMethods {
				if methodName == skip {
					shouldSkip = true
					break
				}
			}
			
			if !shouldSkip {
				newMethodName := strings.ToLower(string(methodName[0])) + methodName[1:]
				if newMethodName != methodName {
					line = matches[1] + newMethodName + matches[3]
					fmt.Printf("  Updating method call: %s -> %s\n", methodName, newMethodName)
					changed = true
				}
			}
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Write back if changed
	if changed {
		output, err := os.Create(path)
		if err != nil {
			return err
		}
		defer output.Close()

		writer := bufio.NewWriter(output)
		for _, line := range lines {
			fmt.Fprintln(writer, line)
		}
		writer.Flush()

		fmt.Printf("  Updated: %s\n", path)
	}

	return nil
}