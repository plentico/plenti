package build

import (
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Client builds the SPA.
func Client(buildPath string) string {

	fmt.Println("\nPrepping client SPA for svelte compiler")

	stylePath := buildPath + "/spa/bundle.css"
	// Clear out any previous CSS.
	/*
		if _, stylePathExistsErr := os.Stat(stylePath); stylePathExistsErr == nil {
			deleteStyleErr := os.Remove(stylePath)
			if deleteStyleErr != nil {
				fmt.Println(deleteStyleErr)
				return
			}
		}
	*/

	// Set up counters for logging output.
	copiedSourceCounter := 0
	compiledComponentCounter := 0

	// Start the string that will be sent to nodejs for compiling.
	clientBuildStr := "["

	// Go through all file paths in the "/layout" folder.
	layoutFilesErr := filepath.Walk("layout", func(layoutPath string, layoutFileInfo os.FileInfo, err error) error {
		// Create destination path.
		destFile := buildPath + strings.Replace(layoutPath, "layout", "/spa", 1)
		// Make sure path is a directory
		if layoutFileInfo.IsDir() {
			// Create any sub directories need for filepath.
			os.MkdirAll(destFile, os.ModePerm)
		} else {
			// Make list of files not to copy to build.
			excludedFiles := []string{
				"layout/ejected/build.js",
				"layout/ejected/build_client.js",
				"layout/ejected/build_static.js",
				"layout/ejected/server_router.js",
			}
			// Check if the current file is in the excluded list.
			excluded := false
			for _, excludedFile := range excludedFiles {
				if excludedFile == layoutPath {
					excluded = true
				}
			}
			// If the file is already in .js format just copy it straight over to build dir.
			if filepath.Ext(layoutPath) == ".js" && !excluded {
				from, err := os.Open(layoutPath)
				if err != nil {
					fmt.Printf("Could not open source .js file for copying: %s\n", err)
				}
				defer from.Close()

				to, err := os.Create(destFile)
				if err != nil {
					fmt.Printf("Could not create destination .js file for copying: %s\n", err)
				}
				defer to.Close()

				_, fileCopyErr := io.Copy(to, from)
				if err != nil {
					fmt.Printf("Could not copy .js from source to destination: %s\n", fileCopyErr)
				}

				copiedSourceCounter++

			}

			// If the file is in .svelte format, compile it to .js
			if filepath.Ext(layoutPath) == ".svelte" {

				fileContentByte, readFileErr := ioutil.ReadFile(layoutPath)
				if readFileErr != nil {
					fmt.Printf("Could not read contents of svelte source file: %s\n", readFileErr)
				}
				fileContentStr := string(fileContentByte)
				// Convert file extensions to be snowpack friendly.
				fileContentStr = strings.Replace(fileContentStr, ".svelte", ".js", -1)
				//fileContentStr = strings.Replace(fileContentStr, "from \"svelte/internal\";", "from \"../web_modules/svelte/internal/index.js\";", -1)
				fileContentStr = strings.Replace(fileContentStr, "from \"navaid\";", "from \"../web_modules/navaid.js\";", -1)

				// Encode HTML so it can be represented as a string.
				fileContentStr = html.EscapeString(fileContentStr)

				// Remove newlines.
				reN := regexp.MustCompile(`\r?\n`)
				fileContentStr = reN.ReplaceAllString(fileContentStr, " ")
				// Remove tabs.
				reT := regexp.MustCompile(`\t`)
				fileContentStr = reT.ReplaceAllString(fileContentStr, " ")
				// Reduce extra whitespace to a single space.
				reS := regexp.MustCompile(`\s+`)
				fileContentStr = reS.ReplaceAllString(fileContentStr, " ")

				// Convert opening curly brackets to HTML escape character.
				fileContentStr = strings.Replace(fileContentStr, "{", "&#123;", -1)
				// Convert closing curly brackets to HTML escape character.
				fileContentStr = strings.Replace(fileContentStr, "}", "&#125;", -1)

				// Replace .svelte file extension with .js.
				destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".js"

				// Create string representing array of objects to be passed to nodejs.
				clientBuildStr = clientBuildStr + "{ \"component\": \"" + fileContentStr + "\", \"destPath\": \"" + destFile + "\", \"stylePath\": \"" + stylePath + "\"},"

				compiledComponentCounter++

			}
		}
		return nil
	})
	if layoutFilesErr != nil {
		fmt.Printf("Could not get layout file: %s", layoutFilesErr)
	}

	// End the string that will be sent to nodejs for compiling.
	clientBuildStr = strings.TrimSuffix(clientBuildStr, ",") + "]"

	fmt.Printf("Number of source files copied: %d\n", copiedSourceCounter)
	fmt.Printf("Number of components to be compiled: %d\n", compiledComponentCounter)

	return clientBuildStr

}
