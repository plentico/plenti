package build

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"plenti/readers"
	"regexp"
	"strings"
	"time"
)

// NodeDataSource gathers data json from "content/" directory to use in NodeJS build (NOTE: This is legacy).
func NodeDataSource(buildPath string, siteConfig readers.SiteConfig) (string, string, error) {

	defer Benchmark(time.Now(), "Creating data_source")

	Log("\nGathering data source from 'content/' folder")

	contentJSPath := buildPath + "/spa/ejected/content.js"
	if err := os.MkdirAll(buildPath+"/spa/ejected", os.ModePerm); err != nil {
		return "", "", err
	}

	// Set up counter for logging output.
	contentFileCounter := 0

	// Start the string that will be sent to nodejs for compiling.
	staticBuildStr := "["
	allContentStr := "["

	// Start the new content.js file.
	err := ioutil.WriteFile(contentJSPath, []byte(`const contentSource = [`), 0755)
	if err != nil {
		fmt.Printf("Unable to write content.js file: %v", err)
	}

	// Go through all sub directories in "content/" folder.
	contentFilesErr := filepath.Walk("content", func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return fmt.Errorf("can't stat %s: %w", path, err)
		}
		if !info.IsDir() {
			// Get individual path arguments.
			parts := strings.Split(path, "/")
			contentType := parts[1]
			fileName := parts[len(parts)-1]

			// Don't add _blueprint.json or other special named files starting with underscores.
			if fileName[:1] != "_" && fileName[:1] != "." {

				// Get the contents of the file.
				fileContentBytes, err := ioutil.ReadFile(path)
				if err != nil {
					fmt.Printf("Could not read content file: %s\n", err)
					return err
				}
				fileContentStr := string(fileContentBytes)

				// Remove the "content" folder from path.
				path = strings.TrimPrefix(path, "content")

				// Check for index file at any level.
				if fileName == "index.json" {
					// Remove entire filename from path.
					path = strings.TrimSuffix(path, fileName)
					// Remove trailing slash, unless it's the homepage.
					if path != "/" && path[len(path)-1:] == "/" {
						path = strings.TrimSuffix(path, "/")
					}
				} else {
					// Remove file extension only from path for files other than index.json.
					path = strings.TrimSuffix(path, filepath.Ext(path))
				}

				// Get field key/values from content source.
				typeFields := readers.GetTypeFields(fileContentBytes)
				// Setup regex to find field name.
				reField := regexp.MustCompile(`:field\((.*?)\)`)
				// Create regex for allowed characters when slugifying path.
				reSlugify := regexp.MustCompile("[^a-z0-9/]+")

				// Check for path overrides from plenti.json config file.
				for configContentType, slug := range siteConfig.Types {
					if configContentType == contentType {
						// Replace :filename.
						slug = strings.Replace(slug, ":filename", strings.TrimSuffix(fileName, filepath.Ext(fileName)), -1)

						// Replace :field().
						fieldReplacements := reField.FindAllStringSubmatch(slug, -1)
						// Loop through all :field() replacements found in config file.
						for _, replacement := range fieldReplacements {
							// Loop through all top level keys found in content source file.
							for field, fieldValue := range typeFields.Fields {
								// Check if field name in the replacement pattern is found in data source.
								if replacement[1] == field {
									// Use the field value in the path.
									slug = strings.ReplaceAll(slug, replacement[0], fieldValue)
								}
							}

						}

						// Slugify output using reSlugify regex defined above.
						slug = strings.Trim(reSlugify.ReplaceAllString(strings.ToLower(slug), "-"), "-")
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

				contentDetailsStr := "{\n" +
					"\"path\": \"" + path + "\",\n" +
					"\"type\": \"" + contentType + "\",\n" +
					"\"filename\": \"" + fileName + "\",\n" +
					"\"fields\": " + fileContentStr + "\n}"

				// Create new content.js file if it doesn't already exist, or add to it if it does.
				contentJSFile, err := os.OpenFile(contentJSPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Printf("Could not open content.js for writing: %s", err)
				}
				// Write to the file with info from current file in "/content" folder.
				defer contentJSFile.Close()
				if _, err := contentJSFile.WriteString(contentDetailsStr + ","); err != nil {
					return err
				}

				// Need to encode html so it can be send as string to NodeJS in exec.Command.
				encodedContentDetails := contentDetailsStr
				// Remove newlines.
				reN := regexp.MustCompile(`\r?\n`)
				encodedContentDetails = reN.ReplaceAllString(encodedContentDetails, " ")
				// Remove tabs.
				reT := regexp.MustCompile(`\t`)
				encodedContentDetails = reT.ReplaceAllString(encodedContentDetails, " ")
				// Reduce extra whitespace to a single space.
				reS := regexp.MustCompile(`\s+`)
				encodedContentDetails = reS.ReplaceAllString(encodedContentDetails, " ")

				// Add node info for being referenced in allContent object.
				allContentStr = allContentStr + encodedContentDetails + ","

				// Create path for source .svelte template.
				componentPath := "layout/content/" + contentType + ".svelte"
				// Do not add a content source without a corresponding template to the build string.
				if _, noEndpointErr := os.Stat(componentPath); os.IsNotExist(noEndpointErr) {
					// The componentPath does not exist, go to the next content source.
					// this is/should be an error?
					return nil
				}
				// Add to list of data_source files for creating static HTML.
				staticBuildStr = staticBuildStr + "{ \"content\": " + encodedContentDetails + ", \"componentPath\": \"" + componentPath + "\", \"destPath\": \"" + destPath + "\"},"

				// Increment counter for logging purposes.
				contentFileCounter++

			}
		}
		return nil
	})
	if contentFilesErr != nil {
		fmt.Printf("Could not get layout file: %s", contentFilesErr)
	}

	// Complete the content.js file.
	contentJSFile, err := os.OpenFile(contentJSPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Could not open content.js for writing: %s", err)
		return "", "", err
	}
	defer contentJSFile.Close()
	if _, err := contentJSFile.WriteString("];\n\nexport default contentSource;"); err != nil {
		return "", "", err
	}

	// End the string that will be sent to nodejs for compiling.
	staticBuildStr = strings.TrimSuffix(staticBuildStr, ",") + "]"
	allContentStr = strings.TrimSuffix(allContentStr, ",") + "]"

	Log(fmt.Sprintf("Number of content files used: %d", contentFileCounter))

	return staticBuildStr, allContentStr, nil

}
