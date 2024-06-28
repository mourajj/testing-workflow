package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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
	jsonFile, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening JSON file:", err)
		return err
	}
	defer jsonFile.Close()

	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return err
	}

	// Unmarshal the schema and perform a MarshalIndent to format the JSON schema
	var result map[string]interface{}
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)
		return err
	}

	val, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return err
	}

	// Convert the slice of bytes into a slice of strings, based on each line of the formatted schema
	myString := string(val[:])
	formattedString := strings.Split(myString, "\n")

	dataClassifiedCount := 0
	count := 0
	countObjects := 0
	countFalseProperties := 0

	for line, x := range formattedString {

		//Edge-case handling for types defined as arrays, e.g:
		//"type":[
		//       "null",
		//       "string"
		//    ]
		if strings.Contains(x, fmt.Sprintf("%q", "type")) && strings.Contains(x, "[") {
			boundary := line
			for !strings.Contains(formattedString[boundary], "]") {
				if strings.Contains(formattedString[boundary], "string") || strings.Contains(formattedString[boundary], "boolean") || strings.Contains(formattedString[boundary], "number") {
					CheckStructure(boundary, formattedString, &dataClassifiedCount, &count)
					break
				}
				boundary++
			}
			//Edge-case handling for normal types and for attributes also called as 'type', e.g:
			//"type": "string"
			//"type": {}
			//the Sprintf format checks for strings contaning "type" (including quotation marks to prevent any description or any non-important text)
		} else if strings.Contains(x, fmt.Sprintf("%q", "type")) && !strings.Contains(x, "object") && !strings.Contains(x, "array") && !strings.Contains(x, "{") {
			CheckStructure(line, formattedString, &dataClassifiedCount, &count)
		} else if strings.Contains(x, fmt.Sprintf("%q", "object")) {
			countObjects++
		} else if strings.Contains(x, fmt.Sprintf("%q", "additionalProperties")) && strings.Contains(x, "false") {
			countFalseProperties++
		}
	}

	if count == 0 {
		fmt.Println("Please check if there is a 'type' attribute defined in the json objects")
	} else {
		fmt.Println("\033[0m")
		fmt.Println(count, "structures were checked")
		fmt.Println("There are\033[31m", dataClassifiedCount, "\033[0mstructures with no DataClassification field")
	}
	if countObjects != countFalseProperties {
		fmt.Printf("There are %d structures of type 'object', and %d of them has additionalProperties field defined as false", countObjects, countFalseProperties)
	}
	return nil
}

func CheckStructure(bottomPointer int, formattedString []string, dataClassifiedCount, count *int) {

	topPointer := bottomPointer - 1
	dataClassification := false
	for !strings.Contains(formattedString[bottomPointer], "}") || !strings.Contains(formattedString[topPointer], "{") && !dataClassification {

		if strings.Contains(formattedString[bottomPointer], "dataClassification") || strings.Contains(formattedString[topPointer], "dataClassification") {
			dataClassification = true
		}

		if !strings.Contains(formattedString[bottomPointer], "}") && !strings.Contains(formattedString[topPointer], "{") {
			topPointer--
			bottomPointer++
		} else if !strings.Contains(formattedString[bottomPointer], "}") {
			bottomPointer++
		} else {
			topPointer--
		}

		//Edge-case handling for regex-expressions that contains '{' or '}', for example:
		// "pattern": "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
		if strings.Contains(formattedString[bottomPointer], "pattern") {
			bottomPointer += 1
		}
		if strings.Contains(formattedString[topPointer], "pattern") {
			topPointer -= 1
		}
	}
	*count++
	if !dataClassification {
		fmt.Println("\033[31m", strings.TrimSpace(formattedString[topPointer]), strings.TrimSpace(formattedString[bottomPointer]))
		*dataClassifiedCount++
	} else {
		fmt.Println("\033[0m", strings.TrimSpace(formattedString[topPointer]), strings.TrimSpace(formattedString[bottomPointer]))
	}

}
