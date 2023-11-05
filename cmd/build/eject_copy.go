package build

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// EjectCopy does a direct copy of any ejectable js files needed in spa build dir.
func EjectCopy(destPath string, coreFS embed.FS) error {

	defer Benchmark(time.Now(), "Copying ejectable core files for build")

	Log("\nCopying ejectable core files to their destination:")

	copiedSourceCounter := 0

	coreDefaults, err := fs.Sub(coreFS, ".")
	if err != nil {
		return fmt.Errorf("Unable to get ejected defaults: %w\n", err)
	}

	ejectedFilesErr := fs.WalkDir(coreDefaults, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", path, err)
		}

		// If the file is already in .js format just copy it straight over to build dir.
		if filepath.Ext(path) == ".js" {
			if err := os.MkdirAll(destPath+"core", os.ModePerm); err != nil {
				return err
			}
			var destContent []byte
			// Set error if path doesn't exist in project filesystem
			_, err := os.Stat(path)
			// Check if theme is being used
			if ThemeFs != nil {
				// Set error if path doesn't exist in virtual theme filesystem
				_, err = ThemeFs.Stat(path)
			}
			// Check if file has been ejected to project or virtual theme filesystem.
			if err == nil {
				// No stat errors, the file has been ejected.
				// Get the file from the project or virtual theme.
				destContent, err = getVirtualFileIfThemeBuild(path)
				if err != nil {
					return fmt.Errorf("can't read .js file: %s %w\n", path, err)
				}
			} else if os.IsNotExist(err) {
				// The file has not been ejected.
				// Get the file from embedded defaults.
				coreFile, err := coreDefaults.Open(path)
				if err != nil {
					return fmt.Errorf("Could not open source .js file for copying: %w\n", err)
				}
				destContent, err = ioutil.ReadAll(coreFile)
				if err != nil {
					return fmt.Errorf("Can't read ejected .js file: %w\n", err)
				}
			}
			destFile := destPath + path
			// Create any sub directories need for filepath.
			if err := os.MkdirAll(filepath.Dir(destFile), os.ModePerm); err != nil {
				return fmt.Errorf("can't make folders for '%s': %w\n", destFile, err)
			}
			// Write file to public build directory
			if err := ioutil.WriteFile(destFile, destContent, os.ModePerm); err != nil {
				return fmt.Errorf("Unable to write file: %w\n", err)
			}

			copiedSourceCounter++
		}
		return nil
	})
	if ejectedFilesErr != nil {
		return fmt.Errorf("Could not get ejectable file: %w\n", ejectedFilesErr)
	}

	Log(fmt.Sprintf("Number of ejectable core files copied: %d", copiedSourceCounter))
	return nil

}
