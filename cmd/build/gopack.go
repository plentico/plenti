package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"plenti/readers"
	"regexp"
	"strings"
	"time"
)

// Gopack ensures ESM support for NPM dependencies.
func Gopack(buildPath string) {

	start := time.Now()

	gopackDir := buildPath + "/spa/web_modules"

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
				// Change the .mjs file extension to .js.
				modulePath = strings.TrimSuffix(modulePath, filepath.Ext(modulePath)) + ".js"
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
	convertErr := filepath.Walk(buildPath+"/spa", func(convertPath string, convertFileInfo os.FileInfo, err error) error {
		if !convertFileInfo.IsDir() && filepath.Ext(convertPath) == ".js" {
			contentBytes, err := ioutil.ReadFile(convertPath)
			if err != nil {
				fmt.Printf("Could not read file to convert to esm: %s\n", err)
			}
			fmt.Printf("The file to convert to esm is: %s\n", convertPath)
			reImport := regexp.MustCompile("import(.*)from(.*);")
			importStatements := reImport.FindAll(contentBytes, -1)
			for _, importStatement := range importStatements {
				fmt.Printf("the import statement is: %s\n", importStatement)
				rePath := regexp.MustCompile(`(?:'|").*(?:'|")`)
				importPath := rePath.Find(importStatement)
				//fmt.Printf("the path is: %s\n", importPath)
				importPathStr := string(importPath)
				// Remove single or double quotes around path.
				importPathStr = strings.Trim(importPathStr, `'"`)
				fmt.Printf("the path is: %s\n", importPathStr)
				fullImportPath := filepath.Dir(convertPath) + "/" + importPathStr
				fmt.Printf("Full import path is: %s\n", fullImportPath)
				if _, importExistsErr := os.Stat(fullImportPath); !os.IsNotExist(importExistsErr) && filepath.Ext(fullImportPath) == ".js" {
					fmt.Printf("Skipping converting import in %s because import is valid: %s\n", convertPath, importStatement)
				}
				//importStatement.ReplaceAll()
			}
			//contentBytes = re.ReplaceAll(contentBytes, []byte("T"))
			//contentStr := string(contentBytes)
			//if strings.Contains(contentStr, "import") {}
		}
		return nil
	})
	if convertErr != nil {
		fmt.Printf("Could not convert file to support esm: %s", convertErr)
	}

	elapsed := time.Since(start)
	fmt.Printf("Gopack took %s\n", elapsed)

}
