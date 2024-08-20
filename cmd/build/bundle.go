package build

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/evanw/esbuild/pkg/api"
)

var bundleFlag bool

func CheckBundleFlag(flag bool) {
	bundleFlag = flag
}

func Bundle(spaPath string) error {

	defer Benchmark(time.Now(), "Running Bundler")

	Log("\nRunning esbuild to bundle the JS output")

	if bundleFlag {

		mainJsFilePath := path.Join(spaPath, "core/main.js")

		if err := mkCustomCmsFieldsDir(spaPath); err != nil {
			return err
		}

		result := api.Build(api.BuildOptions{
			EntryPoints:    []string{mainJsFilePath},
			Bundle:         true,
		})
		if len(result.Errors) != 0 {
			errs := make([]error, 0, len(result.Errors))
			for _, msg := range result.Errors {
				errs = append(errs, errors.New(msg.Text))
			}
			return fmt.Errorf("\nCould not bundle js output:\n%w\n", errors.Join(errs...))
		}

		outFileContent := result.OutputFiles[0].Contents

		styleFilePath := path.Join(spaPath, "bundle.css")
		styleFileContent, err := os.ReadFile(styleFilePath)
		if err != nil {
			return fmt.Errorf("\nCould not read bundle.css in bundle process: %w\n", err)
		}
	
		if err := os.RemoveAll(spaPath); err != nil {
			return fmt.Errorf("\nCould not clear spa directory in bundle process: %w\n", err)
		}

		if err := os.MkdirAll(path.Join(spaPath, "core"), 0755); err != nil {
			return fmt.Errorf("\nCould recreate spa directory in bundle process: %w\n", err)
		}

		if err := os.WriteFile(mainJsFilePath, outFileContent, os.ModePerm); err != nil {
			return fmt.Errorf("\nCould not write js to output file in bundle process: %w\n", err)
		}

		if err := os.WriteFile(styleFilePath, styleFileContent, os.ModePerm); err != nil {
			return fmt.Errorf("\nCould not write bundle.css to spa dir in bundle process: %w\n", err)
		}

	}
	return nil
}

// mkCustomCmsFieldsDir ensures 'layouts/_fields/' exists to avoid an esbuild error
// related to a dynamic import in core/cms/dynamic_form_input[.js|.svelte]
func mkCustomCmsFieldsDir(spaPath string) error {
	path := path.Join(spaPath, "layouts/_fields/")
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return fmt.Errorf("\nCould not create custom cms field directory '%s': %w\n", path, err)
	}
	return nil
}

