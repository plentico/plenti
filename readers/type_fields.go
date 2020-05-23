package readers

import (
	"encoding/json"
	"fmt"
)

// TypeFields maps to field key/values for content types.
type TypeFields struct {
	Fields map[string]string `json:",string"`
}

// GetTypeFields reads the key/values for an individual content type JSON file.
func GetTypeFields(typeFileContents []byte) TypeFields {

	var typeFields TypeFields
	err := json.Unmarshal(typeFileContents, &typeFields.Fields)
	if err != nil {
		fmt.Printf("Unable to read content type source file: %s\n", err)
	}

	return typeFields
}
