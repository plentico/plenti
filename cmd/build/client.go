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

	fmt.Println("\nBuilding client SPA using svelte compiler")

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

	copiedSourceCounter := 0
	compiledComponentCounter := 0
	clientBuildStr := "["

	// Go through all file paths in the "/layout" folder.
	layoutFilesErr := filepath.Walk("layout", func(layoutPath string, layoutFileInfo os.FileInfo, err error) error {
		// Create destination path.
		destFile := buildPath + strings.Replace(layoutPath, "layout", "/spa", 1)
		// Make sure path is a directory
		if layoutFileInfo.IsDir() {
			// Create any sub directories need for filepath.
			os.MkdirAll(destFile, os.ModePerm)
		}
		// Make list of files not to copy to build.
		excludedFiles := []string{
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
			fileContentStr = strings.Replace(fileContentStr, "from \"svelte/internal\";", "from \"../web_modules/svelte/internal/index.js\";", -1)
			fileContentStr = strings.Replace(fileContentStr, "from \"navaid\";", "from \"../web_modules/navaid.js\";", -1)
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

			destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".js"

			clientBuildStr = clientBuildStr + "{ \"component\": \"" + fileContentStr + "\", \"destPath\": \"" + destFile + "\", \"stylePath\": \"" + stylePath + "\"},"
			/*
							clientBuildStr = clientBuildStr + fmt.Sprintf(`{
					"component": "%s",
					"destPath": "%s",
					"stylePath": "%s"
				},`, escapedFileContentStr, destFile, stylePath)
			*/
			/*
				clientBuildStr = clientBuildStr + "{" +
					"\"component\": `" + escapedFileContentStr + "`," +
					"\"destPath\": \"" + destFile + "\"," +
					"\"stylePath\": \"" + stylePath + "\"},"
			*/
			/*
				// Execute node script to compile .svelte to .js
				compiledBytes, buildErr := exec.Command("node", "layout/ejected/build_client.js", fileContentStr).Output()
				if buildErr != nil {
					fmt.Printf("Could not compile svelte to JS: %s\n", buildErr)
				}

				compiledStr := string(compiledBytes)
				compiledStrArray := strings.Split(compiledStr, "!plenti-split!")

				// Get the JS only from the script output.
				jsStr := strings.TrimSpace(compiledStrArray[0])

				// Write compiled .js to build directory.
				jsBytes := []byte(jsStr)
				err := ioutil.WriteFile(destFile, jsBytes, 0755)
				if err != nil {
					fmt.Printf("Unable to write file: %v", err)
				}

				// Get the CSS only from the script output.
				cssStr := strings.TrimSpace(compiledStrArray[1])
				// If there is CSS, write it into the bundle.css file.
				if cssStr != "null" {
					cssFile, WriteStyleErr := os.OpenFile(stylePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if WriteStyleErr != nil {
						fmt.Printf("Could not open bundle.css for writing: %s", WriteStyleErr)
					}
					defer cssFile.Close()
					if _, err := cssFile.WriteString(cssStr); err != nil {
						log.Println(err)
					}
				}
			*/

			compiledComponentCounter++

		}
		return nil
	})

	clientBuildStr = strings.TrimSuffix(clientBuildStr, ",") + "]"

	if layoutFilesErr != nil {
		fmt.Printf("Could not get layout file: %s", layoutFilesErr)
	}

	fmt.Printf("Number of source files copied: %d\n", copiedSourceCounter)
	fmt.Printf("Number of components compiled: %d\n", compiledComponentCounter)

	return clientBuildStr

}
