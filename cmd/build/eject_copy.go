package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// EjectCopy does a direct copy of any ejectable js files needed in spa build dir.
func EjectCopy(buildPath string, tempBuildDir string, ejectedDir string) {

	defer Benchmark(time.Now(), "Copying ejectable core files for build")

	Log("\nCopying ejectable core files to their destination:")

	copiedSourceCounter := 0

	ejectedFilesErr := filepath.Walk(ejectedDir, func(ejectPath string, ejectFileInfo os.FileInfo, err error) error {
		// Make list of files not to copy to build.
		excludedFiles := []string{
			ejectedDir + "/build.js",
		}
		// Check if the current file is in the excluded list.
		excluded := false
		for _, excludedFile := range excludedFiles {
			if excludedFile == ejectPath {
				excluded = true
			}
		}
		// If the file is already in .js format just copy it straight over to build dir.
		if filepath.Ext(ejectPath) == ".js" && !excluded {

			destPath := buildPath + "/spa/"
			os.MkdirAll(destPath+strings.TrimPrefix(ejectedDir, tempBuildDir), os.ModePerm)

			from, err := os.Open(ejectPath)
			if err != nil {
				fmt.Printf("Could not open source .js file for copying: %s\n", err)
			}
			defer from.Close()

			to, err := os.Create(destPath + strings.TrimPrefix(ejectPath, tempBuildDir))
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
		return nil
	})
	if ejectedFilesErr != nil {
		fmt.Printf("Could not get ejectable file: %s", ejectedFilesErr)
	}

	Log("Number of ejectable core files copied: " + strconv.Itoa(copiedSourceCounter))

}
