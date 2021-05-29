package build

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/plentico/plenti/common"
	"github.com/plentico/plenti/readers"
)

var (
	// Match dynamic import statments, e.g. import("") or import('').
	reDynamicImport = regexp.MustCompile(`import\((?:'|").*(?:'|")\)`)

	// Find any import statement in the file (including multiline imports).
	// () = brackets for grouping
	// \s = space
	// .* = any character
	// | = or statement
	// \n = newline
	// {0,} = repeat any number of times
	// \{ = just a closing curly bracket (escaped)
	reStaticImportGoPk = regexp.MustCompile(`(?m)^import(\s)(.*from(.*);|((.*\n){0,})\}(\s)from(.*);)`)
	reStaticExportGoPk = regexp.MustCompile(`export(\s)(.*from(.*);|((.*\n){0,})\}(\s)from(.*);)`)
	// Find the path specifically (part between single or double quotes).
	rePath = regexp.MustCompile(`(?:'|").*(?:'|")`)
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
		nodeModuleErr := filepath.WalkDir("node_modules/"+module, func(modulePath string, moduleFileInfo fs.DirEntry, err error) error {

			if err != nil {
				return fmt.Errorf("can't stat %s: %w", modulePath, err)
			}
			// Only get ESM supported files.
			if !moduleFileInfo.IsDir() && filepath.Ext(modulePath) == ".mjs" {
				from, err := os.Open(modulePath)
				if err != nil {
					return fmt.Errorf("Could not open source .mjs %s file for copying: %w%s\n", modulePath, err, common.Caller())
				}
				defer from.Close()

				// Remove "node_modules" from path and add "web_modules".
				outPathFile := gopackDir + strings.Replace(modulePath, "node_modules", "", 1)

				if err != nil {
					return fmt.Errorf("Could not open source .mjs %s file for copying: %w%s\n", modulePath, err, common.Caller())
				}
				// should do this maybe just once for node files separate to regular gopack files
				if common.UseMemFS {
					// This is a naive approach as maybe an npm update would change content or by hand.
					// The issue is this overwrites the map content if we keeping reading from disk and any errors in GoPack will leave bad import paths.
					// proabably need to hash but somewhere else...
					if v := common.Get(outPathFile); v != nil {
						v.Processed = true
						return nil
					}

					b, err := ioutil.ReadAll(from)
					if err != nil {
						return fmt.Errorf("Could not read source .mjs %s file for copying: %w%s\n", modulePath, err, common.Caller())
					}

					// newH := common.CRC32Hasher(b)
					// no change to content so  set processed to true and we wont gopack again...

					// no hashing as there are npm package files. We Gopack them so woukld always be different after
					common.Set(outPathFile,
						modulePath,
						&common.FData{B: b})

					return nil
				}
				// Create any subdirectories need to write file to "web_modules" destination.
				if err = os.MkdirAll(filepath.Dir(outPathFile), os.ModePerm); err != nil {
					return fmt.Errorf("Could not create subdirectories %s: %w%s\n", filepath.Dir(modulePath), err, common.Caller())
				}
				to, err := os.Create(outPathFile)
				if err != nil {
					return fmt.Errorf("Could not create destination %s file for copying: %w%s\n", modulePath, err, common.Caller())
				}
				defer to.Close()

				_, err = io.Copy(to, from)
				if err != nil {
					return fmt.Errorf("Could not copy .mjs  from source to destination: %w%s\n", err, common.Caller())
				}
			}
			return nil
		})
		if nodeModuleErr != nil {
			return fmt.Errorf("Could not get node module: %w%s\n", nodeModuleErr, common.Caller())
		}

	}
	if common.UseMemFS {

		// log n start lookup
		for convertPath := range common.StartFrom(buildPath + "/spa") {

			// end of "dir(s)"
			if !strings.HasPrefix(convertPath, buildPath+"/spa") {
				return nil
			}

			// todo: make sure entries/map are in sync i.e add proper delete logic

			if v := common.Get(convertPath); v != nil && !v.Processed {

				if err := runPack(buildPath, convertPath); err != nil {
					return err

				}
			}
		}

		return nil
	}
	convertErr := filepath.WalkDir(buildPath+"/spa", func(convertPath string, convertFileInfo fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", convertPath, err)
		}
		if convertFileInfo.IsDir() {
			return nil
		}

		return runPack(buildPath, convertPath)

	})

	return convertErr

}

func runPack(buildPath, convertPath string) error {

	if filepath.Ext(convertPath) != ".js" && filepath.Ext(convertPath) != ".mjs" {
		return nil

	}
	var contentBytes []byte
	var err error

	if !common.UseMemFS {
		contentBytes, err = ioutil.ReadFile(convertPath)
		if err != nil {
			return fmt.Errorf("Could not read file %s to convert to esm: %w%s\n", convertPath, err, common.Caller())
		}
	} else {
		contentBytes = common.Get(convertPath).B
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
	// Get all the export statements.
	staticExportStatements := reStaticExportGoPk.FindAll(contentBytes, -1)
	// Get all import and export statements.
	allStaticStatements := append(staticImportStatements, staticExportStatements...)
	for _, staticStatement := range allStaticStatements {

		// Get path from the full import/export statement.
		pathBytes := rePath.Find(staticStatement)
		// Convert path to a string.
		pathStr := string(pathBytes)
		// Remove single or double quotes around path.
		pathStr = strings.Trim(pathStr, `'"`)

		// Intialize the path that we are replacing.
		var foundPath string

		// Convert .svelte file extensions to .js so the browser can read them.
		if filepath.Ext(pathStr) == ".svelte" {
			foundPath = strings.Replace(pathStr, ".svelte", ".js", 1)
		}

		// Make sure the import/export path doesn't start with a dot (.) or double dot (..)
		// and make sure that the path doesn't have a file extension.
		if pathStr[:1] != "." && filepath.Ext(pathStr) == "" {
			if foundPath, err = checkNpmPath(buildPath, pathStr); err != nil {
				return err
			}
			// Make absolute foundPath relative to the current file so it works with baseurls.
			foundPath, err = filepath.Rel(path.Dir(convertPath), foundPath)
			if err != nil {
				fmt.Printf("Could not make path to NPM dependency relative: %s", err)
			}
		}

		if foundPath != "" {
			// Remove "public" build dir from path.
			//replacePath := filepath.Clean(strings.Replace(foundPath, buildPath, "", 1))
			replacePath := strings.Replace(foundPath, buildPath, "", 1)
			// Wrap path in quotes.
			replacePath = "'" + replacePath + "'"
			// Convert string path to bytes.
			replacePathBytes := []byte(replacePath)
			// Actually replace the path to the dependency in the source content.
			contentBytes = bytes.ReplaceAll(contentBytes, staticStatement,
				rePath.ReplaceAll(staticStatement, rePath.ReplaceAll(pathBytes, replacePathBytes)))

		}
	}
	if common.UseMemFS {
		// Overwrite the old file with the new content that contains the updated import path.
		common.Set(convertPath, "", &common.FData{B: contentBytes, Hash: common.CRC32Hasher(contentBytes)})
		return nil
	}
	// Overwrite the old file with the new content that contains the updated import path.
	err = ioutil.WriteFile(convertPath, contentBytes, 0644)
	if err != nil {
		return fmt.Errorf("Could not overwite %s with new import: %w%s\n", convertPath, err, common.Caller())
	}
	return nil

}

func checkNpmPath(buildPath, pathStr string) (string, error) {
	// A named import/export is being used, look for this in "web_modules/" dir.
	namedPath := buildPath + "/spa/web_modules/" + pathStr

	// Check all files in the current directory first.
	foundPath, err := findJSFile(namedPath)
	if err != nil {
		return "", err
	}

	// our loop goes till we have no matching prefix in SeacrhPath so this is as far as that goes.
	if !common.UseMemFS && foundPath == "" {
		// If JS file was not found in the current directory, check nested directories.
		findNamedPathErr := filepath.WalkDir(namedPath, func(subPath string, subPathFileInfo fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("can't stat %s: %w%s\n", subPath, err, common.Caller())
			}
			// We've already checked all files, so look in next dir.
			if subPathFileInfo.IsDir() {
				// Check for any JS files at this dir level.
				// should stop on success?
				foundPath, err = findJSFile(subPath)
				if err != nil {
					return err
				}

			}
			return nil
		})
		if findNamedPathErr != nil {
			return "", fmt.Errorf("Could not find related .js file from named import: %w%s\n",
				findNamedPathErr, common.Caller())
		}
	}
	return foundPath, nil
}

// Checks for a JS file in the directory given.
func findJSFile(path string) (string, error) {

	if common.UseMemFS {
		return common.SearchPath(path)
	}

	var foundPath string
	files, err := os.ReadDir(path)
	if err != nil {
		return "", fmt.Errorf("Could not read files in current dir: %s %w%s\n", path, err, common.Caller())
	}
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".js" || filepath.Ext(f.Name()) == ".mjs" {
			foundPath = path + "/" + f.Name()
		}
	}

	return foundPath, nil
}
