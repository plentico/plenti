package build

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/plentico/plenti/common"
)

// ThemesMerge combines any nested themes with the current project.
func ThemesMerge(buildDir string) error {

	defer Benchmark(time.Now(), "Merging themes with your project")

	copiedProjectFileCounter := 0

	// Make list of files not to copy to build.
	excludedFiles := []string{
		".git",
		".gitignore",
		"themes",
		buildDir,
	}

	themeFilesErr := filepath.WalkDir(".", func(projectFilePath string, projectFileInfo fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", projectFilePath, err)
		}
		// Check if the current directory is in the excluded list.
		for _, excludedFile := range excludedFiles {
			if projectFileInfo.IsDir() && projectFileInfo.Name() == excludedFile {
				return filepath.SkipDir
			}
			if !projectFileInfo.IsDir() && projectFileInfo.Name() == excludedFile {
				return nil
			}
		}

		// Read the source project file.
		from, err := os.Open(projectFilePath)
		if err != nil {
			return fmt.Errorf("Could not open project file for copying: %w%s\n", err, common.Caller())
		}
		defer from.Close()

		// Create the folders needed to write files to tempDir.
		if projectFileInfo.IsDir() {
			// Make directory if it doesn't exist and move on to next path.
			return ThemeFs.MkdirAll(projectFilePath, os.ModePerm)

		}

		to, err := ThemeFs.Create(projectFilePath)
		if err != nil {
			return fmt.Errorf("Could not create destination project file for copying: %w%s\n", err, common.Caller())
		}
		defer to.Close()

		_, fileCopyErr := io.Copy(to, from)
		if err != nil {
			return fmt.Errorf("Could not copy project file from source to destination: %w%s\n", fileCopyErr, common.Caller())
		}

		copiedProjectFileCounter++

		return nil
	})
	if themeFilesErr != nil {
		return fmt.Errorf("Could not get project file: %w%s\n", themeFilesErr, common.Caller())
	}

	Log("Number of project files copied: " + strconv.Itoa(copiedProjectFileCounter))
	return nil

}
