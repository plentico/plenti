package build

import (
	"errors"
	"fmt"
	"io"
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

		tmpDir, err := os.MkdirTemp("", "*")
		if err != nil {
			return fmt.Errorf("\nCould not create temporary directory for bundle output: %w\n", err)
		}
		defer os.RemoveAll(tmpDir)

		tmpOutputFilePath := path.Join(tmpDir, "out.js")

		result := api.Build(api.BuildOptions{
			EntryPoints:    []string{mainJsFilePath},
			Bundle:         true,
			Outfile:        tmpOutputFilePath,
			Write:          true,
		})
		if len(result.Errors) != 0 {
			errs := make([]error, 0, len(result.Errors))
			for _, msg := range result.Errors {
				errs = append(errs, errors.New(msg.Text))
			}
			return fmt.Errorf("\nCould not bundle js output:\n%w\n", errors.Join(errs...))
		}

		// Not sure if this should be done, so putting in a true if block for right now
		const cleanupSpaDir = true
		if cleanupSpaDir {
			styleFilePath := path.Join(spaPath, "bundle.css")
			tmpSytleFilePath := path.Join(tmpDir, "bundle.css")
			if err := copyFile(styleFilePath, tmpSytleFilePath); err != nil {
				return fmt.Errorf("\nCould not copy bundle.css to temp file in bundle process: %w\n", err)
			}
	
			if err := os.RemoveAll(spaPath); err != nil {
				return fmt.Errorf("\nCould clear spa directory in bundle process: %w\n", err)
			}

			if err := os.MkdirAll(path.Join(spaPath, "core"), 0755); err != nil {
				return fmt.Errorf("\nCould recreate spa directory in bundle process: %w\n", err)
			}

			if err := copyFile(tmpSytleFilePath, styleFilePath); err != nil {
				return fmt.Errorf("\nCould not copy bundle.css to temp file in bundle process: %w\n", err)
			}
		}

		if err := copyFile(tmpOutputFilePath, mainJsFilePath); err != nil {
			return fmt.Errorf("\nCould not copy temp file to main.js in bundle process: %w\n", err)
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

