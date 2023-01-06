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
func EjectCopy(buildPath string, defaultsEjectedFS embed.FS) error {

	defer Benchmark(time.Now(), "Copying ejectable core files for build")

	Log("\nCopying ejectable core files to their destination:")

	copiedSourceCounter := 0

	ejected, err := fs.Sub(defaultsEjectedFS, ".")
	if err != nil {
		return fmt.Errorf("Unable to get ejected defaults: %w\n", err)
	}

	destPath := buildPath + "/spa/"
	ejectedFilesErr := fs.WalkDir(ejected, ".", func(ejectPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", ejectPath, err)
		}

		// If the file is already in .js format just copy it straight over to build dir.
		if filepath.Ext(ejectPath) == ".js" {
			if err := os.MkdirAll(destPath+"ejected", os.ModePerm); err != nil {
				return err
			}
			var ejectedContent []byte
			_, err := os.Stat(ejectPath)
			if ThemeFs != nil {
				_, err = ThemeFs.Stat(ejectPath)
			}
			// Check if file has been ejected to project or virtual theme filesystem.
			if err == nil {
				// The file has been ejected.
				// Get the file from the project or virtual theme.
				ejectedContent, err = getVirtualFileIfThemeBuild(ejectPath)
				if err != nil {
					return fmt.Errorf("can't read .js file: %s %w\n", ejectPath, err)
				}
			} else if os.IsNotExist(err) {
				// The file has not been ejected.
				// Get the file from embedded defaults.
				ejectedFile, err := ejected.Open(ejectPath)
				if err != nil {
					return fmt.Errorf("Could not open source .js file for copying: %w\n", err)
				}
				ejectedContent, err = ioutil.ReadAll(ejectedFile)
				if err != nil {
					return fmt.Errorf("Can't read ejected .js file: %w\n", err)
				}
			}
			destFile := destPath + ejectPath
			// Create any sub directories need for filepath.
			if err := os.MkdirAll(filepath.Dir(destFile), os.ModePerm); err != nil {
				return fmt.Errorf("can't make folders for '%s': %w\n", destFile, err)
			}
			if err := ioutil.WriteFile(destFile, ejectedContent, os.ModePerm); err != nil {
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
