package build

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rogchap.com/v8go"
)

// Client builds the SPA.
func Client(buildPath string) string {

	defer Benchmark(time.Now(), "Prepping client SPA data")

	Log("\nPrepping client SPA for svelte compiler")

	stylePath := buildPath + "/spa/bundle.css"

	// Set up counter for logging output.
	compiledComponentCounter := 0

	// Start the string that will be sent to nodejs for compiling.
	clientBuildStr := "["

	content, err := ioutil.ReadFile("node_modules/svelte/compiler.js")
	if err != nil {
		log.Fatal(err)
	}
	// Remove reference to 'self' that breaks v8go.
	contentStr := strings.Replace(string(content), "self.performance.now();", "'';", 1)
	ctx, _ := v8go.NewContext(nil)
	ctx.RunScript(contentStr, "ejected/bundle.js")

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

				//val, err := ctx.RunScript("component='"+layoutPath+"'", "ejected/bundle.js")
				val, err := ctx.RunScript("svelte.compile('"+layoutPath+"', {css: false});", "ejected/bundle.js")
				if err != nil {
					fmt.Printf("V8go could not execute: %v", err)
				}
				fmt.Println(val)
				//fmt.Println(string(bundledContent))

				// Create string representing array of objects to be passed to nodejs.
				//clientBuildStr = clientBuildStr + "{ \"layoutPath\": \"" + layoutPath + "\", \"destPath\": \"" + destFile + "\", \"stylePath\": \"" + stylePath + "\"},"

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

	Log("Number of components to be compiled: " + strconv.Itoa(compiledComponentCounter))

	return clientBuildStr

}
