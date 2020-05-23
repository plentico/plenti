package readers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// TypeFields maps to field key/values for content types.
type TypeFields struct {
	Fields map[string]string `json:"fields"`
}

// GetTypeFields reads the key/values for an individual content type JSON file.
func GetTypeFields(typeFile string) TypeFields {

	var typeFields TypeFields

	// Read site config file from the project
	typeFileContents, _ := ioutil.ReadFile(typeFile)
	err := json.Unmarshal(typeFileContents, &typeFields)
	if err != nil {
		fmt.Printf("Unable to read content type source file: %s\n", err)
	}

	return typeFields
}
