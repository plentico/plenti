package build

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/plentico/plenti/common"
)

// EjectCopy does a direct copy of any ejectable js files needed in spa build dir.
func EjectCopy(buildPath string, defaultsEjectedFS embed.FS) error {

	defer Benchmark(time.Now(), "Copying ejectable core files for build")

	Log("\nCopying ejectable core files to their destination:")

	copiedSourceCounter := 0

	ejected, err := fs.Sub(defaultsEjectedFS, "defaults")
	if err != nil {
		return fmt.Errorf("Unable to get ejected defaults: %w%s\n", err, common.Caller())
	}

	destPath := buildPath + "/spa/"
	ejectedFilesErr := fs.WalkDir(ejected, ".", func(ejectPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", ejectPath, err)
		}

		// If the file is already in .js format just copy it straight over to build dir.
		if filepath.Ext(ejectPath) == ".js" {

			if common.UseMemFS {

				bd, err := fs.ReadFile(ejected, ejectPath)
				if err != nil {
					return fmt.Errorf("can't read fs file: %s %w%s\n", ejectPath, err, common.Caller())
				}

				common.Set(destPath+ejectPath, ejectPath, &common.FData{B: bd})
				return nil

			}

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
					return fmt.Errorf("can't read .js file: %s %w%s\n", ejectPath, err, common.Caller())
				}
			} else if os.IsNotExist(err) {
				// The file has not been ejected.
				// Get the file from embedded defaults.
				ejectedFile, err := ejected.Open(ejectPath)
				if err != nil {
					return fmt.Errorf("Could not open source .js file for copying: %w%s\n", err, common.Caller())
				}
				ejectedContent, err = ioutil.ReadAll(ejectedFile)
				if err != nil {
					return fmt.Errorf("Can't read ejected .js file: %w%s\n", err, common.Caller())
				}
			}
			destFile := destPath + ejectPath
			// Create any sub directories need for filepath.
			if err := os.MkdirAll(filepath.Dir(destFile), os.ModePerm); err != nil {
				return fmt.Errorf("can't make folders for '%s': %w%s\n", destFile, err, common.Caller())
			}
			if err := ioutil.WriteFile(destFile, ejectedContent, os.ModePerm); err != nil {
				return fmt.Errorf("Unable to write file: %w%s\n", err, common.Caller())
			}

			copiedSourceCounter++
		}
		return nil
	})
	if ejectedFilesErr != nil {
		return fmt.Errorf("Could not get ejectable file: %w%s\n", ejectedFilesErr, common.Caller())
	}

	Log(fmt.Sprintf("Number of ejectable core files copied: %d", copiedSourceCounter))
	return nil

}
