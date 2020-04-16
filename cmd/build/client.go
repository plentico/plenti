package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Client builds the SPA.
func Client(buildPath string) {
	// Create list of all file paths in the "/layout" folder.
	var layoutFiles []string
	layoutFilesErr := filepath.Walk("layout", func(path string, info os.FileInfo, err error) error {
		layoutFiles = append(layoutFiles, path)
		return nil
	})
	if layoutFilesErr != nil {
		fmt.Printf("Could not get layout file: %s", layoutFilesErr)
	}

	for _, layoutFile := range layoutFiles {
		// Create destination path.
		destFile := buildPath + strings.Replace(layoutFile, "layout", "/spa", 1)
		// Make sure path is a directory
		fileInfo, _ := os.Stat(layoutFile)
		if fileInfo.IsDir() {
			// Create any sub directories need for filepath.
			os.MkdirAll(destFile, os.ModePerm)
		}
		excludedFiles := []string{
			"layout/ejected/build_client.js",
			"layout/ejected/build_static.js",
			"layout/ejected/server_router.js",
		}
		excluded := false
		for _, excludedFile := range excludedFiles {
			if excludedFile == layoutFile {
				excluded = true
			}
		}
		// If the file is already .js just copy it straight over to build dir.
		if filepath.Ext(layoutFile) == ".js" && !excluded {
			from, err := os.Open(layoutFile)
			if err != nil {
				fmt.Printf("Could not open source .js file for copying: %s\n", err)
			}
			defer from.Close()

			to, err := os.Create(destFile)
			if err != nil {
				fmt.Printf("Could not create destination .js file for copying: %s\n", err)
			}
			defer to.Close()

			_, fileCopyErr := io.Copy(to, from)
			if err != nil {
				fmt.Printf("Could not copy .js from source to destination: %s\n", fileCopyErr)
			}
		}
		if filepath.Ext(layoutFile) == ".svelte" {
			fileContentByte, readFileErr := ioutil.ReadFile(layoutFile)
			if readFileErr != nil {
				fmt.Printf("Could not read contents of svelte source file: %s\n", readFileErr)
			}
			fileContentStr := string(fileContentByte)
			//fmt.Printf("contents: %s\n", fileContentStr)
			output, buildErr := exec.Command("node", "layout/ejected/build_client.js", fileContentStr).Output()
			//fmt.Printf("svelte output: %s\n", output)
			if buildErr != nil {
				fmt.Printf("Could not compile svelte: %s", buildErr)
			}
			destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".js"
			err := ioutil.WriteFile(destFile, output, 0755)
			if err != nil {
				fmt.Printf("Unable to write file: %v", err)
			}
		}

	}

	fmt.Println("Running snowpack to build dependencies for ESM support")
	snowpack := exec.Command("npx", "snowpack", "--include", "'public/spa/**/*'", "--dest", "'public/spa/web_modules'")
	snowpack.Stdout = os.Stdout
	snowpack.Stderr = os.Stderr
	snowpack.Run()

}
