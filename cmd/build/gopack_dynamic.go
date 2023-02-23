package build

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/plentico/plenti/defaults"
)

// Create ESM support for files that don't have static imports in the app
func GopackDynamic(buildPath string) error {
	// Dynamically imported CMS FieldWidgets
	fieldWidgetPath := "core/cms/fields"
	fieldWidgetDefaults, err := fs.Sub(defaults.CoreFS, fieldWidgetPath)
	if err != nil {
		return fmt.Errorf("\nUnable to get core FieldWidgets: %w", err)
	}
	// Walk all defaults from embeded filesystem
	err = fs.WalkDir(fieldWidgetDefaults, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("\nCan't walk FieldWidget path %s: %w", path, err)
		}
		if d.IsDir() {
			return nil
		}
		jsPath := strings.TrimSuffix(path, ".svelte") + ".js"
		err = Gopack(buildPath, buildPath+"/spa/"+fieldWidgetPath+"/"+jsPath)
		if err != nil {
			return fmt.Errorf("\nError running Gopack for default FieldWidgets: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("\nCould not get default FieldWidget file: %w\n", err)
	}

	// Check if there are any custom FieldWidgets in the project
	if _, err := os.Stat(buildPath + "/" + fieldWidgetPath); !os.IsNotExist(err) {
		// There are custom FieldWidgets, so add ESM support for each
		err = filepath.Walk(buildPath+"/"+fieldWidgetPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("\nCan't walk ejected FieldWidget path %s: %w", path, err)
			}
			if info.IsDir() {
				return nil
			}
			jsPath := strings.TrimSuffix(path, ".svelte") + ".js"
			err = Gopack(buildPath, buildPath+"/spa/"+fieldWidgetPath+"/"+jsPath)
			if err != nil {
				return fmt.Errorf("\nError running Gopack for ejected FieldWidgets: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("\nCould not get ejected FieldWidget file: %w\n", err)
		}
	}
	return nil

}
