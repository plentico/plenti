package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ThemesMerge combines any nested themes with the current project.
func ThemesMerge(tempBuildDir string, buildDir string) error {

	defer Benchmark(time.Now(), "Merging themes with your project")

	copiedProjectFileCounter := 0

	// Make list of files not to copy to build.
	excludedFiles := []string{
		".git",
		".gitignore",
		"themes",
		strings.TrimSuffix(tempBuildDir, "/"),
		buildDir,
	}

	themeFilesErr := filepath.Walk(".", func(projectFilePath string, projectFileInfo os.FileInfo, err error) error {

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
			return fmt.Errorf("Could not open project file for copying: %w", err)
		}
		defer from.Close()

		// Create path for the file to be written to.
		destPath := tempBuildDir + projectFilePath

		// Create the folders needed to write files to tempDir.
		if projectFileInfo.IsDir() {
			// Make directory if it doesn't exist.
			return os.MkdirAll(destPath, os.ModePerm)

		}

		to, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("Could not create destination project file for copying: %w", err)
		}
		defer to.Close()

		_, fileCopyErr := io.Copy(to, from)
		if err != nil {
			return fmt.Errorf("Could not copy project file from source to destination: %w", fileCopyErr)
		}

		copiedProjectFileCounter++

		return nil
	})
	if themeFilesErr != nil {
		return fmt.Errorf("Could not get project file: %w", themeFilesErr)
	}

	Log("Number of project files copied: " + strconv.Itoa(copiedProjectFileCounter))
	return nil

}
