package build

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/plentico/plenti/readers"
	"github.com/spf13/afero"
	"rogchap.com/v8go"
)

// Local is set to true when using dev webserver, otherwise bools default to false.
var Local bool

// Path404 stores path to 404 html fallback for use in local webserver
var Path404 string

// Doreload and other flags should probably be part of a config accessible across build.
// It gets set using server flags.
var Doreload bool
var (
	// Setup regex to find field name.
	reField = regexp.MustCompile(`:field\((.*?)\)`)
	// Setup regex to find pagination and a leading forward slash.
	rePaginate = regexp.MustCompile(`:paginate\((.*?)\)`)

	// Create regex for allowed characters when slugifying path.
	reSlugify = regexp.MustCompile("[^a-z0-9./*]+")
	// Remove newlines.
	reN = regexp.MustCompile(`\r?\n`)

	// Remove tabs.
	reT = regexp.MustCompile(`\t`)

	// Reduce extra whitespace to a single space.
	reS = regexp.MustCompile(`\s+`)
)

// Holds info related to a particular content node.
type content struct {
	contentType      string
	contentPath      string
	contentDest      string
	contentDetails   string
	contentFilepath  string
	contentFilename  string
	contentFields    string
	contentPagerDest string
	contentPagerPath string
	contentPagerNums []string
}

// Holds sitewide environment variables.
type env struct {
	local      string
	baseurl    string
	entrypoint string
	cms        cms
}
type cms struct {
	repo        string
	redirectUrl string
	appId       string
	branch      string
}

// DataSource builds json list from "content/" directory.
func DataSource(buildPath string, siteConfig readers.SiteConfig) error {

	defer Benchmark(time.Now(), "Creating data_source")

	Log("\nGathering data source from 'content/' folder")

	// Set some defaults
	contentJSPath := buildPath + "/spa/generated/content.js"
	envPath := buildPath + "/spa/generated/env.js"
	env := env{
		local:      strconv.FormatBool(Local),
		baseurl:    siteConfig.BaseURL,
		entrypoint: siteConfig.EntryPoint,
		cms: cms{
			repo:        siteConfig.CMS.Repo,
			redirectUrl: siteConfig.CMS.RedirectUrl,
			appId:       siteConfig.CMS.AppId,
			branch:      siteConfig.CMS.Branch,
		},
	}

	// Create env magic prop.
	envStr := "export let env = { local: " + env.local +
		", baseurl: '" + env.baseurl +
		"', entrypoint: '" + env.entrypoint +
		"', cms: { repo: '" + env.cms.repo +
		"', redirectUrl: '" + env.cms.redirectUrl +
		"', appId: '" + env.cms.appId +
		"', branch: '" + env.cms.branch +
		"' } };"

	// Start the new content.js file.
	err := ioutil.WriteFile(contentJSPath, []byte(`const allContent = [`), 0755)
	if err != nil {
		fmt.Printf("Unable to write content.js file: %v", err)
		return err
	}
	// Create the env.js file.
	err = ioutil.WriteFile(envPath, []byte(envStr), 0755)
	if err != nil {
		fmt.Printf("Unable to write env.js file: %v", err)
		return err
	}

	// Set up counter for logging output.
	contentFileCounter := 0
	// Start the string that will be used for allContent object.
	allContentStr := "["
	// Store each content file in array we can iterate over for creating static html.
	allContent := []content{}
	// Start the string that will be used for allDefaults object.
	allDefaultsStr := "const allDefaults = ["
	// Start the string that will be used for allSchemas object.
	allSchemasStr := "const allSchemas = {"
	// Start the string that will be used for allComponentDefaults object.
	allComponentDefaultsStr := "const allComponentDefaults = {"
	// Start the string that will be used for allComponentSchemas object.
	allComponentSchemasStr := "const allComponentSchemas = {"

	// Go through all sub directories in "content/" folder.
	if ThemeFs != nil {
		if err := afero.Walk(ThemeFs, "content", func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			contentFileCounter, allContentStr, allContent, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr, err = getContent(path, info, err, siteConfig, buildPath, contentJSPath, allContentStr, allContent, contentFileCounter, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			return fmt.Errorf("\nCould not get all content from virtual theme %w", err)
		}
	} else {
		if err := filepath.Walk("content", func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			contentFileCounter, allContentStr, allContent, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr, err = getContent(path, info, err, siteConfig, buildPath, contentJSPath, allContentStr, allContent, contentFileCounter, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			return fmt.Errorf("\nCould not get all content from project %w", err)
		}
	}

	// End the string that will be used in allContent object.
	allContentStr = strings.TrimSuffix(allContentStr, ",") + "]"
	// End the string that will be used in allDefaults object.
	allDefaultsStr = strings.TrimSuffix(allDefaultsStr, ",") + "];\n\nexport default allDefaults;"
	err = writeContentJS(buildPath+"/spa/generated/defaults.js", allDefaultsStr)
	if err != nil {
		return fmt.Errorf("\nCould not write defaults.js file")
	}
	// End the string that will be used in allSchemas object.
	allSchemasStr = strings.TrimSuffix(allSchemasStr, ",") + "\n};\n\nexport default allSchemas;"
	err = writeContentJS(buildPath+"/spa/generated/schemas.js", allSchemasStr)
	if err != nil {
		return fmt.Errorf("\nCould not write schemas.js file")
	}
	// End the string that will be used in allComponentDefaults object.
	allComponentDefaultsStr = strings.TrimSuffix(allComponentDefaultsStr, ",") + "\n};\n\nexport default allComponentDefaults;"
	err = writeContentJS(buildPath+"/spa/generated/component_defaults.js", allComponentDefaultsStr)
	if err != nil {
		return fmt.Errorf("\nCould not write component_defaults.js file")
	}
	// End the string that will be used in allComponentSchemas object.
	allComponentSchemasStr = strings.TrimSuffix(allComponentSchemasStr, ",") + "\n};\n\nexport default allComponentSchemas;"
	err = writeContentJS(buildPath+"/spa/generated/component_schemas.js", allComponentSchemasStr)
	if err != nil {
		return fmt.Errorf("\nCould not write component_schemas.js file")
	}

	//var wg sync.WaitGroup
	wg := sync.WaitGroup{}
	for _, currentContent := range allContent {
		wg.Add(1)
		go func(currentContent content, wg *sync.WaitGroup) error {
			defer wg.Done()
			fmt.Println("creating SSR ctx")
			ssrCtx, err := createSSRCtx()
			if err != nil {
				return fmt.Errorf("\nCan't create SSR Context for %s %w", currentContent.contentFilepath, err)
			}

			fmt.Println("creating Props")
			ssrCtx, err = createProps(ssrCtx, currentContent, allContentStr, env)
			if err != nil {
				return fmt.Errorf("\nCan't create props for %s %w", currentContent.contentFilepath, err)
			}

			fmt.Println("creating HTML")
			err = createHTML(ssrCtx, currentContent)
			if err != nil {
				return fmt.Errorf("\nCan't create HTML for %s %w", currentContent.contentFilepath, err)
			}

			fmt.Println("Paginating")
			allPaginatedContent, err := paginate(ssrCtx, currentContent, contentJSPath)
			if err != nil {
				return err
			}
			for _, paginatedContent := range allPaginatedContent {
				if ssrCtx, err = createProps(ssrCtx, paginatedContent, allContentStr, env); err != nil {
					return err
				}

				if err = createHTML(ssrCtx, paginatedContent); err != nil {
					return err
				}

			}
			ssrCtx.Close()
			//ssrCtx.Isolate().TerminateExecution()
			//ssrCtx.Isolate().Dispose()
			return nil
		}(currentContent, &wg)

	}
	wg.Wait()

	Log("Number of content files used: " + fmt.Sprint(contentFileCounter))
	// Complete the content.js file.
	if err := writeContentJS(contentJSPath, "];\n\nexport default allContent;"); err != nil {
		return err
	}
	return nil

}

func createSSRCtx() (*v8go.Context, error) {
	ssrCtx := v8go.NewContext(nil)
	// Fix "ReferenceError: exports is not defined" errors on line 1319 (exports.current_component;).
	if _, err := ssrCtx.RunScript("var exports = {};", "create_ssr"); err != nil {
		return nil, err
	}

	var svelteLibs = [6]string{
		"node_modules/svelte/animate/index.js",
		"node_modules/svelte/easing/index.js",
		"node_modules/svelte/internal/index.js",
		"node_modules/svelte/motion/index.js",
		"node_modules/svelte/store/index.js",
		"node_modules/svelte/transition/index.js",
	}

	for _, svelteLib := range svelteLibs {
		// Use v8go and add create_ssr_component() function.
		createSsrComponent, err := getVirtualFileIfThemeBuild(svelteLib)
		if err != nil {
			return nil, err

		}
		// Fix: TypeError: Cannot read properties of undefined (reading 'noop')
		createSsrStr := strings.ReplaceAll(string(createSsrComponent), "internal.noop", "noop")
		_, err = ssrCtx.RunScript(createSsrStr, "create_ssr")
		/*
			// TODO: Can't check error because `ReferenceError: require is not defined` error on build so cannot quit ...
			if err != nil {
				fmt.Println(fmt.Errorf("Could not add create_ssr_component() func from svelte/internal for file %s: %w%s\n", svelteLib, err, common.Caller()))
			}
		*/

	}
	// Add all SSR components to context for HTML rendering
	if err := afero.Walk(SSRFs, ".", func(layoutPath string, layoutFileInfo os.FileInfo, err error) error {
		//fmt.Println("virtual file: " + layoutPath)
		if layoutFileInfo.IsDir() {
			return nil
		}
		ssrBytes, err := afero.ReadFile(SSRFs, layoutPath)
		ssrStr := string(ssrBytes)
		//fmt.Println(ssrStr)
		_, err = ssrCtx.RunScript(ssrStr, "create_ssr")
		if err != nil {
			return fmt.Errorf("Could not add SSR Component for %s: %w\n", layoutPath, err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("\nCould not get layout from virtual theme build %w", err)
	}
	return ssrCtx, nil
}

func getContent(path string, info os.FileInfo, err error, siteConfig readers.SiteConfig,
	buildPath string, contentJSPath string, allContentStr string, allContent []content,
	contentFileCounter int, allDefaultsStr string, allSchemasStr string, allComponentDefaultsStr string, allComponentSchemasStr string) (int, string, []content, string, string, string, string, error) {

	if err != nil {
		return contentFileCounter, allContentStr, allContent, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr, fmt.Errorf("can't stat %s: %w", path, err)
	}

	filePath, contentType, fileName := getFileInfo(path)

	// Don't process hidden files, like .DS_Store
	if fileName[:1] == "." {
		// Skip silently so we don't stop the build or clutter the terminal
		return contentFileCounter, allContentStr, allContent, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr, nil
	}

	// Get the contents of the file.
	fileContentBytes, err := getVirtualFileIfThemeBuild(path)
	if err != nil {
		return contentFileCounter, allContentStr, allContent, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr, fmt.Errorf("file: %s %w\n", path, err)
	}
	fileContentStr := string(fileContentBytes)

	// Remove the extension (if it exists) from single types since the filename = the type name.
	contentType = strings.TrimSuffix(contentType, filepath.Ext(contentType))

	// Get field key/values from content source.
	typeFields, err := readers.GetTypeFields(fileContentBytes)
	if err != nil {
		return contentFileCounter, allContentStr, allContent, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr, fmt.Errorf("\nError getting content from %s %w", filePath, err)
	}

	// Setup regex to find field name.
	reField := regexp.MustCompile(`:fields\((.*?)\)`)
	// Check for path overrides from plenti.json config file.
	for configContentType, slug := range siteConfig.Routes {
		if configContentType == contentType {
			// Replace :filename.
			slug = strings.Replace(slug, ":filename", fileName, -1)

			// Replace :fields().
			fieldReplacements := reField.FindAllStringSubmatch(slug, -1)
			// Loop through all :fields() replacements found in config file.
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

	// Initialize vars for path with replacement patterns still intact.
	var pagerPath string
	var pagerDestPath string
	// If there is a /:paginate() replacement found.
	if rePaginate.MatchString(path) {
		// Save path before slugifying to preserve pagination.
		pagerPath = path
		// Get Destination path before slugifying to preserve pagination.
		pagerDestPath = buildPath + "/" + path + "/index.html"
		// Remove /:pagination()
		path = rePaginate.ReplaceAllString(path, "")
	}

	// Convert path to a route the browser can understand
	path = makeWebPath(path, fileName)
	path = slugify(path)
	path = fixBlankPaths(path)
	path = removeExtraSlashes(path)

	// Add "public" folder and remove wildcards
	destPath := buildPath + "/" + strings.Replace(path, "*", "", -1)

	if !strings.HasSuffix(destPath, ".html") {
		destPath = destPath + "/index.html"
		//fmt.Println(destPath)
	}

	// Set 404 path for local webserver
	if contentType == "404" {
		Path404 = path
	}

	// Don't add _components
	if contentType == "_components" {
		component := strings.Split(filePath, "/")[2]
		configType := strings.Split(filePath, "/")[3]
		if configType == "_defaults.json" {
			componentDetailsStr := "\n\"" + component + "\": " + fileContentStr
			allComponentDefaultsStr = allComponentDefaultsStr + componentDetailsStr + ","
		}
		if configType == "_schema.json" {
			componentSchemaDetailsStr := "\n\"" + component + "\": " + fileContentStr
			allComponentSchemasStr = allComponentSchemasStr + componentSchemaDetailsStr + ","
		}

		return contentFileCounter, allContentStr, allContent, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr, nil
	}
	// Don't add _defaults.json
	if fileName == "_defaults.json" {
		defaultsDetailsStr := "{\n" +
			"\"pager\": null,\n" +
			"\"type\": \"" + contentType + "\",\n" +
			"\"path\": \"" + path + "\",\n" +
			"\"filepath\": \"" + filePath + "\",\n" +
			"\"filename\": \"" + fileName + "\",\n" +
			"\"fields\": " + fileContentStr + "\n}"

		allDefaultsStr = allDefaultsStr + defaultsDetailsStr + ","

		return contentFileCounter, allContentStr, allContent, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr, nil
	}
	// Don't add _schema.json
	if strings.HasPrefix(fileName, "_schema") {
		// Remove _schema_ prefix for single types
		contentType = strings.TrimPrefix(contentType, "_schema_")
		schemaDetailsStr := "\n\"" + contentType + "\": " + fileContentStr

		allSchemasStr = allSchemasStr + schemaDetailsStr + ","

		return contentFileCounter, allContentStr, allContent, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr, nil
	}

	contentDetailsStr := "{\n" +
		"\"pager\": null,\n" +
		"\"type\": \"" + contentType + "\",\n" +
		"\"path\": \"" + path + "\",\n" +
		"\"filepath\": \"" + filePath + "\",\n" +
		"\"filename\": \"" + fileName + "\",\n" +
		"\"fields\": " + fileContentStr + "\n}"

	// Write to the content.js client data source file.
	if err = writeContentJS(contentJSPath, contentDetailsStr+","); err != nil {
		return contentFileCounter, allContentStr, allContent, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr, fmt.Errorf("file: %s %w\n", contentJSPath, err)
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
		contentFilepath:  filePath,
		contentFilename:  fileName,
		contentFields:    encodeString(fileContentStr),
		contentPagerDest: pagerDestPath,
		contentPagerPath: pagerPath,
	}
	allContent = append(allContent, content)

	// Increment counter for logging purposes.
	contentFileCounter++

	return contentFileCounter, allContentStr, allContent, allDefaultsStr, allSchemasStr, allComponentDefaultsStr, allComponentSchemasStr, nil
}

func getFileInfo(path string) (string, string, string) {
	filePath := path
	// Get individual path arguments.
	parts := strings.Split(path, "/")
	contentType := parts[1]
	fileName := parts[len(parts)-1]
	return filePath, contentType, fileName
}

func makeWebPath(path string, fileName string) string {
	// Remove the "content/" folder from path.
	path = strings.TrimPrefix(path, "content/")
	// Check for index file at any level.
	if fileName == "_index.json" {
		// Remove entire filename from path.
		path = strings.TrimSuffix(path, fileName)
	} else {
		// Remove file extension only from path for files other than index.json.
		path = strings.TrimSuffix(path, ".json")
	}
	return path
}

func slugify(path string) string {
	// Slugify output using reSlugify regex defined above.
	return strings.Trim(reSlugify.ReplaceAllString(strings.ToLower(path), "-"), "-")
}

func fixBlankPaths(path string) string {
	// Check for any missing/blank paths.
	if path == "" {
		// Add the forward slash back for the index page.
		return "/"
	}
	// Let the user know if path is blank. <- TODO: Can this actual run?
	if len(path) < 1 {
		fmt.Println("Content path can't be blank, check your route overrides in plenti.json.")
	}
	return path
}

func removeExtraSlashes(path string) string {
	// Remove any repeating forward slashes.
	path = filepath.Clean(path)
	// Remove trailing slash, unless it's the homepage.
	if path != "/" && path[len(path)-1:] == "/" {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

func createProps(ssrCtx *v8go.Context, currentContent content, allContentStr string, env env) (*v8go.Context, error) {
	componentSignature := "layouts_content_" + currentContent.contentType + "_svelte"
	//fmt.Println("create props: " + componentSignature)
	_, err := ssrCtx.RunScript("var props = {content: "+currentContent.contentDetails+
		", layout: "+componentSignature+
		", allContent: "+allContentStr+
		", shadowContent: {}"+
		", env: {local: "+env.local+
		", baseurl: '"+env.baseurl+
		"', cms: { repo: '"+env.cms.repo+
		"', redirectUrl: '"+env.cms.redirectUrl+
		"', appId: '"+env.cms.appId+
		"', branch: '"+env.cms.branch+
		"'}}};", "create_ssr")
	if err != nil {
		return nil, fmt.Errorf("\nCould not create props for %s\n%+v", componentSignature, err)
	}
	// Render the HTML with props needed for the current content.
	entrySignature := strings.ReplaceAll(
		strings.ReplaceAll(
			"layouts/"+env.entrypoint,
			"/", "_"),
		".", "_")
	_, err = ssrCtx.RunScript(fmt.Sprintf("var { html, css: staticCss} = %s.render(props);", entrySignature), "create_ssr")
	if err != nil {
		return nil, fmt.Errorf("\nCan't render htmlComponent for %s\n%+v", componentSignature, err)
	}
	return ssrCtx, nil
}

func createHTML(ssrCtx *v8go.Context, currentContent content) error {
	// Get the rendered HTML from v8go.
	renderedHTML, err := ssrCtx.RunScript("html;", "create_ssr")
	if err != nil {
		return fmt.Errorf("V8go could not execute js default: %w\n", err)

	}
	// Get the string value of the static HTML.
	renderedHTMLStr := renderedHTML.String()
	// Convert the string to byte array that can be written to file system.
	htmlBytes := []byte(renderedHTMLStr)
	// Inject <!DOCTYPE html>
	htmlBytes = bytes.Replace(htmlBytes, []byte("<html"), []byte("<!DOCTYPE html><html"), 1)
	// Inject data-content-filepath attribute
	htmlBytes = bytes.Replace(htmlBytes, []byte("<html"), []byte("<html data-content-filepath='"+currentContent.contentFilepath+"' "), 1)
	if Doreload {
		// Inject live-reload script (stored in ejected core).
		htmlBytes = bytes.Replace(htmlBytes, []byte("</body>"), []byte("<script type='text/javascript' src='/spa/ejected/live-reload.js'></script></body>"), 1)
	}
	// Create any folders need to write file.
	if err := os.MkdirAll(strings.TrimSuffix(currentContent.contentDest, path.Base(currentContent.contentDest)), os.ModePerm); err != nil {
		return fmt.Errorf("couldn't create dirs in createHTML: %w\n", err)
	}
	// Write static HTML to the filesystem.
	err = ioutil.WriteFile(currentContent.contentDest, htmlBytes, 0755)
	if err != nil {
		return fmt.Errorf("unable to write SSR file: %w\n", err)
	}
	return nil
}

func paginate(ssrCtx *v8go.Context, currentContent content, contentJSPath string) ([]content, error) {
	paginatedContent, _ := getPagination()
	var err error
	allNewContent := []content{}
	// Loop through all :paginate() replacements found in config file.
	for _, pager := range paginatedContent {
		// Check if the config file specifies pagination for this Type.
		if len(pager.paginationVars) > 0 && pager.contentType == currentContent.contentType {
			// Increment the pager.
			allNewContent, err = incrementPager(ssrCtx, pager.paginationVars, currentContent, contentJSPath, allNewContent)
			if err != nil {
				return nil, err
			}
		}
	}
	return allNewContent, err
}

func incrementPager(ssrCtx *v8go.Context, paginationVars []string, currentContent content, contentJSPath string, allNewContent []content) ([]content, error) {
	// Pop first item from the list.
	paginationVar, paginationVars := paginationVars[0], paginationVars[1:]
	// Copy the current content so we can increment the pager.
	newContent := currentContent
	// Get the number of pages for the pager.
	totalPagesInt, err := getTotalPages(ssrCtx, paginationVar)
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
			allNewContentTmp, err := incrementPager(ssrCtx, paginationVars, newContent, contentJSPath, allNewContent)
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
			"\"type\": \"" + newContent.contentType + "\",\n" +
			"\"path\": \"" + newContent.contentPath + "\",\n" +
			"\"filepath\": \"" + newContent.contentFilepath + "\",\n" +
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

func getTotalPages(ssrCtx *v8go.Context, paginationVar string) (int, error) {
	totalPages, err := ssrCtx.RunScript("plenti_global_pager_"+paginationVar, "create_ssr")
	if err != nil {
		return 0, fmt.Errorf("Could not get value of '%v' used in pager: %w\n", paginationVar, err)
	}
	// Convert string total page value to integer.
	totalPagesInt, err := strconv.Atoi(totalPages.String())
	if err != nil {
		return 0, fmt.Errorf("Can't convert pager value '%s' to an integer: %w\n", totalPages.String(), err)
	}
	return totalPagesInt, nil
}

func writeContentJS(contentJSPath string, contentDetailsStr string) error {
	// Create new content.js file if it doesn't already exist, or add to it if it does.
	contentJSFile, err := os.OpenFile(contentJSPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open content.js for writing: %w\n", err)
	}
	// Write to the file with info from current file in "/content" folder.
	defer contentJSFile.Close()
	if _, err := contentJSFile.WriteString(contentDetailsStr); err != nil {

		return fmt.Errorf("could not write to file %s: %w\n", contentJSPath, err)

	}
	return nil
}

func encodeString(encodedStr string) string {
	// Remove newlines.
	encodedStr = reN.ReplaceAllString(encodedStr, " ")
	// Remove tabs.

	encodedStr = reT.ReplaceAllString(encodedStr, " ")
	// Reduce extra whitespace to a single space.

	encodedStr = reS.ReplaceAllString(encodedStr, " ")
	return encodedStr
}
