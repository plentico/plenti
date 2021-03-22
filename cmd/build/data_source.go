package build

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/plentico/plenti/common"
	"github.com/plentico/plenti/readers"
)

// Doreload and other flags should probably be part of a config accessible across build.
// It gets set using server flags.
var Doreload bool

// inject wherever/however, this is just to get it working.

var scr = []byte(`
<script>
let socket;

let pop = document.createElement('div');
pop.style.cssText = 'font-size: 18px;position:absolute;top:60px;right:10px;padding:30px;color:#fff;z-index:100;border-radius: 8px;background:#599118';

window.onload = () => {
  console.debug("onload");
  
  pop.innerHTML= "Reload Connected..."
  document.body.appendChild(pop);
  setTimeout(function(){
	  pop.style.display="none";
  },1500)
  

};
document.addEventListener("DOMContentLoaded", () => {
	socket = new WebSocket('ws://' + window.location.host + '/reload');

	socket.onmessage = function (event) {
	  var datastr = event.data.split(":")[0];
	   if (datastr === 'reload'){
	
		console.debug("reloading")
		window.location.reload();
		
		  
	   }
	};
    
	socket.onopen = function() {
	  console.debug("connected");
	  socket.send("loaded");
	
	}
  
	socket.onclose = function(e) {
	pop.innerHTML= "Reload Disconnected...";
	pop.style.display="block";
	pop.style.backgroundColor="rgb(236,64,3)";
	console.debug("connection closed (" + e.code + ")");
	}
  
  });
</script></body>
`)

type content struct {
	contentType      string
	contentPath      string
	contentDest      string
	contentDetails   string
	contentFilename  string
	contentFields    string
	contentPagerDest string
	contentPagerPath string
	contentPagerNums []string
}

// DataSource builds json list from "content/" directory.
func DataSource(buildPath string, siteConfig readers.SiteConfig, tempBuildDir string) error {

	defer Benchmark(time.Now(), "Creating data_source")

	Log("\nGathering data source from 'content/' folder")

	contentJSPath := buildPath + "/spa/ejected/content.js"
	if err := os.MkdirAll(buildPath+"/spa/ejected", os.ModePerm); err != nil {
		return err
	}

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
		return err
	}

	// Go through all sub directories in "content/" folder.
	contentFilesErr := filepath.Walk(tempBuildDir+"content", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", path, err)
		}
		if !info.IsDir() {
			// Get individual path arguments.
			parts := strings.Split(path, "/")
			contentType := parts[1]
			if tempBuildDir != "" {
				contentType = parts[2]
			}
			fileName := parts[len(parts)-1]

			// Don't add _blueprint.json or other special named files starting with underscores.
			if fileName[:1] != "_" && fileName[:1] != "." {

				// Get the contents of the file.
				fileContentBytes, err := ioutil.ReadFile(path)
				if err != nil {

					return fmt.Errorf("file: %s %w%s", path, err, common.Caller())
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
				for configContentType, slug := range siteConfig.Routes {
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

				// Setup regex to find pagination and a leading forward slash.
				rePaginate := regexp.MustCompile(`/:paginate\((.*?)\)`)
				// Initialize vars for path with replacement patterns still intact.
				var pagerPath string
				var pagerDestPath string
				// If there is a /:paginate() replacement found.
				if rePaginate.MatchString(path) {
					// Save path before slugifying to preserve pagination.
					pagerPath = path
					// Get Destination path before slugifying to preserve pagination.
					pagerDestPath = buildPath + path + "/index.html"
					// Remove /:pagination()
					path = rePaginate.ReplaceAllString(path, "")
					// If paginating the homepage, the forward slash shouldn't be removed.
					if path == "" {
						// Add the forward slash back for the index page.
						path = "/"
					}
				}

				// Create regex for allowed characters when slugifying path.
				reSlugify := regexp.MustCompile("[^a-z0-9/]+")
				// Slugify output using reSlugify regex defined above.
				path = strings.Trim(reSlugify.ReplaceAllString(strings.ToLower(path), "-"), "-")

				// Remove trailing slash, unless it's the homepage.
				if path != "/" && path[len(path)-1:] == "/" {
					path = strings.TrimSuffix(path, "/")
				}

				destPath := buildPath + path + "/index.html"

				contentDetailsStr := "{\n" +
					"\"pager\": 1,\n" +
					"\"path\": \"" + path + "\",\n" +
					"\"type\": \"" + contentType + "\",\n" +
					"\"filename\": \"" + fileName + "\",\n" +
					"\"fields\": " + fileContentStr + "\n}"

				// Write to the content.js client data source file.
				if err = writeContentJS(contentJSPath, contentDetailsStr+","); err != nil {
					return fmt.Errorf("file: %s %w%s", contentJSPath, err, common.Caller())
				}

				// Remove newlines, tabs, and extra space.
				encodedContentDetails := encodeString(contentDetailsStr)
				// Add info for being referenced in allContent object.
				allContentStr = allContentStr + encodedContentDetails + ","

				content := content{
					contentType:      contentType,
					contentPath:      path,
					contentDest:      destPath,
					contentDetails:   encodedContentDetails,
					contentFilename:  fileName,
					contentFields:    encodeString(fileContentStr),
					contentPagerDest: pagerDestPath,
					contentPagerPath: pagerPath,
				}
				allContent = append(allContent, content)

				// Increment counter for logging purposes.
				contentFileCounter++

			}
		}
		return nil
	})
	if contentFilesErr != nil {
		return fmt.Errorf("Could not get layout %w", contentFilesErr)

	}

	// End the string that will be used in allContent object.
	allContentStr = strings.TrimSuffix(allContentStr, ",") + "]"

	for _, currentContent := range allContent {

		if err = createProps(currentContent, allContentStr); err != nil {
			return err
		}

		if err = createHTML(currentContent); err != nil {
			return err
		}

		allPaginatedContent, err := paginate(currentContent, contentJSPath)
		if err != nil {
			return err
		}
		for _, paginatedContent := range allPaginatedContent {
			if err = createProps(paginatedContent, allContentStr); err != nil {
				return err
			}

			if err = createHTML(paginatedContent); err != nil {
				return err
			}

		}

	}

	Log("Number of content files used: " + fmt.Sprint(contentFileCounter))
	// Complete the content.js file.
	return writeContentJS(contentJSPath, "];\n\nexport default contentSource;")

}

func createProps(currentContent content, allContentStr string) error {
	routeSignature := "layouts_content_" + currentContent.contentType + "_svelte"
	_, err := SSRctx.RunScript("var props = {route: "+routeSignature+", content: "+currentContent.contentDetails+", allContent: "+allContentStr+"};", "create_ssr")
	if err != nil {

		return fmt.Errorf("Could not create props: %w%s", err, common.Caller())

	}
	// Render the HTML with props needed for the current content.
	_, err = SSRctx.RunScript("var { html, css: staticCss} = layouts_global_html_svelte.render(props);", "create_ssr")
	if err != nil {
		return fmt.Errorf("Can't render htmlComponent: %w%s", err, common.Caller())

	}
	return nil
}

func createHTML(currentContent content) error {
	// Get the rendered HTML from v8go.
	renderedHTML, err := SSRctx.RunScript("html;", "create_ssr")
	if err != nil {
		return fmt.Errorf("V8go could not execute js default: %w%s", err, common.Caller())

	}
	// Get the string value of the static HTML.
	renderedHTMLStr := renderedHTML.String()
	// Convert the string to byte array that can be written to file system.
	htmlBytes := []byte(renderedHTMLStr)
	if Doreload {
		htmlBytes = bytes.Replace(htmlBytes, []byte("</body>"), scr, 1)

	}
	// Create any folders need to write file.
	if err := os.MkdirAll(strings.TrimSuffix(currentContent.contentDest, "/index.html"), os.ModePerm); err != nil {
		return fmt.Errorf("couldn't create dirs in createHTML: %w%s", err, common.Caller())
	}
	// Write static HTML to the filesystem.
	err = ioutil.WriteFile(currentContent.contentDest, htmlBytes, 0755)
	if err != nil {
		return fmt.Errorf("unable to write SSR file: %w%s", err, common.Caller())
	}
	return nil
}

func paginate(currentContent content, contentJSPath string) ([]content, error) {
	paginatedContent, _ := getPagination()
	var err error
	allNewContent := []content{}
	// Loop through all :paginate() replacements found in config file.
	for _, pager := range paginatedContent {
		// Check if the config file specifies pagination for this Type.
		if len(pager.paginationVars) > 0 && pager.contentType == currentContent.contentType {
			// Increment the pager.
			allNewContent, err = incrementPager(pager.paginationVars, currentContent, contentJSPath, allNewContent)
			if err != nil {
				return nil, err
			}
		}
	}
	return allNewContent, err
}

func incrementPager(paginationVars []string, currentContent content, contentJSPath string, allNewContent []content) ([]content, error) {
	// Pop first item from the list.
	paginationVar, paginationVars := paginationVars[0], paginationVars[1:]
	// Copy the current content so we can increment the pager.
	newContent := currentContent
	// Get the number of pages for the pager.
	totalPagesInt, err := getTotalPages(paginationVar)
	if err != nil {
		return nil, err
	}
	// Loop through total number of pages for current pager.
	for i := 1; i <= totalPagesInt; i++ {
		// Convert page number to a string that can be used in a path.
		currentPageNumber := strconv.Itoa(i)
		// Update the path by replacing the current :paginate() pattern.
		newContent.contentPath = strings.Replace(currentContent.contentPagerPath, ":paginate("+paginationVar+")", currentPageNumber, 1)
		newContent.contentDest = strings.Replace(currentContent.contentPagerDest, ":paginate("+paginationVar+")", currentPageNumber, 1)
		// Now we need to update the pager path to replace the pattern with a number so it's ready if we call incrementPager() again for a second pager.
		newContent.contentPagerPath = newContent.contentPath
		newContent.contentPagerDest = newContent.contentDest

		// Collect each pager value.
		newContent.contentPagerNums = append(newContent.contentPagerNums, currentPageNumber)

		// Check if there are more pagers for the route override.
		if len(paginationVars) > 0 {
			// Recursively call func to increment second pager.
			// todo: a better approach
			allNewContentTmp, err := incrementPager(paginationVars, newContent, contentJSPath, allNewContent)
			if err != nil {
				return nil, err
			}
			allNewContent = allNewContentTmp
			// Remove first item in the array to move to the next number in the first pager.
			newContent.contentPagerNums = newContent.contentPagerNums[1:]
			// Continue because you don't want to complete the loop with a partially updated path (we have more pagers!).
			continue
		}

		// Set the content.pager value if only one pager is used.
		pageNums := newContent.contentPagerNums[0]
		// Check if there are multiple pagers in the router override.
		if len(newContent.contentPagerNums) > 1 {
			// Make the content.pager value an array with the current pager values.
			pageNums = "[" + strings.Join(newContent.contentPagerNums, ", ") + "]"
		}

		// Add current page number to the content source so it can be pulled in as the current page.
		newContent.contentDetails = "{\n" +
			"\"pager\": " + pageNums + ",\n" +
			"\"path\": \"" + newContent.contentPath + "\",\n" +
			"\"type\": \"" + newContent.contentType + "\",\n" +
			"\"filename\": \"" + newContent.contentFilename + "\",\n" +
			"\"fields\": " + newContent.contentFields + "\n}"

		// Add paginated entries to content.js file.
		if err := writeContentJS(contentJSPath, newContent.contentDetails+","); err != nil {
			return nil, err
		}
		// Add to array of content for creating paginated static HTML fallbacks.
		allNewContent = append(allNewContent, newContent)

		// Remove last number from array to get next page in current pager.
		newContent.contentPagerNums = newContent.contentPagerNums[:len(newContent.contentPagerNums)-1]
	}
	return allNewContent, nil
}

func getTotalPages(paginationVar string) (int, error) {
	totalPages, err := SSRctx.RunScript("plenti_global_pager_"+paginationVar, "create_ssr")
	if err != nil {
		return 0, fmt.Errorf("Could not get value of '%v' used in pager: %w%s", paginationVar, err, common.Caller())
	}
	// Convert string total page value to integer.
	totalPagesInt, err := strconv.Atoi(totalPages.String())
	if err != nil {
		return 0, fmt.Errorf("Can't convert pager value '%s' to an integer: %w%s", totalPages.String(), err, common.Caller())
	}
	return totalPagesInt, nil
}

func writeContentJS(contentJSPath string, contentDetailsStr string) error {
	// Create new content.js file if it doesn't already exist, or add to it if it does.
	contentJSFile, err := os.OpenFile(contentJSPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open content.js for writing: %w%s", err, common.Caller())
	}
	// Write to the file with info from current file in "/content" folder.
	defer contentJSFile.Close()
	if _, err := contentJSFile.WriteString(contentDetailsStr); err != nil {

		return fmt.Errorf("could not write to file %s: %w%s", contentJSPath, err, common.Caller())

	}
	return nil
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
