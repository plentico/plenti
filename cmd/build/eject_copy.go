package build

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/plentico/plenti/common"
)

// EjectCopy does a direct copy of any ejectable js files needed in spa build dir.
func EjectCopy(buildPath string, tempBuildDir string, defaultsEjectedFS embed.FS) error {

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

			if err := os.MkdirAll(destPath+strings.TrimPrefix("ejected", tempBuildDir), os.ModePerm); err != nil {
				return err
			}
			var ejectedContent []byte
			if _, err := os.Stat(ejectPath); err == nil {
				ejectedContent, err = ioutil.ReadFile(ejectPath)
				if err != nil {
					return fmt.Errorf("can't read .js file: %s %w%s\n", ejectPath, err, common.Caller())
				}
			} else if os.IsNotExist(err) {
				ejectedFile, err := ejected.Open(ejectPath)
				if err != nil {
					return fmt.Errorf("Could not open source .js file for copying: %w%s\n", err, common.Caller())
				}
				ejectedContent, err = ioutil.ReadAll(ejectedFile)
				if err != nil {
					return fmt.Errorf("Can't read ejected .js file: %w%s\n", err, common.Caller())
				}
			}
			if err := ioutil.WriteFile(destPath+ejectPath, ejectedContent, os.ModePerm); err != nil {
				return fmt.Errorf("Unable to write file: %w%s\n", err, common.Caller())
			}

			copiedSourceCounter++
		}
		return nil
	})
	if ejectedFilesErr != nil {
		return fmt.Errorf("Could not get ejectable file: %w%s\n", ejectedFilesErr, common.Caller())
	}

	Log(fmt.Sprintf("Number of ejectable core files copied: %d\n", copiedSourceCounter))
	return nil

}
