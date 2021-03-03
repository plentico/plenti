package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/plentico/plenti/common"
)

// NodeClient preps the client SPA for execution via NodeJS (NOTE: This is legacy functionality).
func NodeClient(buildPath string) (string, error) {

	defer Benchmark(time.Now(), "Prepping client SPA data")

	Log("\nPrepping client SPA for svelte compiler")

	stylePath := buildPath + "/spa/bundle.css"

	// Set up counter for logging output.
	compiledComponentCounter := 0

	// Start the string that will be sent to nodejs for compiling.
	clientBuildStr := "["

	// Go through all file paths in the "/layout" folder.
	layoutFilesErr := filepath.Walk("layout", func(layoutPath string, layoutFileInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", layoutPath, err)
		}
		// Create destination path.
		destFile := buildPath + strings.Replace(layoutPath, "layout", "/spa", 1)
		// Make sure path is a directory
		if layoutFileInfo.IsDir() {
			// Create any sub directories need for filepath.
			if err = os.MkdirAll(destFile, os.ModePerm); err != nil {
				return fmt.Errorf("cannot create sub directories need for filepath %s: %w%s",
					destFile, err, common.Caller())
			}
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
		return "", fmt.Errorf("Could not get layout file: %w%s", layoutFilesErr, common.Caller())
	}

	// Get router from ejected core. NOTE if you remove this, trim the trailing comma below.
	clientBuildStr = clientBuildStr + "{ \"layoutPath\": \"ejected/router.svelte\", \"destPath\": \"" + buildPath + "/spa/ejected/router.js\", \"stylePath\": \"" + stylePath + "\"}"

	// End the string that will be sent to nodejs for compiling.
	//clientBuildStr = strings.TrimSuffix(clientBuildStr, ",") + "]"
	clientBuildStr = clientBuildStr + "]"

	Log("Number of components to be compiled: " + strconv.Itoa(compiledComponentCounter))

	return clientBuildStr, nil

}
