package build

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"plenti/readers"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DataSource builds json list from "content/" directory.
func DataSource(buildPath string, siteConfig readers.SiteConfig) (string, string) {

	defer Benchmark(time.Now(), "Creating data_source")

	Log("\nGathering data source from 'content/' folder")

	nodesJSPath := buildPath + "/spa/ejected/nodes.js"
	os.MkdirAll(buildPath+"/spa/ejected", os.ModePerm)

	// Set up counter for logging output.
	contentFileCounter := 0

	// Start the string that will be sent to nodejs for compiling.
	staticBuildStr := "["
	allNodesStr := "["

	// Start the new nodes.js file.
	err := ioutil.WriteFile(nodesJSPath, []byte(`const nodes = [`), 0755)
	if err != nil {
		fmt.Printf("Unable to write nodes.js file: %v", err)
	}

	// Go through all sub directories in "content/" folder.
	contentFilesErr := filepath.Walk("content", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			// Get individual path arguments.
			parts := strings.Split(path, "/")
			contentType := parts[1]
			fileName := parts[len(parts)-1]

			// Don't add _blueprint.json or other special named files starting with underscores.
			if fileName[:1] != "_" {

				// Get the contents of the file.
				fileContentBytes, readFileErr := ioutil.ReadFile(path)
				if readFileErr != nil {
					fmt.Printf("Could not read content file: %s\n", readFileErr)
				}
				fileContentStr := string(fileContentBytes)

				// Check for index file at any level.
				if fileName == "index.json" {
					// Remove entire filename from path.
					path = strings.TrimSuffix(path, fileName)
				} else {
					// Remove file extension only from path for files other than index.json.
					path = strings.TrimSuffix(path, filepath.Ext(path))
				}
				// Remove the "content" folder from path.
				path = strings.TrimPrefix(path, "content")

				// Get field key/values from content source.
				typeFields := readers.GetTypeFields(fileContentBytes)
				// Setup regex to find field name.
				reField := regexp.MustCompile(`:field\((.*)\)`)

				// Check for path overrides from plenti.json config file.
				for configContentType, slug := range siteConfig.Types {
					if configContentType == contentType {
						// Replace :filename.
						slug = strings.Replace(slug, ":filename", strings.TrimSuffix(fileName, filepath.Ext(fileName)), -1)
						//slug = strings.TrimSuffix(slug, filepath.Ext(file))
						fmt.Printf("slug is: %s", slug)

						// Replace :field().
						//fieldReplacements := reField.FindAll([]byte(slug), -1)
						//fieldReplacements := reField.FindStringSubmatch(slug)
						fieldReplacements := reField.FindAllStringSubmatch(slug, -1)
						for _, replacement := range fieldReplacements {
							fmt.Printf("Replacement: %s", replacement[1])
							for field, fieldValue := range typeFields.Fields {
								if replacement[1] == field {
									slug = reField.ReplaceAllString(slug, fieldValue)
									//fmt.Printf(fieldValue)
									//fmt.Printf(field)
								}
							}

						}

						// Slugify output.
						slug = strings.Replace(slug, "_", "-", -1)
						slug = strings.Replace(slug, " ", "-", -1)
						slug = strings.Replace(slug, ".", "", -1)
						slug = strings.ToLower(slug)
						path = slug
					}
				}

				// Add trailing slash.
				//path = path + "/"

				// Check for files outside of a type declaration.
				if len(parts) == 2 {
					// Remove the extension since the filename = the type name.
					contentType = strings.TrimSuffix(contentType, filepath.Ext(contentType))
				}

				destPath := buildPath + "/" + path + "/index.html"

				nodeDetailsStr := "{\n" +
					"\"path\": \"" + path + "\",\n" +
					"\"type\": \"" + contentType + "\",\n" +
					"\"filename\": \"" + fileName + "\",\n" +
					"\"fields\": " + fileContentStr + "\n}"

				// Create new nodes.js file if it doesn't already exist, or add to it if it does.
				nodesJSFile, openNodesJSErr := os.OpenFile(nodesJSPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if openNodesJSErr != nil {
					fmt.Printf("Could not open nodes.js for writing: %s", openNodesJSErr)
				}
				// Write to the file with info from current file in "/content" folder.
				defer nodesJSFile.Close()
				if _, err := nodesJSFile.WriteString(nodeDetailsStr + ","); err != nil {
					log.Println(err)
				}

				// Need to encode html so it can be send as string to NodeJS in exec.Command.
				encodedNodeDetails := nodeDetailsStr
				// Remove newlines.
				reN := regexp.MustCompile(`\r?\n`)
				encodedNodeDetails = reN.ReplaceAllString(encodedNodeDetails, " ")
				// Remove tabs.
				reT := regexp.MustCompile(`\t`)
				encodedNodeDetails = reT.ReplaceAllString(encodedNodeDetails, " ")
				// Reduce extra whitespace to a single space.
				reS := regexp.MustCompile(`\s+`)
				encodedNodeDetails = reS.ReplaceAllString(encodedNodeDetails, " ")

				// Add node info for being referenced in allNodes object.
				allNodesStr = allNodesStr + encodedNodeDetails + ","

				// Create path for source .svelte template.
				componentPath := "layout/content/" + contentType + ".svelte"
				// Do not add a content source without a corresponding template to the build string.
				if _, noEndpointErr := os.Stat(componentPath); os.IsNotExist(noEndpointErr) {
					// The componentPath does not exist, go to the next content source.
					return nil
				}
				// Add to list of data_source files for creating static HTML.
				staticBuildStr = staticBuildStr + "{ \"node\": " + encodedNodeDetails + ", \"componentPath\": \"" + componentPath + "\", \"destPath\": \"" + destPath + "\"},"

				// Increment counter for logging purposes.
				contentFileCounter++

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
	nodesJSStr := "];\n\nexport default nodes;"
	if _, err := nodesJSFile.WriteString(nodesJSStr); err != nil {
		log.Println(err)
	}

	// End the string that will be sent to nodejs for compiling.
	staticBuildStr = strings.TrimSuffix(staticBuildStr, ",") + "]"
	allNodesStr = strings.TrimSuffix(allNodesStr, ",") + "]"

	Log("Number of content files used: " + strconv.Itoa(contentFileCounter))

	return staticBuildStr, allNodesStr

}
