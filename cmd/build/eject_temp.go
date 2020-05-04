package build

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plenti/generated"
	"time"
)

// EjectTemp temporarily writes ejectable core files to project filesystem.
func EjectTemp() []string {

	start := time.Now()

	ejectedPath := "layout/ejected"

	tempFiles := []string{}

	// Loop through generated ejected file defaults.
	for file, content := range generated.Ejected {
		filePath := ejectedPath + file
		// Create the directories needed for the current file
		os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if _, ejectedFileExistsErr := os.Stat(filePath); os.IsNotExist(ejectedFileExistsErr) {
			fmt.Printf("Temp writing '%s' file.\n", file)
			// Create the current default file
			writeCoreFileErr := ioutil.WriteFile(filePath, content, os.ModePerm)
			if writeCoreFileErr != nil {
				fmt.Printf("Unable to write file: %v\n", writeCoreFileErr)
			} else {
				tempFiles = append(tempFiles, filePath)
			}
		} else {
			fmt.Printf("File '%s' has been ejected already, skipping temp write.\n", file)
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("Creating non-ejected core files for build took %s\n", elapsed)

	return tempFiles

}
