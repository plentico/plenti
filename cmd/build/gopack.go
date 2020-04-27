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

	if _, err := os.Stat(gopackDir); os.IsNotExist(err) {
		// If directory doesn't already exist, run gopack.
		fmt.Println("\nRunning gopack to build dependencies for esm support")
		// Create the web_modules directory in client app.
		//os.MkdirAll(gopackDir, os.ModePerm)
		for module, version := range readers.GetNpmConfig().Dependencies {
			fmt.Printf("npm config module: %s\n", module)
			fmt.Printf("npm config version: %s\n\n", version)
			nodeModuleErr := filepath.Walk("node_modules/"+module, func(modulePath string, moduleFileInfo os.FileInfo, err error) error {
				// Make sub directories.
				/*
					if moduleFileInfo.IsDir() {
						os.MkdirAll(gopackDir+"/"+moduleFileInfo.Name(), os.ModePerm)
					}
				*/
				// Only get ESM supported files.
				if !moduleFileInfo.IsDir() && filepath.Ext(modulePath) == ".mjs" {
					from, err := os.Open(modulePath)
					if err != nil {
						fmt.Printf("Could not open source .mjs file for copying: %s\n", err)
					}
					defer from.Close()

					modulePath = gopackDir + strings.Replace(modulePath, "node_modules", "", 1)
					fmt.Printf("modulePath is: %s\n", modulePath)
					//destPath := gopackDir + "/" + modulePath
					//to, err := os.Create(gopackDir + "/" + moduleFileInfo.Name())
					os.MkdirAll(filepath.Dir(modulePath), os.ModePerm)
					to, err := os.Create(modulePath)
					//to, err := os.Create(destPath)
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
