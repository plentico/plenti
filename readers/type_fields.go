package readers

import (
	"encoding/json"
	"fmt"
)

// TypeFields maps to field key/values for content types.
type TypeFields struct {
	Fields map[string]string `json:",string"`
}

// UnmarshalJSON method filters out non-string values from content source.
func (t *TypeFields) UnmarshalJSON(data []byte) error {
	var unknownValues map[string]interface{}
	if err := json.Unmarshal(data, &unknownValues); err != nil {
		return err
	}
	t.Fields = map[string]string{}
	for field, unknownValue := range unknownValues {
		_, isString := unknownValue.(string)
		if isString {
			t.Fields[field] = fmt.Sprintf("%v", unknownValue)
			fmt.Printf("Its a string: %s", field)
		}
	}
	fmt.Printf("I HAVE RUN!!!!")
	return nil
}

// GetTypeFields reads the key/values for an individual content type JSON file.
func GetTypeFields(typeFileContents []byte) TypeFields {

	var typeFields TypeFields
	//typeFields.UnmarshalJSON(typeFileContents)

	var unknownValues map[string]interface{}
	err := json.Unmarshal(typeFileContents, &unknownValues)
	//err := json.Unmarshal(typeFileContents, &typeFields.Fields)
	if err != nil {
		fmt.Printf("Unable to read content type source file: %s\n", err)
	}
	typeFields.Fields = map[string]string{}
	for field, unknownValue := range unknownValues {
		_, isString := unknownValue.(string)
		if isString {
			typeFields.Fields[field] = fmt.Sprintf("%v", unknownValue)
			fmt.Printf("Its a string: %s", field)
		}
	}

	return typeFields
}
