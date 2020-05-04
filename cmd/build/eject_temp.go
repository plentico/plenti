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

	removeFiles := []string{}
	if _, ejectedDirExistsErr := os.Stat(ejectedPath); os.IsNotExist(ejectedDirExistsErr) {
		ejectedDirErr := os.MkdirAll(ejectedPath, os.ModePerm)
		if ejectedDirErr != nil {
			fmt.Printf("Could not created 'ejected' directory: %s\n", ejectedDirErr)
		}
	} else {
		fmt.Printf("Ejected directory already exists\n")
	}
	// Loop through generated ejected file defaults.
	for file, content := range generated.Ejected {
		// Create the directories needed for the current file
		os.MkdirAll(ejectedPath+filepath.Dir(file), os.ModePerm)
		filePath := ejectedPath + file
		if _, ejectedFileExistsErr := os.Stat(filePath); os.IsNotExist(ejectedFileExistsErr) {
			fmt.Printf("Temp writing '%s' file.\n", file)
			// Create the current default file
			writeCoreFileErr := ioutil.WriteFile(filePath, content, os.ModePerm)
			if writeCoreFileErr != nil {
				fmt.Printf("Unable to write file: %v\n", writeCoreFileErr)
			} else {
				removeFiles = append(removeFiles, filePath)
			}
		} else {
			fmt.Printf("File '%s' has been ejected already, skipping temp write.\n", file)
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("Creating non-ejected core files for build took %s\n", elapsed)

	return removeFiles

}
