package build

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/afero"
)

// themeTmpDir holds the location of the materialized theme filesystem.
var themeTmpDir string

// MaterializeThemeFs writes the virtual theme filesystem to a temp dir on
// disk so tools that read from the real filesystem (e.g. Pico) can use it.
// Returns the temp directory path.
func MaterializeThemeFs() (string, error) {
	if themeTmpDir != "" {
		return themeTmpDir, nil
	}
	if ThemeFs == nil {
		return "", fmt.Errorf("theme filesystem is not initialized")
	}
	tmpDir, err := os.MkdirTemp("", "plenti-theme")
	if err != nil {
		return "", fmt.Errorf("can't create temp dir for theme build: %w", err)
	}
	err = afero.Walk(ThemeFs, ".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		dest := filepath.Join(tmpDir, path)
		if info.IsDir() {
			return os.MkdirAll(dest, os.ModePerm)
		}
		content, err := afero.ReadFile(ThemeFs, path)
		if err != nil {
			return fmt.Errorf("can't read theme file %s: %w", path, err)
		}
		if err := os.WriteFile(dest, content, 0644); err != nil {
			return fmt.Errorf("can't write theme file %s: %w", dest, err)
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("can't materialize theme filesystem: %w", err)
	}
	themeTmpDir = tmpDir
	return tmpDir, nil
}

// ThemesClean removes the materialized theme temp dir after the build.
func ThemesClean() {
	defer Benchmark(time.Now(), "Clean up theme temp dir")
	if themeTmpDir != "" {
		Log("Removing temporary theme build directory")
		os.RemoveAll(themeTmpDir)
		themeTmpDir = ""
	}
}
