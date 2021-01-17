package build

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plenti/generated"
	"time"
)

// NpmDefaults creates the node_modules folder with core defaults if it doesn't already exist.
func NpmDefaults(tempBuildDir string) error {

	defer Benchmark(time.Now(), "Setting up core NPM packages")

	Log("\nChecking if 'node_modules' directory exists.")

	destPath := tempBuildDir + "node_modules"

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		for file, content := range generated.Defaults_node_modules {
			// Make file relative to where CLI is executed
			file = destPath + "/" + file
			// Create the directories needed for the current file
			if err = os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
				return fmt.Errorf("Unable to MkdirAll in NpmDefaults: %w", err)
			}
			// Create the current default file
			err := ioutil.WriteFile(file, content, os.ModePerm)
			if err != nil {
				return fmt.Errorf("Unable to write npm dependency file: %w", err)
			}
		}
	}
	return nil
}
