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

// ThemesCopy combines any nested themes with the current project.
func ThemesCopy(theme string) {

	defer Benchmark(time.Now(), "Building themes")

	// Name of temporary directory to run build inside.
	tempDir := "temp_build/"

	// Get the directory of the current theme.
	themeDir := "themes/" + theme

	copiedThemeFileCounter := 0

	themeFilesErr := filepath.Walk(themeDir, func(themeFilePath string, themeFileInfo os.FileInfo, err error) error {

		// Read the source theme file.
		from, err := os.Open(themeFilePath)
		if err != nil {
			fmt.Printf("Could not open theme file for copying: %s\n", err)
		}
		defer from.Close()

		// Create path for the file to be written to.
		destPath := tempDir + strings.TrimLeft(themeFilePath, themeDir)

		// Create the folders needed to write files to tempDir.
		if themeFileInfo.IsDir() {
			// Make directory if it doesn't exist.
			os.MkdirAll(destPath, os.ModePerm)
			// Move on to next path.
			return nil
		}

		to, err := os.Create(destPath)
		if err != nil {
			fmt.Printf("Could not create destination theme file for copying: %s\n", err)
		}
		defer to.Close()

		_, fileCopyErr := io.Copy(to, from)
		if err != nil {
			fmt.Printf("Could not copy theme file from source to destination: %s\n", fileCopyErr)
		}

		copiedThemeFileCounter++

		return nil
	})
	if themeFilesErr != nil {
		fmt.Printf("Could not get theme file: %s", themeFilesErr)
	}

	Log("Number of theme files copied: " + strconv.Itoa(copiedThemeFileCounter))

}
