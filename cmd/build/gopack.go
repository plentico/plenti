package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"plenti/readers"
	"strings"
	"time"
)

// Gopack ensures ESM support for NPM dependencies.
func Gopack(buildPath string) {

	start := time.Now()

	gopackDir := buildPath + "/spa/web_modules"

	// If directory doesn't already exist, run gopack.
	if _, err := os.Stat(gopackDir); os.IsNotExist(err) {
		fmt.Println("\nRunning gopack to build dependencies for esm support")
		// Find all the "dependencies" specified in package.json.
		for module, version := range readers.GetNpmConfig().Dependencies {
			fmt.Printf("npm module: %s, version %s\n", module, version)
			// Walk through all sub directories of each dependency declared.
			nodeModuleErr := filepath.Walk("node_modules/"+module, func(modulePath string, moduleFileInfo os.FileInfo, err error) error {
				// Only get ESM supported files.
				if !moduleFileInfo.IsDir() && filepath.Ext(modulePath) == ".mjs" {
					from, err := os.Open(modulePath)
					if err != nil {
						fmt.Printf("Could not open source .mjs file for copying: %s\n", err)
					}
					defer from.Close()

					// Remove "node_modules" from path and add "web_modules".
					modulePath = gopackDir + strings.Replace(modulePath, "node_modules", "", 1)
					// Create any subdirectories need to write file to "web_modules" destination.
					os.MkdirAll(filepath.Dir(modulePath), os.ModePerm)

					to, err := os.Create(modulePath)
					if err != nil {
						fmt.Printf("Could not create destination .mjs file for copying: %s\n", err)
					}
					defer to.Close()

					_, fileCopyErr := io.Copy(to, from)
					if err != nil {
						fmt.Printf("Could not copy .mjs from source to destination: %s\n", fileCopyErr)
					}
				}
				return nil
			})
			if nodeModuleErr != nil {
				fmt.Printf("Could not get node module: %s", nodeModuleErr)
			}
		}
	} else {
		fmt.Printf("\nThe %s/web_modules directory already exists, skipping gopack\n", buildPath)
	}

	elapsed := time.Since(start)
	fmt.Printf("Gopack took %s\n", elapsed)

}
