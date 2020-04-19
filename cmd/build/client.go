package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Client builds the SPA.
func Client(buildPath string) {

	fmt.Println("\nBuilding client SPA using svelte compiler")

	var wg sync.WaitGroup

	stylePath := buildPath + "/spa/bundle.css"
	// Clear out any previous CSS.
	if _, stylePathExistsErr := os.Stat(stylePath); stylePathExistsErr == nil {
		deleteStyleErr := os.Remove(stylePath)
		if deleteStyleErr != nil {
			fmt.Println(deleteStyleErr)
			return
		}
	}

	copiedSourceCounter := 0
	compiledComponentCounter := 0

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

			wg.Add(1)
			go compileSvelte(fileContentStr, destFile, stylePath, &wg)

			compiledComponentCounter++

		}
		return nil
	})
	if layoutFilesErr != nil {
		fmt.Printf("Could not get layout file: %s", layoutFilesErr)
	}

	wg.Wait()

	fmt.Printf("Number of source files copied: %d\n", copiedSourceCounter)
	fmt.Printf("Number of components compiled: %d\n", compiledComponentCounter)

}

func compileSvelte(fileContentStr string, destFile string, stylePath string, wg *sync.WaitGroup) {

	// Execute node script to compile .svelte to .js
	compiledBytes, buildErr := exec.Command("node", "layout/ejected/build_client.js", fileContentStr).Output()
	if buildErr != nil {
		fmt.Printf("Could not compile svelte to JS: %s\n", buildErr)
	}
	compiledStr := string(compiledBytes)
	compiledStrArray := strings.Split(compiledStr, "!plenti-split!")

	// Get the JS only from the script output.
	jsStr := strings.TrimSpace(compiledStrArray[0])
	// Convert file extensions to be snowpack friendly.
	jsStr = strings.Replace(jsStr, ".svelte", ".js", -1)
	jsStr = strings.Replace(jsStr, "from \"svelte/internal\";", "from \"../web_modules/svelte/internal/index.js\";", -1)
	jsStr = strings.Replace(jsStr, "from \"navaid\";", "from \"../web_modules/navaid.js\";", -1)

	// Write compiled .js to build directory.
	jsBytes := []byte(jsStr)
	destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".js"
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
	wg.Done()
}
