package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plenti/common"
	"strings"
)

// Packages static files in your Go binary.
func main() {
	// Reads all files in "defaults" folder and
	// encodes them as string literals in generated/defaults.go
	common.CheckErr(generate("defaults"))
	// Reads all files in "defaults_bare" folder and
	// encodes them as string literals in generated/defaults_bare.go
	common.CheckErr(generate("defaults_bare"))
	// Reads all files in "defaults_node_modules" folder and
	// encodes them as string literals in generated/defaults_node_modules.go
	common.CheckErr(generate("defaults_node_modules"))
	// Reads all files in "ejected" folder and
	// encodes them as string literals in generated/ejected.go
	common.CheckErr(generate("ejected"))
}

func generate(name string) error {
	fp := fmt.Sprintf("generated/%s.go", name)
	out, err := os.Create(fp)
	if err != nil {
		return fmt.Errorf("couldn't create %s in 'generate': %w%s", fp, err, common.Caller())
	}
	if _, err = out.Write([]byte("package generated")); err != nil {
		return fmt.Errorf("error in 'generate':  %w%s", err, common.Caller())
	}
	if _, err = out.Write([]byte("\n\n// Do not edit, this file is automatically generated.")); err != nil {
		return fmt.Errorf("error in 'generate':  %w%s", err, common.Caller())
	}
	if _, err = out.Write([]byte(
		fmt.Sprintf("\n\n// %s: scaffolding used in 'build' command", strings.Title(name)))); err != nil {
		return fmt.Errorf("error in 'generate':  %w%s", err, common.Caller())
	}
	if _, err = out.Write([]byte(
		fmt.Sprintf(
			"\nvar %s = map[string][]byte{\n", strings.Title(name),
		))); err != nil {
		return fmt.Errorf("error in 'generate':  %w%s", err, common.Caller())
	}

	if err = filepath.Walk(name,
		func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return fmt.Errorf("can't stat %s: %w", path, err)
			}
			if !info.IsDir() {
				// Get the contents of the current file.
				content, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				// Correct filename of the .gitignore file.
				if strings.HasSuffix(path, "_plenti_replace") {
					path = strings.TrimSuffix(path, "_plenti_replace")
				}
				// Add a key for the filename to the map.
				if _, err := out.Write([]byte(
					fmt.Sprintf("\t\"%s\": []byte(`", strings.TrimPrefix(path, name))),
				); err != nil {
					return err
				}

				// Escape the backticks that would break string literals
				escapedContent := strings.Replace(string(content), "`", "`+\"`\"+`", -1)
				// Add the content as the value of the map.
				if _, err := out.Write([]byte(escapedContent)); err != nil {
					return err
				}
				// End the specific file entry in the map.
				if _, err = out.Write([]byte("`),\n")); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
		return fmt.Errorf("error walking in 'generate': %w%s", err, common.Caller())
	}
	if _, err := out.Write([]byte("}\n")); err != nil {
		return fmt.Errorf("error in 'generate': %w%s", err, common.Caller())
	}
	return nil
}
