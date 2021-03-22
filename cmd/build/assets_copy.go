package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/plentico/plenti/common"
)

// AssetsCopy does a direct copy of any static assets.
func AssetsCopy(buildPath string, tempBuildDir string) error {

	defer Benchmark(time.Now(), "Copying static assets into build dir")

	Log("\nCopying static assets:")

	copiedSourceCounter := 0

	assetsDir := tempBuildDir + "assets"

	// Exit function if "assets/" directory does not exist.
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist: %w", assetsDir, err)
	}

	err := filepath.Walk(assetsDir, func(assetPath string, assetFileInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", assetPath, err)
		}
		destPath := buildPath + "/" + strings.TrimPrefix(assetPath, tempBuildDir)
		if assetFileInfo.IsDir() {
			// Make directory if it doesn't exist.
			// Move on to next path.
			if err = os.MkdirAll(destPath, os.ModePerm); err != nil {
				return fmt.Errorf("cannot create asset dir %s: %w", assetPath, err)
			}
			return nil

		}
		from, err := os.Open(assetPath)
		if err != nil {
			return fmt.Errorf("Could not open asset %s for copying: %w%s", assetPath, err, common.Caller())

		}
		defer from.Close()

		to, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("Could not create destination asset %s for copying: %w%s", destPath, err, common.Caller())

		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			return fmt.Errorf("Could not copy asset from source %s to destination: %w%s", assetPath, err, common.Caller())

		}

		copiedSourceCounter++
		return nil
	})
	if err != nil {
		return fmt.Errorf("Could not get asset file: %w%s", err, common.Caller())

	}

	Log(fmt.Sprintf("Number of assets copied: %d", copiedSourceCounter))
	return nil

}
