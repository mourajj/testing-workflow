package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		filePath := scanner.Text()
		if strings.HasSuffix(filePath, ".json") {
			fmt.Println("Validating:", filePath)
			err := validateJSONSchema(filePath)
			if err != nil {
				fmt.Println("Validation failed for", filePath, ":", err)
				os.Exit(1)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	fmt.Println("All schemas are valid")
}

func validateJSONSchema(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var schema map[string]interface{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&schema); err != nil {
		return fmt.Errorf("invalid JSON schema: %w", err)
	}
	return nil
}
