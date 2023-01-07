package build

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// StaticCopy does a direct copy of any static assets (e.g. global.css, robots.txt).
func StaticCopy(buildPath string) error {

	defer Benchmark(time.Now(), "Copying static files into build dir")

	Log("\nCopying static files:")

	staticDir := "static"
	copiedSourceCounter := 0
	var err error

	if ThemeFs != nil {
		copiedSourceCounter, err = copyStaticFilesFromTheme(staticDir, buildPath, copiedSourceCounter)
		if err != nil {
			return err
		}
	} else {
		copiedSourceCounter, err = copyStaticFilesFromProject(staticDir, buildPath, copiedSourceCounter)
		if err != nil {
			return err
		}
	}

	Log(fmt.Sprintf("Number of static files copied: %d", copiedSourceCounter))
	return nil

}

func copyStaticFilesFromTheme(staticDir string, buildPath string, copiedSourceCounter int) (int, error) {

	if err := afero.Walk(ThemeFs, staticDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fullPath := buildPath + "/" + strings.TrimPrefix(path, staticDir)
		if info.IsDir() {
			if err = os.MkdirAll(fullPath, os.ModePerm); err != nil {
				return fmt.Errorf("cannot create static dir %s: %w", path, err)
			}
			return nil
		}
		from, err := ThemeFs.Open(path)
		if err != nil {
			return fmt.Errorf("Could not open static file \"%s\" for copying: %w\n", path, err)

		}
		defer from.Close()

		to, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("Could not create destination static file \"%s\" for copying from virtual theme: %w\n", fullPath, err)

		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			return fmt.Errorf("Could not copy static file from virtual theme source %s to destination: %w\n", path, err)

		}

		copiedSourceCounter++
		return nil
	}); err != nil {
		return 0, fmt.Errorf("Could not get static file from virtual theme build: %w\n", err)
	}

	return copiedSourceCounter, nil
}

func copyStaticFilesFromProject(staticDir string, buildPath string, copiedSourceCounter int) (int, error) {

	if _, err := os.Stat(staticDir); err == nil {
		// The "static" folder exists, loop through contents
		err := filepath.WalkDir(staticDir, func(staticPath string, staticFileInfo fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("\ncan't stat %s: %w", staticPath, err)
			}
			destPath := buildPath + "/" + strings.TrimPrefix(staticPath, staticDir)
			if staticFileInfo.IsDir() {
				// Make directory if it doesn't exist.
				// Move on to next path.
				if err = os.MkdirAll(destPath, os.ModePerm); err != nil {
					return fmt.Errorf("\ncannot create static dir %s: %w", staticPath, err)
				}
				return nil

			}
			from, err := os.Open(staticPath)
			if err != nil {
				return fmt.Errorf("\nCould not open static file \"%s\" for copying: %w\n", staticPath, err)

			}
			defer from.Close()

			to, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("\nCould not create destination static file \"%s\" for copying: %w\n", destPath, err)

			}
			defer to.Close()

			_, err = io.Copy(to, from)
			if err != nil {
				return fmt.Errorf("\nCould not copy static file from source \"%s\" to destination: %w\n", staticPath, err)

			}

			copiedSourceCounter++
			return nil
		})
		if err != nil {
			return 0, fmt.Errorf("\nCould not get static file: %w\n", err)
		}
	}

	return copiedSourceCounter, nil
}
