package build

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"plenti/common"
	"plenti/readers"
	"regexp"
	"strings"
	"time"
)

// Gopack ensures ESM support for NPM dependencies.
func Gopack(buildPath string) error {

	defer Benchmark(time.Now(), "Running Gopack")

	gopackDir := buildPath + "/spa/web_modules"

	Log("\nRunning gopack to build esm support for npm dependencies:")

	// Find all the "dependencies" specified in package.json.
	for module, version := range readers.GetNpmConfig().Dependencies {
		Log("- " + module + ", version " + version)
		// Walk through all sub directories of each dependency declared.
		nodeModuleErr := filepath.Walk("node_modules/"+module, func(modulePath string, moduleFileInfo os.FileInfo, err error) error {

			if err != nil {
				return fmt.Errorf("can't stat %s: %w", modulePath, err)
			}
			// Only get ESM supported files.
			if !moduleFileInfo.IsDir() && filepath.Ext(modulePath) == ".mjs" {
				from, err := os.Open(modulePath)
				if err != nil {
					return fmt.Errorf("Could not open source .mjs %s file for copying: %w%s", modulePath, err, common.Caller())
				}
				defer from.Close()

				// Remove "node_modules" from path and add "web_modules".
				modulePath = gopackDir + strings.Replace(modulePath, "node_modules", "", 1)
				// Create any subdirectories need to write file to "web_modules" destination.
				if err = os.MkdirAll(filepath.Dir(modulePath), os.ModePerm); err != nil {
					return fmt.Errorf("Could not create subdirectories %s: %w%s", filepath.Dir(modulePath), err, common.Caller())
				}
				to, err := os.Create(modulePath)
				if err != nil {
					return fmt.Errorf("Could not create destination %s file for copying: %w%s", modulePath, err, common.Caller())
				}
				defer to.Close()

				_, err = io.Copy(to, from)
				if err != nil {
					return fmt.Errorf("Could not copy .mjs  from source to destination: %w%s", err, common.Caller())
				}
			}
			return nil
		})
		if nodeModuleErr != nil {
			return fmt.Errorf("Could not get node module: %w%s", nodeModuleErr, common.Caller())
		}

	}
	convertErr := filepath.Walk(buildPath+"/spa", func(convertPath string, convertFileInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", convertPath, err)
		}
		if !convertFileInfo.IsDir() && (filepath.Ext(convertPath) == ".js" || filepath.Ext(convertPath) == ".mjs") {
			contentBytes, err := ioutil.ReadFile(convertPath)
			if err != nil {
				return fmt.Errorf("Could not read file %s to convert to esm: %w%s", convertPath, err, common.Caller())
			}

			// Match dynamic import statments, e.g. import("") or import('').
			reDynamicImport := regexp.MustCompile(`import\((?:'|").*(?:'|")\)`)
			// Created byte array of all dynamic imports in the current file.
			dynamicImportPaths := reDynamicImport.FindAll(contentBytes, -1)
			for _, dynamicImportPath := range dynamicImportPaths {
				// Inside the dynamic import change any svelte file extensions to reference regular javascript files.
				fixedImportPath := bytes.Replace(dynamicImportPath, []byte(".svelte"), []byte(".js"), 1)
				// Add the updated import back into the file contents for writing later.
				contentBytes = bytes.Replace(contentBytes, dynamicImportPath, fixedImportPath, 1)
			}

			// Find any import statement in the file (including multiline imports).
			// () = brackets for grouping
			// \s = space
			// .* = any character
			// | = or statement
			// \n = newline
			// {0,} = repeat any number of times
			// \{ = just a closing curly bracket (escaped)
			reStaticImport := regexp.MustCompile(`(?m)^import(\s)(.*from(.*);|((.*\n){0,})\}(\s)from(.*);)`)
			reStaticExport := regexp.MustCompile(`export(\s)(.*from(.*);|((.*\n){0,})\}(\s)from(.*);)`)
			// Get all the import statements.
			staticImportStatements := reStaticImport.FindAll(contentBytes, -1)
			// Get all the export statements.
			staticExportStatements := reStaticExport.FindAll(contentBytes, -1)
			// Get all import and export statements.
			allStaticStatements := append(staticImportStatements, staticExportStatements...)
			for _, staticStatement := range allStaticStatements {
				// Find the path specifically (part between single or double quotes).
				rePath := regexp.MustCompile(`(?:'|").*(?:'|")`)
				// Get path from the full import/export statement.
				pathBytes := rePath.Find(staticStatement)
				// Convert path to a string.
				pathStr := string(pathBytes)
				// Remove single or double quotes around path.
				pathStr = strings.Trim(pathStr, `'"`)
				// Make the path relative to the file that is specifying it as an import/export.
				fullPath := filepath.Dir(convertPath) + "/" + pathStr
				// Intialize the path that we are replacing.
				var foundPath string
				// Convert .svelte file extensions to .js so the browser can read them.
				if filepath.Ext(fullPath) == ".svelte" {
					fullPath = strings.Replace(fullPath, ".svelte", ".js", 1)
					foundPath = fullPath
				}
				// If the import/export points to a path that exists and it is a .js file (imports must reference the file specifically) then we don't need to convert anything.
				if _, pathExistsErr := os.Stat(fullPath); !os.IsNotExist(pathExistsErr) && filepath.Ext(fullPath) == ".js" {
					// error?
					Log("Skipping converting import/export in " + convertPath + " because import/export is valid: " + string(staticStatement))
				} else if pathStr[:1] == "." {
					// If the import/export path starts with a dot (.) or double dot (..) look for the file it's trying to import from this relative path.
					findRelativePathErr := filepath.Walk(fullPath, func(relativePath string, relativePathFileInfo os.FileInfo, err error) error {
						if err != nil {
							return fmt.Errorf("can't stat %s: %w", relativePath, err)
						}
						// Only use .js files in imports (folders aren't specific enough).
						if filepath.Ext(relativePath) == ".js" {
							foundPath = relativePath
						}
						return nil
					})
					if findRelativePathErr != nil {
						return fmt.Errorf("Could not find related .mjs file: %w%s", findRelativePathErr, common.Caller())
					}
				} else {
					// A named import/export is being used, look for this in "web_modules/" dir.
					namedPath := buildPath + "/spa/web_modules/" + pathStr
					// Check all files in the current directory first.
					foundPath = findJSFile(namedPath)
					if foundPath == "" {
						// If JS file was not found in the current directory, check nested directories.
						findNamedPathErr := filepath.Walk(namedPath, func(subPath string, subPathFileInfo os.FileInfo, err error) error {
							if err != nil {
								return fmt.Errorf("can't stat %s: %w", subPath, err)
							}
							// We've already checked all files, so look in next dir.
							if subPathFileInfo.IsDir() {
								// Check for any JS files at this dir level.
								foundPath = findJSFile(subPath)
							}
							return nil
						})
						if findNamedPathErr != nil {
							return fmt.Errorf("Could not find related .js file from named import: %w%s",
								findNamedPathErr, common.Caller())
						}
					}
				}
				if foundPath != "" {
					// Remove "public" build dir from path.
					replacePath := strings.Replace(foundPath, buildPath, "", 1)
					// Wrap path in quotes.
					replacePath = "'" + replacePath + "'"
					// Convert string path to bytes.
					replacePathBytes := []byte(replacePath)
					// Find the specific import statement we're replacing.
					reFoundImport := regexp.MustCompile(string(staticStatement))
					// Actually replace the path to the dependency in the source content.
					contentBytes = reFoundImport.ReplaceAll(contentBytes, rePath.ReplaceAll(staticStatement, rePath.ReplaceAll(pathBytes, replacePathBytes)))
				}
			}
			// Overwrite the old file with the new content that contains the updated import path.
			err = ioutil.WriteFile(convertPath, contentBytes, 0644)
			if err != nil {
				return fmt.Errorf("Could not overwite %s with new import: %w%s", convertPath, err, common.Caller())
			}
		}
		return nil
	})
	if convertErr != nil {
		return fmt.Errorf("Could not convert file to support esm: %w%s", convertErr, common.Caller())
	}
	return nil

}

// Checks for a JS file in the directory given.
func findJSFile(path string) string {
	var foundPath string
	files, err := ioutil.ReadDir(path)
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".js" || filepath.Ext(f.Name()) == ".mjs" {
			foundPath = path + "/" + f.Name()
			Log("The found import path to use is: " + foundPath)
		}
	}
	if err != nil {
		fmt.Printf("Could not read files in current dir: %s", err)
	}
	return foundPath
}
