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

	defer Benchmark(time.Now(), "Creating non-ejected core files for build")

	Log("\nEjecting core files to be used in build:")

	ejectedPath := "ejected"

	tempFiles := []string{}

	// Loop through generated ejected file defaults.
	for file, content := range generated.Ejected {
		filePath := ejectedPath + file
		// Create the directories needed for the current file
		os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
		if _, ejectedFileExistsErr := os.Stat(filePath); os.IsNotExist(ejectedFileExistsErr) {
			Log("Temp writing '" + file + "' file.")
			// Create the current default file
			writeCoreFileErr := ioutil.WriteFile(filePath, content, os.ModePerm)
			if writeCoreFileErr != nil {
				fmt.Printf("Unable to write file: %v\n", writeCoreFileErr)
			} else {
				tempFiles = append(tempFiles, filePath)
			}
		} else {
			Log("File '" + file + "' has been ejected already, skipping temp write.")
		}
	}

	return tempFiles

}
