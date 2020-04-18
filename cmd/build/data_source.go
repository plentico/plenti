package build

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// DataSource builds json list from "content/" directory.
func DataSource(buildPath string) []string {

	nodesJSPath := buildPath + "/spa/ejected/nodes.js"
	os.MkdirAll(buildPath+"/spa/ejected", os.ModePerm)
	// Delete any previous nodes.js file.
	deleteNodesJSErr := os.Remove(nodesJSPath)
	if deleteNodesJSErr != nil {
		fmt.Println(deleteNodesJSErr)
	}

	// Start the new nodes.js file.
	err := ioutil.WriteFile(nodesJSPath, []byte(`const nodes = [`+"\n"), 0755)
	if err != nil {
		fmt.Printf("Unable to write nodes.js file: %v", err)
	}

	var contentFiles []string
	// Go through all sub directories in "content/" folder.
	contentFilesErr := filepath.Walk("content", func(path string, info os.FileInfo, err error) error {
		//contentFiles = append(contentFiles, path)
		if !info.IsDir() {
			// Get individual path arguments.
			parts := strings.Split(path, "/")
			contentType := parts[1]
			fileName := parts[len(parts)-1]

			// Don't add _blueprint.json or other special named files starting with underscores.
			if fileName[:1] != "_" {

				// Get the contents of the file.
				fileContentByte, readFileErr := ioutil.ReadFile(path)
				if readFileErr != nil {
					fmt.Printf("Could not read content file: %s\n", readFileErr)
				}
				fileContentStr := string(fileContentByte)

				// Check for index.json outside of type declaration.
				if contentType == "index.json" {
					contentType = "index"
					path = "/"
				}
				// TODO: Need to check for path overrides from siteConfig reader.
				contents := []byte(`{
	"path": "` + strings.TrimSuffix(path, filepath.Ext(path)) + `",
	"type": "` + contentType + `",
	"filename": "` + fileName + `",
	"fields": ` + fileContentStr + "\n},\n")
				nodesJSFile, openNodesJSErr := os.OpenFile(nodesJSPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if openNodesJSErr != nil {
					fmt.Printf("Could not open nodes.js for writing: %s", openNodesJSErr)
				}
				defer nodesJSFile.Close()
				contentsStr := string(contents)
				if _, err := nodesJSFile.WriteString(contentsStr); err != nil {
					log.Println(err)
				}
			}
		}
		return nil
	})
	if contentFilesErr != nil {
		fmt.Printf("Could not get layout file: %s", contentFilesErr)
	}

	// Complete the nodes.js file.
	nodesJSFile, openNodesJSErr := os.OpenFile(nodesJSPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openNodesJSErr != nil {
		fmt.Printf("Could not open nodes.js for writing: %s", openNodesJSErr)
	}
	defer nodesJSFile.Close()
	nodesJSStr := "\n];\n\nexport default nodes;"
	if _, err := nodesJSFile.WriteString(nodesJSStr); err != nil {
		log.Println(err)
	}

	return contentFiles

}
