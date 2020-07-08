package build

import (
	"fmt"
	"io/ioutil"
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

	// Get svelte compiler code from node_modules.
	compiler, err := ioutil.ReadFile("node_modules/svelte/compiler.js")
	if err != nil {
		fmt.Printf("Can't read node_modules/svelte/compiler.js: %v", err)
	}
	// Remove reference to 'self' that breaks v8go.
	compilerStr := strings.Replace(string(compiler), "self.performance.now();", "'';", 1)
	ctx, _ := v8go.NewContext(nil)
	ctx.RunScript(compilerStr, "ejected/bundle.js")

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

				ctx.RunScript("var { js, css } = svelte.compile('"+layoutPath+"', {css: false});", "ejected/bundle.js")
				jsCode, err := ctx.RunScript("js.code;", "ejected/bundle.js")
				if err != nil {
					fmt.Printf("V8go could not execute js.code: %v", err)
				}
				fmt.Printf("js code: %v", jsCode)
				cssCode, err := ctx.RunScript("css.code;", "ejected/bundle.js")
				if err != nil {
					fmt.Printf("V8go could not execute css.code: %v", err)
				}
				fmt.Printf("css code: %v", cssCode)

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
