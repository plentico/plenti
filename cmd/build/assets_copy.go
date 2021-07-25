package build

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/plentico/plenti/common"
	"github.com/spf13/afero"
)

// AssetsCopy does a direct copy of any static assets.
//func AssetsCopy(buildPath string, themeBuildDir string) error {
func AssetsCopy(buildPath string) error {

	defer Benchmark(time.Now(), "Copying static assets into build dir")

	Log("\nCopying static assets:")

	assetsDir := "assets"
	copiedSourceCounter := 0
	var err error

	//if themeBuildDir != "" {
	if AppFs != nil {
		//copiedSourceCounter, err = copyAssetsFromTheme(assetsDir, themeBuildDir, buildPath, copiedSourceCounter)
		copiedSourceCounter, err = copyAssetsFromTheme(assetsDir, buildPath, copiedSourceCounter)
		if err != nil {
			return err
		}
	} else {
		copiedSourceCounter, err = copyAssetsFromProject(assetsDir, buildPath, copiedSourceCounter)
		if err != nil {
			return err
		}
	}

	Log(fmt.Sprintf("Number of assets copied: %d", copiedSourceCounter))
	return nil

}

//func copyAssetsFromTheme(assetsDir string, themeBuildDir string, buildPath string, copiedSourceCounter int) (int, error) {
func copyAssetsFromTheme(assetsDir string, buildPath string, copiedSourceCounter int) (int, error) {
	//themeAssets := themeBuildDir + assetsDir
	//if err := afero.Walk(AppFs, themeAssets, func(path string, info os.FileInfo, err error) error {
	if err := afero.Walk(AppFs, assetsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fullPath := buildPath + "/" + path
		fmt.Println(fullPath)
		if info.IsDir() {
			if err = os.MkdirAll(fullPath, os.ModePerm); err != nil {
				return fmt.Errorf("cannot create asset dir %s: %w", path, err)
			}
			return nil
		}
		from, err := AppFs.Open(path)
		if err != nil {
			return fmt.Errorf("Could not open asset %s for copying: %w%s\n", path, err, common.Caller())

		}
		defer from.Close()

		//destPath := buildPath + "/assets"
		//destPath := path
		to, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("Could not create destination asset %s for copying from virtual theme: %w%s\n", fullPath, err, common.Caller())

		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			return fmt.Errorf("Could not copy asset from virtual theme source %s to destination: %w%s\n", path, err, common.Caller())

		}

		copiedSourceCounter++
		return nil
	}); err != nil {
		return 0, fmt.Errorf("Could not get asset file from virtual theme build: %w%s\n", err, common.Caller())
	}
	return copiedSourceCounter, nil
}

func copyAssetsFromProject(assetsDir string, buildPath string, copiedSourceCounter int) (int, error) {

	// Exit function if "assets/" directory does not exist.
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		return 0, fmt.Errorf("%s does not exist: %w", assetsDir, err)
	}

	err := filepath.WalkDir(assetsDir, func(assetPath string, assetFileInfo fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", assetPath, err)
		}
		//destPath := buildPath + "/" + strings.TrimPrefix(assetPath, themeBuildDir)
		destPath := buildPath + "/" + assetPath
		if assetFileInfo.IsDir() {
			// no dirs
			if common.UseMemFS {
				return nil
			}
			// Make directory if it doesn't exist.
			// Move on to next path.
			if err = os.MkdirAll(destPath, os.ModePerm); err != nil {
				return fmt.Errorf("cannot create asset dir %s: %w", assetPath, err)
			}
			return nil

		}
		if common.UseMemFS {
			bd, err := os.ReadFile(assetPath)
			if err != nil {
				return fmt.Errorf("Could not open asset %s for copying: %w%s\n", assetPath, err, common.Caller())
			}
			common.Set(destPath, assetPath, &common.FData{B: bd})
			return nil
		}
		from, err := os.Open(assetPath)
		if err != nil {
			return fmt.Errorf("Could not open asset %s for copying: %w%s\n", assetPath, err, common.Caller())

		}
		defer from.Close()

		to, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("Could not create destination asset %s for copying: %w%s\n", destPath, err, common.Caller())

		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			return fmt.Errorf("Could not copy asset from source %s to destination: %w%s\n", assetPath, err, common.Caller())

		}

		copiedSourceCounter++
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("Could not get asset file: %w%s\n", err, common.Caller())
	}
	return copiedSourceCounter, nil
}
