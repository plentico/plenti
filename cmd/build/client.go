package build

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
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
			compiledBytes, buildErr := exec.Command("node", "layout/ejected/build_client.js", fileContentStr).Output()
			if buildErr != nil {
				fmt.Printf("Could not compile svelte to JS: %s\n", buildErr)
			}
			destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".js"
			compiledStr := string(compiledBytes)
			compiledStrs := strings.Split(compiledStr, "!plenti-split!")
			//fmt.Println(compiledStrs)
			//fmt.Println(compiledStrs[0])
			jsStr := strings.TrimSpace(compiledStrs[0])
			//fmt.Println(compiledStrs[1])
			cssStr := strings.TrimSpace(compiledStrs[1])
			jsStr = strings.Replace(jsStr, ".svelte", ".js", -1)
			jsStr = strings.Replace(jsStr, "from \"svelte/internal\";", "from \"../web_modules/svelte/internal/index.js\";", -1)
			jsStr = strings.Replace(jsStr, "from \"navaid\";", "from \"../web_modules/navaid.js\";", -1)
			jsBytes := []byte(jsStr)
			err := ioutil.WriteFile(destFile, jsBytes, 0755)
			if err != nil {
				fmt.Printf("Unable to write file: %v", err)
			}
			if cssStr != "null" {
				cssFile, WriteStyleErr := os.OpenFile(buildPath+"/spa/bundle.css", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if WriteStyleErr != nil {
					fmt.Printf("Could not open bundle.css for writing: %s", WriteStyleErr)
				}
				defer cssFile.Close()
				if _, err := cssFile.WriteString(cssStr); err != nil {
					log.Println(err)
				}
			}
			/*
				compiledStyle, buildStyleErr := exec.Command("node", "layout/ejected/build_client.js", fileContentStr, "css").Output()
				compiledStyleString := string(compiledStyle)
				if buildStyleErr != nil {
					fmt.Printf("Could not compile svelte to CSS: %s\n", buildStyleErr)
				}
				cssFile, WriteStyleErr := os.OpenFile(buildPath+"/spa/bundle.css", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if WriteStyleErr != nil {
					fmt.Printf("Could not write CSS: %s", WriteStyleErr)
				}
				defer cssFile.Close()
				if _, err := cssFile.WriteString(compiledStyleString); err != nil {
					log.Println(err)
				}
			*/
		}

	}

	fmt.Println("Running snowpack to build dependencies for ESM support")
	snowpack := exec.Command("npx", "snowpack", "--include", "'public/spa/**/*'", "--dest", "'public/spa/web_modules'")
	snowpack.Stdout = os.Stdout
	snowpack.Stderr = os.Stderr
	snowpack.Run()

}
