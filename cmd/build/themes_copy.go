package build

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/plentico/plenti/common"
	"github.com/plentico/plenti/readers"
	"github.com/spf13/afero"
)

// Virtual filesystem for doing theme builds without writing to disk.
var AppFs = afero.NewMemMapFs()

// ThemesCopy copies nested themes into a temporary, virtual working directory.
func ThemesCopy(theme string, themeOptions readers.ThemeOptions) error {

	defer Benchmark(time.Now(), "Building themes")

	Log("Found theme named: " + theme)

	siteConfig, _ := readers.GetSiteConfig(theme)
	nestedTheme := siteConfig.Theme
	if nestedTheme != "" {
		// Look for options (like excluded folders) in theme.
		nestedThemeOptions := siteConfig.ThemeConfig[nestedTheme]
		// Recursively run merge on nested theme.
		err := ThemesCopy(theme+"/themes/"+nestedTheme, nestedThemeOptions)
		if err != nil {
			return err
		}
	}

	copiedThemeFileCounter := 0

	// Make list of files not to copy to build.
	excludedFiles := []string{
		".git",
		".gitignore",
		"themes",
	}

	// Merge any user specified exclusions.
	excludedFiles = append(excludedFiles, themeOptions.Exclude...)

	themeFilesErr := filepath.WalkDir(theme, func(themeFilePath string, themeFileInfo fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", themeFilePath, err)
		}

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
			return fmt.Errorf("Could not open theme file for copying: %w%s\n", err, common.Caller())
		}
		defer from.Close()

		// Create path for the file to be written to.
		destPath := strings.TrimPrefix(themeFilePath, theme+"/")

		// Create the folders needed to write files to tempDir.
		if themeFileInfo.IsDir() {
			// Make directory if it doesn't exist.
			// Move on to next path.
			return AppFs.MkdirAll(destPath, os.ModePerm)
		}

		to, err := AppFs.Create(destPath)
		if err != nil {
			return fmt.Errorf("Could not create destination theme file for copying: %w%s\n", err, common.Caller())
		}
		defer to.Close()

		_, fileCopyErr := io.Copy(to, from)
		if err != nil {
			return fmt.Errorf("Could not copy theme file from source to destination: %w%s\n", fileCopyErr, common.Caller())
		}

		copiedThemeFileCounter++

		return nil
	})
	if themeFilesErr != nil {
		return fmt.Errorf("Could not get theme file: %w%s\n", themeFilesErr, common.Caller())
	}

	Log("Number of theme files copied: " + strconv.Itoa(copiedThemeFileCounter))

	return nil

}
