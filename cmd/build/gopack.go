package build

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/plentico/plenti/readers"
)

var (
	// Regexp help:
	// () = brackets for grouping
	// \s = space
	// .* = any character
	// | = or statement
	// \n = newline
	// {0,} = repeat any number of times
	// \{ = just a closing curly bracket (escaped)
	// ? = make quantifier non-greedy (stop at first match)
	// ?: = non-capturing group (https://stackoverflow.com/questions/3512471/what-is-a-non-capturing-group-in-regular-expressions)
	// m = multiline mode (example: https://go.dev/play/p/Fq14BXWuGH)

	// Match dynamic import statments, e.g. import("") or import('').
	reDynamicImport = regexp.MustCompile(`import\((?:'|").*(?:'|")\)`)
	// Find any import statement in the file (including multiline imports).
	reStaticImportGoPk = regexp.MustCompile(`(?m)^import(\s)(.*from(.*);|((.*\n){0,}?)\}(\s)from(.*);)`)
	// Find any 'side-effects only' imports (e.g. import './my-module.js';)
	reSideEffectsImportGoPk = regexp.MustCompile(`(?m)^import(\s)('|")(.*?)('|");`)
	// Find all exported import statements (e.g. export { onDestroy, onMount } from './internal';)
	reStaticExportGoPk = regexp.MustCompile(`export(\s)(.*from(.*);|((.*\n){0,}?)\}(\s)from(.*);)`)
	// Find the path specifically (part between single or double quotes).
	rePath = regexp.MustCompile(`(?:'|").*(?:'|")`)
	// Retrieve minifier and load JS
	m = loadJSMinifier()
)

// Create global var since cmd.minifyFlag is a circular dependency.
var minifyFlag bool

// CheckMinifyFlag sets global var if --minify flag is passed.
func CheckMinifyFlag(flag bool) {
	// If --minify flag is passed by user, this will be set to true.
	minifyFlag = flag
}

// Initialize globally to keep track during recursion.
var alreadyConvertedFiles []string

// Gopack ensures ESM support for NPM dependencies.
func Gopack(buildPath string) error {

	defer Benchmark(time.Now(), "Running Gopack")

	Log("\nRunning gopack to build esm support for npm dependencies")

	// Clear web_modules from previous build.
	alreadyConvertedFiles = []string{}

	// Start at the entry point for the app
	err := runPack(buildPath, buildPath+"/spa/ejected/main.js")
	if err != nil {
		return err
	}

	return nil
}

func runPack(buildPath, convertPath string) error {
	// Destination path for dependencies
	gopackDir := buildPath + "/spa/web_modules"

	// Get the actual contents of the file we want to convert
	contentBytes, err := ioutil.ReadFile(convertPath)
	if err != nil {
		return fmt.Errorf("\nCould not read file %s to convert to esm\n%w", convertPath, err)
	}

	// Created byte array of all dynamic imports in the current file.
	dynamicImportPaths := reDynamicImport.FindAll(contentBytes, -1)
	for _, dynamicImportPath := range dynamicImportPaths {
		// Inside the dynamic import change any svelte file extensions to reference regular javascript files.
		fixedImportPath := bytes.Replace(dynamicImportPath, []byte(".svelte"), []byte(".js"), 1)
		// Add the updated import back into the file contents for writing later.
		contentBytes = bytes.Replace(contentBytes, dynamicImportPath, fixedImportPath, 1)
	}

	// Get all the import statements.
	staticImportStatements := reStaticImportGoPk.FindAll(contentBytes, -1)
	// Get all the exported import statements.
	staticExportStatements := reStaticExportGoPk.FindAll(contentBytes, -1)
	// Get all the side-effects only static import statements
	sideEffectImportStatements := reSideEffectsImportGoPk.FindAll(contentBytes, -1)
	// Combine import and export statements.
	allStaticStatements := append(append(staticImportStatements, staticExportStatements...), sideEffectImportStatements...)
	// Iterate through all static import and export statements.
	for _, staticStatement := range allStaticStatements {
		// Get path from the full import/export statement.
		pathBytes := rePath.Find(staticStatement)
		// Convert path to a string.
		pathStr := string(pathBytes)
		// Remove single or double quotes around path.
		pathStr = strings.Trim(pathStr, `'"`)
		// Intitialize the string that determines if we found the import path.
		var foundPath string
		// Initialize the full path of the import.
		var fullPathStr string

		// Convert .svelte file extensions to .js so the browser can read them.
		if filepath.Ext(pathStr) == ".svelte" {
			// Declare found since path should be relative to the component it's referencing.
			pathStr = strings.Replace(pathStr, ".svelte", ".js", 1)
		}

		// If relative import (catches both previously .svelte paths and those already in .js format)
		if pathStr[:1] == "." {
			// Make relative pathStr a full path that we can find on the filesystem.
			fullPathStr = path.Clean(path.Dir(convertPath) + "/" + pathStr)

			if filepath.Ext(fullPathStr) == "" {
				// The path is a directory
				// Assume npm module since imports from Plenti project should be complete and not need package.json resolution
				modulePath := "node_modules" + strings.TrimPrefix(fullPathStr, gopackDir)
				// Try to get the package.json and locate the appropriate file
				entryPoint, err := getModuleEntrypoint(modulePath)
				if err != nil {
					fmt.Printf("Could not get entrypoint: %s", err)
				}

				src := path.Join(modulePath, entryPoint)
				dest := path.Join(fullPathStr, entryPoint)

				// Set to actual file instead of parent folder
				fullPathStr = dest

				// Get the module from npm if we haven't already
				if !pathExists(dest) {
					err = copyFile(src, dest)
					if err != nil {
						fmt.Printf("Can't copy module for submodule: %s\n", err)
					}
				}
				// path.Join removes proceeding ./ from path so we need to retain it manually when adding the entryPoint
				fixCurrentDir := ""
				if strings.Split(pathStr, "/")[0] == "." {
					fixCurrentDir = strings.Split(pathStr, "/")[0] + "/"
				}
				foundPath = fixCurrentDir + path.Join(pathStr, entryPoint)
			} else if pathExists(fullPathStr) {
				// We found the file in filesystem
				// Set this as a found path
				foundPath = pathStr
			} else if strings.HasPrefix(convertPath, gopackDir) {
				// The relative import is coming from a web_module itself
				// Change out of public/spa/web_modules and go into node_modules
				modulePath := "node_modules" + strings.TrimPrefix(fullPathStr, gopackDir)
				// Get the module from npm
				err = copyFile(modulePath, fullPathStr)
				if err != nil {
					fmt.Printf("Can't copy module for submodule: %s\n", err)
				}
				// Check if it can be found after being copied from 'node_modules'
				if pathExists(fullPathStr) {
					// Set this as a found path
					foundPath = pathStr
				}
			}
		}

		// Make sure the import/export path doesn't start with a dot (.) or double dot (..)
		// and make sure that the path doesn't have a file extension.
		if pathStr[:1] != "." && filepath.Ext(pathStr) == "" {
			// Copy the npm file from /node_modules to /spa/web_modules
			fullPathStr, err = copyNpmModule(pathStr, gopackDir)
			if err != nil {
				fmt.Printf("Can't copy npm module: %s", err)
			}
			if pathExists(fullPathStr) {
				// Make absolute path relative to the current file so it works with baseurls.
				foundPath, err = filepath.Rel(path.Dir(convertPath), fullPathStr)
				if err != nil {
					fmt.Printf("Could not make path to NPM dependency relative: %s", err)
				}
			}
		}

		// Do not convert files that have already been converted to avoid loops.
		if !alreadyConverted(fullPathStr, alreadyConvertedFiles) {
			// Add the current file to list of already converted files.
			alreadyConvertedFiles = append(alreadyConvertedFiles, fullPathStr)
			// Use fullPathStr recursively to find its imports.
			err = runPack(buildPath, fullPathStr)
			if err != nil {
				return fmt.Errorf("\nCan't runPack on %s %w", fullPathStr, err)
			}
		}

		if foundPath != "" {
			// Remove "public" build dir from path.
			replacePath := strings.Replace(foundPath, buildPath, "", 1)
			// Wrap path in quotes.
			replacePath = "'" + replacePath + "'"
			// Convert string path to bytes.
			replacePathBytes := []byte(replacePath)
			// Actually replace the path to the dependency in the source content.
			contentBytes = bytes.ReplaceAll(contentBytes, staticStatement,
				rePath.ReplaceAll(staticStatement, rePath.ReplaceAll(pathBytes, replacePathBytes)))
		} else {
			return fmt.Errorf("\nImport path '%s' not resolvable from file '%s'\n", pathStr, convertPath)
		}
	}
	if minifyFlag {
		contentBytes = minifyJS(m, contentBytes)
	}
	// Overwrite the old file with the new content that contains the updated import path.
	err = ioutil.WriteFile(convertPath, contentBytes, 0644)
	if err != nil {
		return fmt.Errorf("Could not overwite %s with new import: %w\n", convertPath, err)
	}
	return nil

}

func alreadyConverted(convertPath string, alreadyConvertedFiles []string) bool {
	// Check if there are already files that have been converted
	if len(alreadyConvertedFiles) > 0 {
		for _, convertedFile := range alreadyConvertedFiles {
			// Compare the currently queued file with each already converted file
			if convertPath == convertedFile {
				// Exit the function to avoid endless loops where files
				// reference each other (like main.js and router.svelte)
				return true
			}
		}
	}
	return false
}

func pathExists(path string) bool {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		// The path was found on the filesystem
		return true
	}
	return false
}

func copyNpmModule(module string, gopackDir string) (string, error) {

	modulePath := "node_modules/" + module

	entryPoint, err := getModuleEntrypoint(modulePath)
	if err != nil {
		return "", err
	}

	src := path.Clean(modulePath + "/" + entryPoint)
	dest := gopackDir + strings.TrimPrefix(src, "node_modules")
	if !pathExists(dest) {
		err := copyFile(src, dest)
		if err != nil {
			return dest, fmt.Errorf("Can't copy top level module: %s\n", err)
		}
	}

	return dest, nil

}

func getModuleEntrypoint(modulePath string) (string, error) {
	moduleConfigPath := modulePath + "/package.json"
	if pathExists(moduleConfigPath) {
		npmConfig := readers.GetNpmConfig(moduleConfigPath)
		if npmConfig.Module != "" {
			entryPoint := npmConfig.Module
			if filepath.Ext(entryPoint) == "" {
				// Add the .js file extension
				entryPoint += ".js"
			}
			return entryPoint, nil
		}
		return "", fmt.Errorf("Module not set in npm config file: %s", moduleConfigPath)
	}
	return "", fmt.Errorf("Could not get moduleConfigPath: %s", moduleConfigPath)
}

func copyFile(src string, dest string) error {

	from, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Could not open source .mjs '%s' file for copying: %s\n", src, err)
	}
	defer from.Close()

	// Create any subdirectories need to write file to "web_modules" destination.
	if err = os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return fmt.Errorf("Could not create subdirectories '%s': %s\n", filepath.Dir(src), err)
	}

	to, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("Could not create destination %s file for copying: %s\n", src, err)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return fmt.Errorf("Could not copy '%s' (source) to %s (destination): %s\n", src, dest, err)
	}
	return nil

}
