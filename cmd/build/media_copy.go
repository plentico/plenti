package build

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

// MediaCopy does a direct copy of any media files (e.g. images, PDFs).
func MediaCopy(buildPath string) error {

	defer Benchmark(time.Now(), "Copying media files into build dir")

	Log("\nCopying media files:")

	mediaDir := "media"
	// Index of copied media files to list them in media browser
	var index []string
	// Keep track of number of files copied
	copiedSourceCounter := 0
	// Initialize error
	var err error

	if ThemeFs != nil {
		index, copiedSourceCounter, err = copyMediaFromTheme(mediaDir, buildPath, index, copiedSourceCounter)
		if err != nil {
			return err
		}
	} else {
		index, copiedSourceCounter, err = copyMediaFromProject(mediaDir, buildPath, index, copiedSourceCounter)
		if err != nil {
			return err
		}
	}

	// Create the spa/ejected/cms/media.js file
	err = createMediaIndex(buildPath, index)
	if err != nil {
		return err
	}

	Log(fmt.Sprintf("Number of media files copied: %d", copiedSourceCounter))
	return nil

}

func copyMediaFromTheme(mediaDir string, buildPath string, index []string, copiedSourceCounter int) ([]string, int, error) {

	if err := afero.Walk(ThemeFs, mediaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fullPath := buildPath + "/" + path
		if info.IsDir() {
			if err = os.MkdirAll(fullPath, os.ModePerm); err != nil {
				return fmt.Errorf("cannot create media dir %s: %w", path, err)
			}
			return nil
		}
		from, err := ThemeFs.Open(path)
		if err != nil {
			return fmt.Errorf("Could not open media file \"%s\" for copying: %w\n", path, err)

		}
		defer from.Close()

		to, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("Could not create destination media file \"%s\" for copying from virtual theme: %w\n", fullPath, err)

		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			return fmt.Errorf("Could not copy media file from virtual theme source %s to destination: %w\n", path, err)

		}

		index = append(index, path)
		copiedSourceCounter++
		return nil
	}); err != nil {
		return index, copiedSourceCounter, fmt.Errorf("Could not get media file from virtual theme build: %w\n", err)
	}

	return index, copiedSourceCounter, nil
}

func copyMediaFromProject(mediaDir string, buildPath string, index []string, copiedSourceCounter int) ([]string, int, error) {

	if _, err := os.Stat(mediaDir); err == nil {
		// the "media" folder exists, loop through contents
		err := filepath.WalkDir(mediaDir, func(mediaPath string, mediaFileInfo fs.DirEntry, err error) error {
			if err != nil {
				return fmt.Errorf("can't stat %s: %w", mediaPath, err)
			}
			destPath := buildPath + "/" + mediaPath
			if mediaFileInfo.IsDir() {
				// Make directory if it doesn't exist.
				// Move on to next path.
				if err = os.MkdirAll(destPath, os.ModePerm); err != nil {
					return fmt.Errorf("cannot create media dir %s: %w", mediaPath, err)
				}
				return nil

			}
			from, err := os.Open(mediaPath)
			if err != nil {
				return fmt.Errorf("Could not open media file \"%s\" for copying: %w\n", mediaPath, err)

			}
			defer from.Close()

			to, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("Could not create destination media file \"%s\" for copying: %w\n", destPath, err)

			}
			defer to.Close()

			_, err = io.Copy(to, from)
			if err != nil {
				return fmt.Errorf("Could not copy media file from source \"%s\" to destination: %w\n", mediaPath, err)

			}

			index = append(index, mediaPath)
			copiedSourceCounter++
			return nil
		})
		if err != nil {
			return index, copiedSourceCounter, fmt.Errorf("Could not get media file: %w\n", err)
		}
	}

	return index, copiedSourceCounter, nil
}

func createMediaIndex(buildPath string, index []string) error {
	result, err := json.MarshalIndent(index, "", "\t")
	if err != nil {
		return fmt.Errorf("Unable to marshal JSON: %w", err)
	}
	if string(result) == "null" {
		result = []byte("[]")
	}
	result = append(append([]byte("let allMedia = "), result...), []byte(";\nexport default allMedia;")...)
	err = ioutil.WriteFile(buildPath+"/spa/ejected/cms/media.js", result, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Unable to write to media index file: %w\n", err)
	}
	return nil
}
