package build

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"time"
)

// NpmDefaults creates the node_modules folder with core defaults if it doesn't already exist.
func NpmDefaults(defaultsNodeModulesFS embed.FS) error {

	defer Benchmark(time.Now(), "Setting up core NPM packages")

	Log("\nChecking if 'node_modules' directory exists.")

	destPath := "node_modules"

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		nodeModules, err := fs.Sub(defaultsNodeModulesFS, "defaults")
		if err != nil {
			return fmt.Errorf("Unable to get node_modules defaults: %w\n", err)
		}
		fs.WalkDir(nodeModules, ".", func(path string, d fs.DirEntry, err error) error {

			if err != nil {
				return fmt.Errorf("Unable to get stat path %s: %w\n", path, err)
			}

			if d.IsDir() {
				// Create the directories needed for the current file
				if err := os.MkdirAll(path, os.ModePerm); err != nil {
					return fmt.Errorf("Unable to create path(s) %s: %v\n", path, err)
				}
				return nil
			}
			content, err := nodeModules.Open(path)
			if err != nil {
				return fmt.Errorf("Unable to op path %s: %v\n", path, err)
			}
			contentBytes, err := ioutil.ReadAll(content)
			if err != nil {
				return fmt.Errorf("Unable to read node_modules file: %w\n", err)

			}
			// Create the current default file
			if err := ioutil.WriteFile(path, contentBytes, 0755); err != nil {
				return fmt.Errorf("Unable to write node_modules file: %w\n", err)
			}
			return nil
		})
	}
	return nil
}
