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

// AssetsCopy does a direct copy of any static assets.
func AssetsCopy(buildPath string, tempBuildDir string) {

	defer Benchmark(time.Now(), "Copying static assets into build dir")

	Log("\nCopying static assets:")

	copiedSourceCounter := 0

	assetsDir := tempBuildDir + "assets"

	// Exit function if "assets/" directory does not exist.
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		return
	}

	assetFilesErr := filepath.Walk(assetsDir, func(assetPath string, assetFileInfo os.FileInfo, err error) error {
		destPath := buildPath + "/" + strings.TrimPrefix(assetPath, tempBuildDir)
		if assetFileInfo.IsDir() {
			// Make directory if it doesn't exist.
			os.MkdirAll(destPath, os.ModePerm)
			// Move on to next path.
			return nil
		}
		from, err := os.Open(assetPath)
		if err != nil {
			fmt.Printf("Could not open asset for copying: %s\n", err)
		}
		defer from.Close()

		to, err := os.Create(destPath)
		if err != nil {
			fmt.Printf("Could not create destination asset for copying: %s\n", err)
		}
		defer to.Close()

		_, fileCopyErr := io.Copy(to, from)
		if err != nil {
			fmt.Printf("Could not copy asset from source to destination: %s\n", fileCopyErr)
		}

		copiedSourceCounter++
		return nil
	})
	if assetFilesErr != nil {
		fmt.Printf("Could not get asset file: %s", assetFilesErr)
	}

	Log("Number of assets copied: " + strconv.Itoa(copiedSourceCounter))

}
