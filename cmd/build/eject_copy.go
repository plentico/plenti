package build

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"plenti/common"
	"strings"
	"time"
)

// EjectCopy does a direct copy of any ejectable js files needed in spa build dir.
func EjectCopy(buildPath string, tempBuildDir string, defaultsEjectedFS embed.FS) error {

	defer Benchmark(time.Now(), "Copying ejectable core files for build")

	Log("\nCopying ejectable core files to their destination:")

	copiedSourceCounter := 0

	ejected, err := fs.Sub(defaultsEjectedFS, "defaults")
	if err != nil {
		common.CheckErr(fmt.Errorf("Unable to get ejected defaults: %w", err))
	}
	ejectedFilesErr := fs.WalkDir(ejected, ".", func(ejectPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", ejectPath, err)
		}
		// Make list of files not to copy to build.
		excludedFiles := []string{
			"ejected/build.js",
		}
		// Check if the current file is in the excluded list.
		excluded := false
		for _, excludedFile := range excludedFiles {
			if excludedFile == ejectPath {
				excluded = true
			}
		}
		// If the file is already in .js format just copy it straight over to build dir.
		if filepath.Ext(ejectPath) == ".js" && !excluded {

			destPath := buildPath + "/spa/"
			if err := os.MkdirAll(destPath+strings.TrimPrefix("ejected", tempBuildDir), os.ModePerm); err != nil {
				return err
			}

			ejectedFile, err := ejected.Open(ejectPath)
			if err != nil {
				return fmt.Errorf("Could not open source .js file for copying: %w%s", err, common.Caller())
			}
			ejectedContent, err := ioutil.ReadAll(ejectedFile)
			if err != nil {
				return fmt.Errorf("Can't read ejected .js file: %w%s", err, common.Caller())
			}
			if err := ioutil.WriteFile(destPath+ejectPath, ejectedContent, os.ModePerm); err != nil {
				return fmt.Errorf("Unable to write file: %w%s", err, common.Caller())

			}
			/*
				from, err := os.Open(ejectPath)
				if err != nil {
					return fmt.Errorf("Could not open source .js file for copying: %w%s", err, common.Caller())
				}
				defer from.Close()

				to, err := os.Create(destPath + strings.TrimPrefix(ejectPath, tempBuildDir))
				if err != nil {
					return fmt.Errorf("Could not create destination .js file for copying: %w%s", err, common.Caller())
				}
				defer to.Close()

				_, err = io.Copy(to, from)
				if err != nil {
					return fmt.Errorf("Could not copy .js from source to destination: %w%s", err, common.Caller())
				}
			*/

			copiedSourceCounter++
		}
		return nil
	})
	if ejectedFilesErr != nil {
		return fmt.Errorf("Could not get ejectable file: %w%s", ejectedFilesErr, common.Caller())
	}

	Log(fmt.Sprintf("Number of ejectable core files copied: %d\n", copiedSourceCounter))
	return nil

}
