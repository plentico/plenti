package build

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plenti/common"
	"plenti/generated"
	"time"
)

// EjectTemp temporarily writes ejectable core files to project filesystem.
func EjectTemp(tempBuildDir string) ([]string, string, error) {

	defer Benchmark(time.Now(), "Creating non-ejected core files for build")

	Log("\nEjecting core files to be used in build:")

	ejectedPath := tempBuildDir + "ejected"

	tempFiles := []string{}

	// Loop through generated ejected file defaults.
	for file, content := range generated.Ejected {
		filePath := ejectedPath + file
		// Create the directories needed for the current file
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return nil, "", err
		}
		if _, ejectedFileExistsErr := os.Stat(filePath); os.IsNotExist(ejectedFileExistsErr) {
			Log("Temp writing '" + file + "' file.")
			// Create the current default file
			err := ioutil.WriteFile(filePath, content, os.ModePerm)
			if err != nil {
				return nil, "", fmt.Errorf("Unable to write ejected core file: %w%s", err, common.Caller())
			}
			tempFiles = append(tempFiles, filePath)

		} else {
			Log("File '" + file + "' has been ejected already, skipping temp write.")
		}
	}

	return tempFiles, ejectedPath, nil

}
