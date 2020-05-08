package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Client builds the SPA.
func Client(buildPath string) string {

	fmt.Println("\nPrepping client SPA for svelte compiler")

	stylePath := buildPath + "/spa/bundle.css"

	// Set up counter for logging output.
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
			// If the file is in .svelte format, compile it to .js
			if filepath.Ext(layoutPath) == ".svelte" {

				// Replace .svelte file extension with .js.
				destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".js"

				// Create string representing array of objects to be passed to nodejs.
				clientBuildStr = clientBuildStr + "{ \"layoutPath\": \"" + layoutPath + "\", \"destPath\": \"" + destFile + "\", \"stylePath\": \"" + stylePath + "\"},"

				compiledComponentCounter++

			}
		}
		return nil
	})
	if layoutFilesErr != nil {
		fmt.Printf("Could not get layout file: %s", layoutFilesErr)
	}

	// Get router from ejected core. NOTE if you remove this, trim the trailing comma below.
	clientBuildStr = clientBuildStr + "{ \"layoutPath\": \"ejected/router.svelte\", \"destPath\": \"" + buildPath + "/spa/ejected/router.js\", \"stylePath\": \"" + stylePath + "\"}"

	// End the string that will be sent to nodejs for compiling.
	//clientBuildStr = strings.TrimSuffix(clientBuildStr, ",") + "]"
	clientBuildStr = clientBuildStr + "]"

	fmt.Printf("Number of components to be compiled: %d\n", compiledComponentCounter)

	return clientBuildStr

}
