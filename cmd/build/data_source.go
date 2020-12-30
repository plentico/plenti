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

type content struct {
	contentType     string
	contentPath     string
	contentDest     string
	contentDetails  string
	contentFilename string
	contentFields   string
}

// DataSource builds json list from "content/" directory.
func DataSource(buildPath string, siteConfig readers.SiteConfig, tempBuildDir string) {

	defer Benchmark(time.Now(), "Creating data_source")

	Log("\nGathering data source from 'content/' folder")

	contentJSPath := buildPath + "/spa/ejected/content.js"
	os.MkdirAll(buildPath+"/spa/ejected", os.ModePerm)

	// Set up counter for logging output.
	contentFileCounter := 0
	// Start the string that will be used for allContent object.
	allContentStr := "["
	// Store each content file in array we can iterate over for creating static html.
	allContent := []content{}

	// Start the new content.js file.
	err := ioutil.WriteFile(contentJSPath, []byte(`const contentSource = [`), 0755)
	if err != nil {
		fmt.Printf("Unable to write content.js file: %v", err)
	}

	// Go through all sub directories in "content/" folder.
	contentFilesErr := filepath.Walk(tempBuildDir+"content", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			// Get individual path arguments.
			parts := strings.Split(path, "/")
			contentType := parts[1]
			if tempBuildDir != "" {
				contentType = parts[2]
			}
			fileName := parts[len(parts)-1]

			// Don't add _blueprint.json or other special named files starting with underscores.
			if fileName[:1] != "_" {

				// Get the contents of the file.
				fileContentBytes, readFileErr := ioutil.ReadFile(path)
				if readFileErr != nil {
					fmt.Printf("Could not read content file: %s\n", readFileErr)
				}
				fileContentStr := string(fileContentBytes)

				// Remove the "content" folder from path.
				path = strings.TrimPrefix(path, tempBuildDir+"content")

				// Check for index file at any level.
				if fileName == "index.json" {
					// Remove entire filename from path.
					path = strings.TrimSuffix(path, fileName)
				} else {
					// Remove file extension only from path for files other than index.json.
					path = strings.TrimSuffix(path, filepath.Ext(path))
				}

				// Remove the extension (if it exists) from single types since the filename = the type name.
				contentType = strings.TrimSuffix(contentType, filepath.Ext(contentType))

				// Get field key/values from content source.
				typeFields := readers.GetTypeFields(fileContentBytes)
				// Setup regex to find field name.
				reField := regexp.MustCompile(`:field\((.*?)\)`)
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
						path = slug
					}
				}

				// Setup regex to find pagination.
				_, rePaginate := getPagination()
				// Don't slugify paginated paths.
				if !rePaginate.MatchString(path) {
					// Create regex for allowed characters when slugifying path.
					reSlugify := regexp.MustCompile("[^a-z0-9/]+")
					// Slugify output using reSlugify regex defined above.
					path = strings.Trim(reSlugify.ReplaceAllString(strings.ToLower(path), "-"), "-")
				}

				// Remove trailing slash, unless it's the homepage.
				if path != "/" && path[len(path)-1:] == "/" {
					path = strings.TrimSuffix(path, "/")
				}

				destPath := buildPath + path + "/index.html"

				contentDetailsStr := "{\n" +
					"\"path\": \"" + path + "\",\n" +
					"\"type\": \"" + contentType + "\",\n" +
					"\"filename\": \"" + fileName + "\",\n" +
					"\"fields\": " + fileContentStr + "\n}"

				// Write to the content.js client data source file.
				writeContentJS(contentJSPath, contentDetailsStr+",")

				// Remove newlines, tabs, and extra space.
				encodedContentDetails := encodeString(contentDetailsStr)
				// Add info for being referenced in allContent object.
				allContentStr = allContentStr + encodedContentDetails + ","

				content := content{
					contentType:     contentType,
					contentPath:     path,
					contentDest:     destPath,
					contentDetails:  encodedContentDetails,
					contentFilename: fileName,
					contentFields:   encodeString(fileContentStr),
				}
				allContent = append(allContent, content)

				// Increment counter for logging purposes.
				contentFileCounter++

			}
		}
		return nil
	})
	if contentFilesErr != nil {
		fmt.Printf("Could not get layout file: %s", contentFilesErr)
	}

	// End the string that will be used in allContent object.
	allContentStr = strings.TrimSuffix(allContentStr, ",") + "]"

	tempBuildDirSignature := strings.ReplaceAll(strings.ReplaceAll(tempBuildDir, "/", "_"), ".", "_")
	for _, currentContent := range allContent {
		routeSignature := tempBuildDirSignature + "layout_content_" + currentContent.contentType + "_svelte"
		_, createPropsErr := SSRctx.RunScript("var props = {route: "+routeSignature+", content: "+currentContent.contentDetails+", allContent: "+allContentStr+"};", "create_ssr")
		if createPropsErr != nil {
			fmt.Printf("Could not create props: %v\n", createPropsErr)
		}
		// Render the HTML with props needed for the current content.
		_, renderHTMLErr := SSRctx.RunScript("var { html, css: staticCss} = "+tempBuildDirSignature+"layout_global_html_svelte.render(props);", "create_ssr")
		if renderHTMLErr != nil {
			fmt.Printf("Can't render htmlComponent: %v\n", renderHTMLErr)
		}

		paginate(currentContent, contentJSPath)

		createHTML(currentContent)

	}

	// Complete the content.js file.
	writeContentJS(contentJSPath, "];\n\nexport default contentSource;")

	Log("Number of content files used: " + strconv.Itoa(contentFileCounter))

}

func createHTML(currentContent content) {
	// Get the rendered HTML from v8go.
	renderedHTML, err := SSRctx.RunScript("html;", "create_ssr")
	if err != nil {
		fmt.Printf("V8go could not execute js default: %v\n", err)
	}
	// Get the string value of the static HTML.
	renderedHTMLStr := renderedHTML.String()
	// Inject the main.js script the starts the client-side app.
	renderedHTMLStr = strings.Replace(renderedHTMLStr, "</head>", "<script type='module' src='https://unpkg.com/dimport?module' data-main='/spa/ejected/main.js'></script><script nomodule src='https://unpkg.com/dimport/nomodule' data-main='/spa/ejected/main.js'></script></head>", 1)
	// Convert the string to byte array that can be written to file system.
	htmlBytes := []byte(renderedHTMLStr)
	// Create any folders need to write file.
	os.MkdirAll(strings.TrimSuffix(currentContent.contentDest, "/index.html"), os.ModePerm)
	// Write static HTML to the filesystem.
	htmlWriteErr := ioutil.WriteFile(currentContent.contentDest, htmlBytes, 0755)
	if htmlWriteErr != nil {
		fmt.Printf("Unable to write SSR file: %v\n", htmlWriteErr)
	}
}

func paginate(currentContent content, contentJSPath string) {
	paginatedContent, rePaginate := getPagination()
	// Loop through all :paginate() replacements found in config file.
	for _, pager := range paginatedContent {
		// Check if the config file specifies pagination for this Type.
		if pager.contentType == currentContent.contentType {
			// Copy the current content so we can increment the pager.
			newContent := currentContent
			for _, paginationVar := range pager.paginationVars {
				totalPages, getLocalVarErr := SSRctx.RunScript("plenti_global_pager_"+paginationVar, "create_ssr")
				if getLocalVarErr != nil {
					fmt.Printf("Could not get value of '%v' used in pager: %v\n", paginationVar, getLocalVarErr)
				}
				// Convert string total page value to integer.
				totalPagesInt, strToIntErr := strconv.Atoi(totalPages.String())
				if strToIntErr != nil {
					fmt.Printf("Can't convert pager value '%v' to an integer: %v", totalPages.String(), strToIntErr)
				}
				// Loop through total number of pages for current pager.
				for i := 0; i < totalPagesInt; i++ {
					// Update the path WIP
					currentPageNumber := strconv.Itoa(i + 1)
					newContent.contentPath = rePaginate.ReplaceAllString(pager.contentPath, currentPageNumber)
					newContent.contentDest = rePaginate.ReplaceAllString(currentContent.contentDest, currentPageNumber)
					// Add current page number to the content source so it can be pulled in as the current page.
					newContent.contentDetails = "{\n" +
						"\"pager\": \"" + currentPageNumber + "\",\n" +
						"\"path\": \"" + newContent.contentPath + "\",\n" +
						"\"type\": \"" + newContent.contentType + "\",\n" +
						"\"filename\": \"" + newContent.contentFilename + "\",\n" +
						"\"fields\": " + newContent.contentFields + "\n}"

					// Add paginated entries to content.js file.
					writeContentJS(contentJSPath, newContent.contentDetails+",")
					// Create paginated static HTML fallbacks.
					createHTML(newContent)
				}
			}
		}
	}
}

func writeContentJS(contentJSPath string, contentDetailsStr string) {
	// Create new content.js file if it doesn't already exist, or add to it if it does.
	contentJSFile, openContentJSErr := os.OpenFile(contentJSPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openContentJSErr != nil {
		fmt.Printf("Could not open content.js for writing: %s", openContentJSErr)
	}
	// Write to the file with info from current file in "/content" folder.
	defer contentJSFile.Close()
	if _, err := contentJSFile.WriteString(contentDetailsStr); err != nil {
		log.Println(err)
	}
}

func encodeString(encodedStr string) string {
	// Remove newlines.
	reN := regexp.MustCompile(`\r?\n`)
	encodedStr = reN.ReplaceAllString(encodedStr, " ")
	// Remove tabs.
	reT := regexp.MustCompile(`\t`)
	encodedStr = reT.ReplaceAllString(encodedStr, " ")
	// Reduce extra whitespace to a single space.
	reS := regexp.MustCompile(`\s+`)
	encodedStr = reS.ReplaceAllString(encodedStr, " ")
	return encodedStr
}
