package build

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
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
			log.Printf("Could not open asset for copying: %v\n", err)
			return err
		}
		defer from.Close()

		to, err := os.Create(destPath)
		if err != nil {
			log.Printf("Could not create destination asset for copying: %v\n", err)
			return err
		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			log.Printf("Could not copy asset from source to destination: %v\n", err)
			return err
		}

		copiedSourceCounter++
		return nil
	})
	if err != nil {
		log.Printf("Could not get asset file: %s", err)
		return err
	}

	Log("Number of assets copied: " + strconv.Itoa(copiedSourceCounter))
	return nil

}
