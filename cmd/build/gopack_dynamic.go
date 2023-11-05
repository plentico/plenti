package build

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Create ESM support for files that don't have static imports in the app
func GopackDynamic(buildPath string, spaPath string) error {

	defer Benchmark(time.Now(), "Running GopackDynamic")

	Log("\nRunning Gopack manually on dynamic imports")

	// Dynamically imported CMS FieldWidgets
	fieldWidgetPath := "layouts/_fields"
	// Check if there are any custom FieldWidgets in the project
	if _, err := os.Stat(buildPath + "/" + fieldWidgetPath); !os.IsNotExist(err) {
		fmt.Println(buildPath + "/" + fieldWidgetPath)
		// There are custom FieldWidgets, so add ESM support for each
		err = filepath.Walk(buildPath+"/"+fieldWidgetPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("\nCan't walk ejected FieldWidget path %s: %w", path, err)
			}
			if info.IsDir() {
				return nil
			}
			err = Gopack(buildPath, spaPath, spaPath+fieldWidgetPath+"/"+path)
			if err != nil {
				return fmt.Errorf("\nError running Gopack for custom FieldWidget: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("\nCould not get custom FieldWidget: %w\n", err)
		}
	}
	return nil
}
