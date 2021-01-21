package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AssetsCopy does a direct copy of any static assets.
func AssetsCopy(buildPath string, tempBuildDir string) error {

	defer Benchmark(time.Now(), "Copying static assets into build dir")

	Log("\nCopying static assets:")

	copiedSourceCounter := 0

	assetsDir := tempBuildDir + "assets"

	// Exit function if "assets/" directory does not exist.
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		return err
	}

	err := filepath.Walk(assetsDir, func(assetPath string, assetFileInfo os.FileInfo, err error) error {
		destPath := buildPath + "/" + strings.TrimPrefix(assetPath, tempBuildDir)
		if assetFileInfo.IsDir() {
			// Make directory if it doesn't exist.
			// Move on to next path.
			return os.MkdirAll(destPath, os.ModePerm)

		}
		from, err := os.Open(assetPath)
		if err != nil {
			return fmt.Errorf("Could not open asset for copying: %w", err)

		}
		defer from.Close()

		to, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("Could not create destination asset for copying: %w", err)

		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			return fmt.Errorf("Could not copy asset from source to destination: %w", err)

		}

		copiedSourceCounter++
		return nil
	})
	if err != nil {
		return fmt.Errorf("Could not get asset file: %w", err)

	}

	Log(fmt.Sprintf("Number of assets copied: %d", copiedSourceCounter))
	return nil

}
