package readers

import (
	"encoding/json"
	"fmt"
)

// ContentType maps to field key/values for content types.
type ContentType struct {
	Fields map[string]string `json:"-"`
}

// GetTypeFields reads the key/values for an individual content type JSON file.
func GetTypeFields(typeFileContents []byte) ContentType {

	var contentType ContentType
	contentType.Fields = map[string]string{}

	// Use empty interface to store JSON field values of unknown primitive type.
	var unknownValues map[string]interface{}
	// Put JSON key/values into the map of unknown primitives.
	err := json.Unmarshal(typeFileContents, &unknownValues)
	if err != nil {
		fmt.Printf("Unable to read content type source file: %s\n", err)
	}
	for field, unknownValue := range unknownValues {
		// isString will be set to true if the value is a string (vs array, obj, etc).
		_, isString := unknownValue.(string)
		if isString {
			// Convert the empty interface into a string that can be stored in ContentType struct.
			contentType.Fields[field] = fmt.Sprintf("%v", unknownValue)
		}
	}

	return contentType
}
