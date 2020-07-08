package build

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rogchap.com/v8go"
)

// Client builds the SPA.
func Client(buildPath string, bundledContent []byte) string {

	defer Benchmark(time.Now(), "Prepping client SPA data")

	Log("\nPrepping client SPA for svelte compiler")

	stylePath := buildPath + "/spa/bundle.css"

	// Set up counter for logging output.
	compiledComponentCounter := 0

	// Start the string that will be sent to nodejs for compiling.
	clientBuildStr := "["

	/*
		content, err := ioutil.ReadFile("ejected/bundle.js")
		if err != nil {
			fmt.Printf("Could not read ejected/bundle.js file: %v\n", err)
		}
		contentStr := "var self = this;" + string(content)
	*/
	bundledContent = bytes.Replace(bundledContent, []byte("(() => {"), []byte(""), 1)
	bundledContent = bytes.Replace(bundledContent, []byte("})();"), []byte(""), 1) // TODO: Only replace the last instance of this string.
	fmt.Println(string(bundledContent))
	ctx, _ := v8go.NewContext(nil)
	ctx.RunScript(string(bundledContent), "ejected/bundle.js")
	val, _ := ctx.RunScript("var component='';", "ejected/bundle.js")
	fmt.Println(val)

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

				val, err := ctx.RunScript("component='"+layoutPath+"'", "ejected/bundle.js")
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
