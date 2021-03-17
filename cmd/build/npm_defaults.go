package build

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"plenti/common"
	"time"
)

// NpmDefaults creates the node_modules folder with core defaults if it doesn't already exist.
func NpmDefaults(tempBuildDir string, defaultsNodeModulesFS embed.FS) error {

	defer Benchmark(time.Now(), "Setting up core NPM packages")

	Log("\nChecking if 'node_modules' directory exists.")

	destPath := tempBuildDir + "node_modules"

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		nodeModules, err := fs.Sub(defaultsNodeModulesFS, "defaults")
		if err != nil {
			common.CheckErr(fmt.Errorf("Unable to get node_modules defaults: %w", err))
		}
		fs.WalkDir(nodeModules, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				// Create the directories needed for the current file
				if err := os.MkdirAll(path, os.ModePerm); err != nil {
					common.CheckErr(fmt.Errorf("Unable to create path(s) %s: %v", path, err))
				}
				return nil
			}
			content, _ := nodeModules.Open(path)
			contentBytes, err := ioutil.ReadAll(content)
			// Create the current default file
			if err := ioutil.WriteFile(path, contentBytes, 0755); err != nil {
				common.CheckErr(fmt.Errorf("Unable to write node_modules file: %w", err))
			}
			return nil
		})
	}
	return nil
}
