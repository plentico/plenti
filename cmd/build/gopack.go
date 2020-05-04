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

	fmt.Println("\nRunning gopack to build esm support for:")
	// Find all the "dependencies" specified in package.json.
	for module, version := range readers.GetNpmConfig().Dependencies {
		fmt.Printf("- %s, version %s\n", module, version)
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
			//fmt.Printf("The file to convert to esm is: %s\n", convertPath)

			// Find any import statement in the file (including multiline imports).
			// () = brackets for grouping
			// \s = space
			// .* = any character
			// | = or statement
			// \n = newline
			// {0,} = repeat any number of times
			// \{ = just a closing curly bracket (escaped)
			reImport := regexp.MustCompile(`import(\s)(.*from(.*);|((.*\n){0,})\}(\s)from(.*);)`)
			// Get all the import statements.
			importStatements := reImport.FindAll(contentBytes, -1)
			for _, importStatement := range importStatements {
				//fmt.Printf("the import statement is: %s\n", importStatement)
				// Find the path specifically (part between single or double quotes).
				rePath := regexp.MustCompile(`(?:'|").*(?:'|")`)
				// Get path from the full import statement.
				importPath := rePath.Find(importStatement)
				// Convert path to a string.
				importPathStr := string(importPath)
				// Remove single or double quotes around path.
				importPathStr = strings.Trim(importPathStr, `'"`)
				//fmt.Printf("the path is: %s\n", importPathStr)
				// Make the path relative to the file that is specifying it as an import.
				fullImportPath := filepath.Dir(convertPath) + "/" + importPathStr
				//fmt.Printf("Full import path is: %s\n", fullImportPath)
				var foundImportPath string
				if filepath.Ext(fullImportPath) == ".svelte" {
					fullImportPath = strings.Replace(fullImportPath, ".svelte", ".js", 1)
					foundImportPath = fullImportPath
				}
				// If the import points to a path that exists and it is a .js file (imports must reference the file specifically) then we don't need to convert anything.
				if _, importExistsErr := os.Stat(fullImportPath); !os.IsNotExist(importExistsErr) && filepath.Ext(fullImportPath) == ".js" {
					//fmt.Printf("Skipping converting import in %s because import is valid: %s\n", convertPath, importStatement)
				} else if importPathStr[:1] == "." {
					// If the import starts with a dot (.) or double dot (..) look for the file it's trying to import from this relative path.
					findRelativeImportErr := filepath.Walk(fullImportPath, func(relativeImportPath string, relativeImportFileInfo os.FileInfo, err error) error {
						// Only use .js files in imports (folders aren't specific enough).
						if filepath.Ext(relativeImportPath) == ".js" {
							foundImportPath = relativeImportPath
						}
						return nil
					})
					if findRelativeImportErr != nil {
						fmt.Printf("Could not find related .mjs file: %s", findRelativeImportErr)
					}
				} else {
					// A named import is being used, look for this in "web_modules/" dir.
					findNamedImportErr := filepath.Walk(buildPath+"/spa/web_modules/"+importPathStr, func(namedImportPath string, namedImportFileInfo os.FileInfo, err error) error {
						if filepath.Ext(namedImportPath) == ".js" {
							foundImportPath = namedImportPath
							//fmt.Printf("The found import path to use is: %s\n", foundImportPath)
						}
						return nil
					})
					if findNamedImportErr != nil {
						fmt.Printf("Could not find related .js file from named import: %s", findNamedImportErr)
					}
				}
				if foundImportPath != "" {
					// Remove "public" build dir from path.
					replacePath := strings.Replace(foundImportPath, buildPath, "", 1)
					// Wrap path in quotes.
					replacePath = "'" + replacePath + "'"
					// Convert string path to bytes.
					replacePathBytes := []byte(replacePath)
					// Find the specific import statement we're replacing.
					reFoundImport := regexp.MustCompile(string(importStatement))
					// Actually replace the path to the dependency in the source content.
					contentBytes = reFoundImport.ReplaceAll(contentBytes, rePath.ReplaceAll(importStatement, rePath.ReplaceAll(importPath, replacePathBytes)))
				}
			}
			// Overwrite the old file with the new content that contains the updated import path.
			overwriteImportErr := ioutil.WriteFile(convertPath, contentBytes, 0644)
			if overwriteImportErr != nil {
				fmt.Printf("Could not overwite %s with new import: %s", convertPath, overwriteImportErr)
			}
		}
		return nil
	})
	if convertErr != nil {
		fmt.Printf("Could not convert file to support esm: %s", convertErr)
	}

	elapsed := time.Since(start)
	fmt.Printf("Gopack took %s\n", elapsed)

}
