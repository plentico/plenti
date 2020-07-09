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
func Client(buildPath string) {

	defer Benchmark(time.Now(), "Compiling client SPA with Svelte")

	Log("\nCompiling client SPA with svelte")

	stylePath := buildPath + "/spa/bundle.css"

	// Set up counter for logging output.
	compiledComponentCounter := 0

	// Get svelte compiler code from node_modules.
	compiler, err := ioutil.ReadFile("node_modules/svelte/compiler.js")
	if err != nil {
		fmt.Printf("Can't read node_modules/svelte/compiler.js: %v", err)
	}
	// Remove reference to 'self' that breaks v8go.
	compilerStr := strings.Replace(string(compiler), "self.performance.now();", "'';", 1)
	ctx, _ := v8go.NewContext(nil)
	ctx.RunScript(compilerStr, "compile_svelte")
	compileSvelte(ctx, "ejected/router.svelte", buildPath+"/spa/ejected/router.js", stylePath)

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

				compileSvelte(ctx, layoutPath, destFile, stylePath)

				compiledComponentCounter++

			}
		}
		return nil
	})
	if layoutFilesErr != nil {
		fmt.Printf("Could not get layout file: %s", layoutFilesErr)
	}

	Log("Number of components compiled: " + strconv.Itoa(compiledComponentCounter))

}

func compileSvelte(ctx *v8go.Context, layoutPath string, destFile string, stylePath string) {

	component, err := ioutil.ReadFile(layoutPath)
	if err != nil {
		fmt.Printf("Can't read component: %v", err)
	}
	componentStr := string(component)

	// Compile component with Svelte.
	ctx.RunScript("var { js, css } = svelte.compile(`"+componentStr+"`, {css: false});", "compile_svelte")

	// Get the JS code from the compiled result.
	jsCode, err := ctx.RunScript("js.code;", "compile_svelte")
	if err != nil {
		fmt.Printf("V8go could not execute js.code: %v", err)
	}
	jsBytes := []byte(jsCode.String())
	jsWriteErr := ioutil.WriteFile(destFile, jsBytes, 0755)
	if jsWriteErr != nil {
		fmt.Printf("Unable to write file: %v", jsWriteErr)
	}

	// Get the CSS code from the compiled result.
	cssCode, err := ctx.RunScript("css.code;", "compile_svelte")
	if err != nil {
		fmt.Printf("V8go could not execute css.code: %v", err)
	}
	cssStr := strings.TrimSpace(cssCode.String())
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
}
