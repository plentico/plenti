package build

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
)

// Create global var since cmd.minifyFlag is a circular dependency.
var minifyFlag bool

// CheckMinifyFlag sets global var if --minify flag is passed.
func CheckMinifyFlag(flag bool) {
	// If --minify flag is passed by user, this will be set to true.
	minifyFlag = flag
}

func Minify(buildPath string) error {

	defer Benchmark(time.Now(), "Running Minification")

	Log("\nRunning tdewolff/minify reduce size of build assets")

	if minifyFlag {

		// Retrieve minifier
		m := minify.New()
		// Load HTML
		m.AddFunc("text/html", html.Minify)
		// Load CSS
		m.AddFunc("text/css", css.Minify)
		// Load JSON
		m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
		// Load JS
		m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)

		// Loop over built site
		err := filepath.Walk(buildPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("\nCan't walk path %s: %w", path, err)
			}
			if info.IsDir() {
				return nil
			}
			// Read the file
			contentBytes, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("\nCould not read %s for minifying: %w\n", path, err)
			}
			if filepath.Ext(path) == ".html" {
				contentBytes, err = minifyBytes(m, "text/html", contentBytes)
			}
			if filepath.Ext(path) == ".css" {
				contentBytes, err = minifyBytes(m, "text/css", contentBytes)
			}
			if filepath.Ext(path) == ".json" {
				contentBytes, err = minifyBytes(m, "application/json", contentBytes)
			}
			if filepath.Ext(path) == ".js" {
				contentBytes, err = minifyBytes(m, "text/javascript", contentBytes)
			}
			if err != nil {
				return err
			}
			// Write the file
			err = ioutil.WriteFile(path, contentBytes, 0644)
			if err != nil {
				return fmt.Errorf("Could not overwite %s with minified version: %w\n", path, err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("\nCould not minify output: %w\n", err)
		}
	}
	return nil
}

func minifyBytes(m *minify.M, mediatype string, bytes []byte) ([]byte, error) {
	bytes, err := m.Bytes(mediatype, bytes)
	if err != nil {
		return bytes, err
	}
	return bytes, nil
}
