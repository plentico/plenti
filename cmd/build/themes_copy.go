package build

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"plenti/readers"
	"strconv"
	"strings"
	"time"
)

// ThemesCopy copies nested themes into a temporary working directory.
func ThemesCopy(theme string) string {

	defer Benchmark(time.Now(), "Building themes")

	Log("Found theme named: " + theme)

	siteConfig, _ := readers.GetSiteConfig(theme)
	nestedTheme := siteConfig.Theme
	if nestedTheme != "" {
		ThemesCopy(theme + "/themes/" + nestedTheme)
	}

	// Name of temporary directory to run build inside.
	tempBuildDir := "temp_build/"

	copiedThemeFileCounter := 0

	// Make list of files not to copy to build.
	excludedFiles := []string{
		".git",
		".gitignore",
		"themes",
	}

	themeFilesErr := filepath.Walk(theme, func(themeFilePath string, themeFileInfo os.FileInfo, err error) error {

		// Check if the current directory is in the excluded list.
		for _, excluded := range excludedFiles {
			if themeFileInfo.IsDir() && themeFileInfo.Name() == excluded {
				return filepath.SkipDir
			}
			if !themeFileInfo.IsDir() && themeFileInfo.Name() == excluded {
				return nil
			}
		}

		// Read the source theme file.
		from, err := os.Open(themeFilePath)
		if err != nil {
			fmt.Printf("Could not open theme file for copying: %s\n", err)
		}
		defer from.Close()

		// Create path for the file to be written to.
		destPath := tempBuildDir + strings.TrimPrefix(themeFilePath, theme)

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

	return tempBuildDir

}
