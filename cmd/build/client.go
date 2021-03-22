package build

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/plentico/plenti/common"
	"github.com/plentico/plenti/readers"

	"rogchap.com/v8go"
)

// SSRctx is a v8go context for loaded with components needed to render HTML.
var SSRctx *v8go.Context

// Client builds the SPA.
func Client(buildPath string, tempBuildDir string, defaultsEjectedFS embed.FS) error {

	defer Benchmark(time.Now(), "Compiling client SPA with Svelte")

	Log("\nCompiling client SPA with svelte")

	stylePath := buildPath + "/spa/bundle.css"

	// Initialize string for layout.js component list.
	var allComponentsStr string

	// Set up counter for logging output.
	compiledComponentCounter := 0

	// Get svelte compiler code from node_modules.
	compiler, err := ioutil.ReadFile(
		fmt.Sprintf("%snode_modules/svelte/compiler.js", tempBuildDir),
	)
	if err != nil {
		return fmt.Errorf("Can't read %s/node_modules/svelte/compiler.js: %w%s", tempBuildDir, err, common.Caller())

	}
	// Remove reference to 'self' that breaks v8go on line 19 of node_modules/svelte/compiler.js.
	compilerStr := strings.Replace(string(compiler), "self.performance.now();", "'';", 1)
	// Remove 'require' that breaks v8go on line 22647 of node_modules/svelte/compiler.js.
	compilerStr = strings.Replace(compilerStr, "const Url$1 = (typeof URL !== 'undefined' ? URL : require('url').URL);", "", 1)
	ctx, err := v8go.NewContext(nil)
	if err != nil {
		return fmt.Errorf("Could not create Isolate: %w%s", err, common.Caller())

	}
	_, err = ctx.RunScript(compilerStr, "compile_svelte")
	if err != nil {
		return fmt.Errorf("Could not add svelte compiler: %w%s", err, common.Caller())

	}

	SSRctx, err = v8go.NewContext(nil)
	if err != nil {
		return fmt.Errorf("Could not create Isolate: %w%s", err, common.Caller())

	}
	// Fix "ReferenceError: exports is not defined" errors on line 1319 (exports.current_component;).
	if _, err := SSRctx.RunScript("var exports = {};", "create_ssr"); err != nil {
		return err
	}
	// Fix "TypeError: Cannot read property 'noop' of undefined" from node_modules/svelte/store/index.js.
	if _, err := SSRctx.RunScript("function noop(){}", "create_ssr"); err != nil {
		return err
	}

	var svelteLibs = [6]string{
		tempBuildDir + "node_modules/svelte/animate/index.js",
		tempBuildDir + "node_modules/svelte/easing/index.js",
		tempBuildDir + "node_modules/svelte/internal/index.js",
		tempBuildDir + "node_modules/svelte/motion/index.js",
		tempBuildDir + "node_modules/svelte/store/index.js",
		tempBuildDir + "node_modules/svelte/transition/index.js",
	}

	for _, svelteLib := range svelteLibs {
		// Use v8go and add create_ssr_component() function.
		createSsrComponent, err := ioutil.ReadFile(svelteLib)
		if err != nil {
			return fmt.Errorf("Can't read %s: %w%s", svelteLib, err, common.Caller())

		}
		// Fix "Cannot access 'on_destroy' before initialization" errors on line 1320 & line 1337 of node_modules/svelte/internal/index.js.
		createSsrStr := strings.ReplaceAll(string(createSsrComponent), "function create_ssr_component(fn) {", "function create_ssr_component(fn) {var on_destroy= {};")
		// Use empty noop() function created above instead of missing method.
		createSsrStr = strings.ReplaceAll(createSsrStr, "internal.noop", "noop")
		_, err = SSRctx.RunScript(createSsrStr, "create_ssr")

		// TODO: `ReferenceError: require is not defined` error on build so cannot quit ...
		// 	node_modules/svelte/animate/index.js: ReferenceError: require is not defined
		//  node_modules/svelte/easing/index.js: ReferenceError: require is not defined
		//  node_modules/svelte/motion/index.js: ReferenceError: require is not defined
		//  node_modules/svelte/store/index.js: ReferenceError: require is not defined
		//  node_modules/svelte/transition/index.js: ReferenceError: require is not defined
		// if err != nil {
		// 	fmt.Println(fmt.Errorf("Could not add create_ssr_component() func from svelte/internal for file %s: %w%s", svelteLib, err, common.Caller()))

		// }

	}

	routerPath := tempBuildDir + "ejected/router.svelte"
	var componentStr string
	if _, err := os.Stat(routerPath); err == nil {
		// The router has been ejected to the filesystem.
		component, err := ioutil.ReadFile(routerPath)
		if err != nil {
			return fmt.Errorf("can't read component file: %s %w%s", routerPath, err, common.Caller())
		}
		componentStr = string(component)
	} else if os.IsNotExist(err) {
		// The router has not been ejected, use the embedded defaults.
		ejected, err := fs.Sub(defaultsEjectedFS, "defaults")
		if err != nil {
			common.CheckErr(fmt.Errorf("Unable to get ejected defaults: %w", err))
		}
		ejected.Open(routerPath)
		routerComp, _ := ejected.Open(routerPath)
		routerCompBytes, _ := ioutil.ReadAll(routerComp)
		componentStr = string(routerCompBytes)
	}
	// Compile router separately since it's ejected from core.
	if err = (compileSvelte(ctx, SSRctx, "ejected/router.svelte", componentStr, buildPath+"/spa/ejected/router.js", stylePath, tempBuildDir)); err != nil {
		return err
	}

	// Go through all file paths in the "/layout" folder.
	err = filepath.Walk(tempBuildDir+"layout", func(layoutPath string, layoutFileInfo os.FileInfo, err error) error {

		if err != nil {
			return fmt.Errorf("can't stat %s: %w", layoutPath, err)
		}
		// Create destination path.
		destFile := buildPath + "/spa" + strings.TrimPrefix(layoutPath, tempBuildDir+"layout")
		// Make sure path is a directory
		if layoutFileInfo.IsDir() {
			// Create any sub directories need for filepath.
			if err = os.MkdirAll(destFile, os.ModePerm); err != nil {
				return fmt.Errorf("can't make path: %s %w%s", layoutPath, err, common.Caller())
			}
		} else {
			// If the file is in .svelte format, compile it to .js
			if filepath.Ext(layoutPath) == ".svelte" {

				// Replace .svelte file extension with .js.
				destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".js"

				component, err := ioutil.ReadFile(layoutPath)
				if err != nil {
					return fmt.Errorf("can't read component file: %s %w%s", layoutPath, err, common.Caller())
				}
				componentStr := string(component)
				if err = compileSvelte(ctx, SSRctx, layoutPath, componentStr, destFile, stylePath, tempBuildDir); err != nil {
					return fmt.Errorf("%w%s", err, common.Caller())
				}

				// Remove temporary theme build directory.
				destLayoutPath := strings.TrimPrefix(layoutPath, tempBuildDir)
				// Create entry for layout.js.
				layoutSignature := strings.ReplaceAll(strings.ReplaceAll((destLayoutPath), "/", "_"), ".", "_")
				// Remove layout directory.
				destLayoutPath = strings.TrimPrefix(destLayoutPath, "layout/")
				// Compose entry for layout.js file.
				allComponentsStr = allComponentsStr + "export {default as " + layoutSignature + "} from '../" + destLayoutPath + "';\n"

				compiledComponentCounter++

			}
		}
		return nil
	})
	// TODO: return file names here amd anywhere possible
	if err != nil {
		return err

	}

	// Write layout.js to filesystem.
	err = ioutil.WriteFile(buildPath+"/spa/ejected/layout.js", []byte(allComponentsStr), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Unable to write layout.js file: %w%s", err, common.Caller())

	}

	Log("Number of components compiled: " + strconv.Itoa(compiledComponentCounter))
	return nil
}

func compileSvelte(ctx *v8go.Context, SSRctx *v8go.Context, layoutPath string,
	componentStr string, destFile string, stylePath string, tempBuildDir string) error {

	// Compile component with Svelte.
	_, err := ctx.RunScript("var { js, css } = svelte.compile(`"+componentStr+"`, {css: false, hydratable: true});", "compile_svelte")
	if err != nil {
		return fmt.Errorf("can't compile component file %s with Svelte: %w%s", layoutPath, err, common.Caller())
	}
	// Get the JS code from the compiled result.
	jsCode, err := ctx.RunScript("js.code;", "compile_svelte")
	if err != nil {
		return fmt.Errorf("V8go could not execute js.code for %s: %w%s", layoutPath, err, common.Caller())
	}
	jsBytes := []byte(jsCode.String())
	err = ioutil.WriteFile(destFile, jsBytes, 0755)
	if err != nil {
		return fmt.Errorf("Unable to write compiled client file for %s: %w%s", layoutPath, err, common.Caller())
	}

	// Get the CSS code from the compiled result.
	cssCode, err := ctx.RunScript("css.code;", "compile_svelte")
	if err != nil {
		return fmt.Errorf("V8go could not execute css.code  for %s: %w%s", layoutPath, err, common.Caller())
	}
	cssStr := strings.TrimSpace(cssCode.String())
	// If there is CSS, write it into the bundle.css file.
	if cssStr != "null" {
		cssFile, err := os.OpenFile(stylePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("Could not open bundle.css for writing %s: %w%s", layoutPath, err, common.Caller())
		}
		defer cssFile.Close()
		if _, err := cssFile.WriteString(cssStr); err != nil {
			return fmt.Errorf("could not write to cssStr for %s: %w%s", layoutPath, err, common.Caller())
		}
	}

	// Get Server Side Rendered (SSR) JS.
	_, ssrCompileErr := ctx.RunScript("var { js: ssrJs, css: ssrCss } = svelte.compile(`"+componentStr+"`, {generate: 'ssr'});", "compile_svelte")
	if ssrCompileErr != nil {
		return fmt.Errorf("V8go could not compile ssrJs.code for %s: %w%s", layoutPath, ssrCompileErr, common.Caller())
	}
	ssrJsCode, err := ctx.RunScript("ssrJs.code;", "compile_svelte")
	if err != nil {
		return fmt.Errorf("V8go could not get ssrJs.code value for %s: %w%s", layoutPath, err, common.Caller())
	}
	// Regex match static import statements.
	reStaticImport := regexp.MustCompile(`import\s((.*)\sfrom(.*);|(((.*)\n){0,})\}\sfrom(.*);)`)
	// Regex match static export statements.
	reStaticExport := regexp.MustCompile(`export\s(.*);`)
	// Remove static import statements.
	ssrStr := reStaticImport.ReplaceAllString(ssrJsCode.String(), `/*$0*/`)
	// Remove static export statements.
	ssrStr = reStaticExport.ReplaceAllString(ssrStr, `/*$0*/`)
	// Use var instead of const so it can be redeclared multiple times.
	reConst := regexp.MustCompile(`(?m)^const\s`)
	ssrStr = reConst.ReplaceAllString(ssrStr, "var ")
	// Remove temporary theme directory info from path before making a comp signature.
	layoutPath = strings.TrimPrefix(layoutPath, tempBuildDir)
	// Create custom variable name for component based on the file path for the layout.
	componentSignature := strings.ReplaceAll(strings.ReplaceAll(layoutPath, "/", "_"), ".", "_")
	// Use signature instead of generic "Component". Add space to avoid also replacing part of "loadComponent".
	ssrStr = strings.ReplaceAll(ssrStr, " Component ", " "+componentSignature+" ")

	namedExports := reStaticExport.FindAllStringSubmatch(ssrStr, -1)
	// Loop through all export statements.
	for _, namedExport := range namedExports {
		// Get exported functions that aren't default.
		// Ignore names that contain semicolons to try to avoid pulling in CSS map code: https://github.com/sveltejs/svelte/issues/3604
		if !strings.HasPrefix(namedExport[1], "default ") && !strings.Contains(namedExport[1], ";") {
			// Get just the name(s) inside the curly brackets
			exportNames := makeNameList(namedExport)
			for _, exportName := range exportNames {
				if exportName != "" && componentSignature != "" {
					// Create new component signature with variable name appended to the end.
					ssrStr = strings.ReplaceAll(ssrStr, exportName, componentSignature+"_"+exportName)
				}
			}
		}
	}

	// Replace import references with variable signatures.
	reStaticImportPath := regexp.MustCompile(`(?:'|").*(?:'|")`)
	reStaticImportName := regexp.MustCompile(`import\s(.*)\sfrom`)
	namedImports := reStaticImport.FindAllString(ssrStr, -1)
	for _, namedImport := range namedImports {
		// Get path only from static import statement.
		importPath := reStaticImportPath.FindString(namedImport)
		importNameSlice := reStaticImportName.FindStringSubmatch(namedImport)
		importNameStr := ""
		var namedImportNameStrs []string
		if len(importNameSlice) > 0 {
			importNameStr = importNameSlice[1]
			// Check if it's a named import (starts w curly bracket)
			// and import path should not have spaces (ignores CSS mapping: https://github.com/sveltejs/svelte/issues/3604).
			if strings.Contains(importNameSlice[1], "{") && !strings.Contains(importPath, " ") {
				namedImportNameStrs = makeNameList(importNameSlice)
			}
		}
		// Remove quotes around path.
		importPath = strings.Trim(importPath, `'"`)
		// Get individual path arguments.
		layoutParts := strings.Split(layoutPath, "/")
		layoutFileName := layoutParts[len(layoutParts)-1]
		layoutRootPath := strings.TrimSuffix(layoutPath, layoutFileName)

		importParts := strings.Split(importPath, "/")
		// Initialize the import signature that will be used for unique variable names.
		importSignature := ""
		// Check if the path ends with a file extension, e.g. ".svelte".
		if len(filepath.Ext(importParts[len(importParts)-1])) > 1 {
			for _, importPart := range importParts {
				// Check if path starts relative to current folder.
				if importPart == "." {
					// Remove the proceeding dot so the file can be combined with the root.
					importPath = strings.TrimPrefix(importPath, "./")
				}
				// Check if path goes up a folder.
				if importPart == ".." {
					// Remove the proceeding double dots so it can be combined with root.
					importPath = strings.TrimPrefix(importPath, importPart+"/")
					// Split the layout root path so we can remove the last segment since the double dots indicates going back a folder.
					layoutParts = strings.Split(layoutRootPath, "/")
					layoutRootPath = strings.TrimSuffix(layoutRootPath, layoutParts[len(layoutParts)-2]+"/")
				}
			}
			// Create the variable name from the full path.
			importSignature = strings.ReplaceAll(strings.ReplaceAll((layoutRootPath+importPath), "/", "_"), ".", "_")
		}
		// TODO: Add an else ^ to account for NPM dependencies?

		// Check that there is a valid import to replace.
		if importNameStr != "" && importSignature != "" {
			// Only use comp signatures inside JS template literal placeholders.
			reTemplatePlaceholder := regexp.MustCompile(`(?s)\$\{validate_component\(.*\)\}`)
			// Only replace this specific variable, so not anything that has letters, underscores, or numbers attached to it.
			reImportNameUse := regexp.MustCompile(`([^a-zA-Z_0-9])` + importNameStr + `([^a-zA-Z_0-9])`)
			// Find the template placeholders.
			ssrStr = reTemplatePlaceholder.ReplaceAllStringFunc(ssrStr,
				func(placeholder string) string {
					// Use the signature instead of variable name.
					return reImportNameUse.ReplaceAllString(placeholder, "${1}"+importSignature+"${2}")
				},
			)
		}

		// Handle each named import, e.g. import { first, second } from "./whatever.svelte".
		for _, currentNamedImport := range namedImportNameStrs {
			// Remove whitespace on sides that might occur when splitting into array by comma.
			currentNamedImport = strings.TrimSpace(currentNamedImport)
			// Check that there is a valid named import.
			if currentNamedImport != "" && importSignature != "" {
				// Only add named imports to create_ssr_component().
				reCreateFunc := regexp.MustCompile(`(create_ssr_component\(\(.*\)\s=>\s\{)`)
				// Entry should be block scoped, like: let count = layout_scripts_stores_svelte_count;
				blockScopedVar := "\n let " + currentNamedImport + " = " + importSignature + "_" + currentNamedImport + ";"
				// Add block scoped var inside create_ssr_component.
				ssrStr = reCreateFunc.ReplaceAllString(ssrStr, "${1}"+blockScopedVar)
			}
		}
	}

	// Remove allComponents object (leaving just componentSignature) for SSR.
	// Match: allComponents.layout_components_grid_svelte
	reAllComponentsDot := regexp.MustCompile(`allComponents\.(layout_.*_svelte)`)
	ssrStr = reAllComponentsDot.ReplaceAllString(ssrStr, "${1}")
	// Match: allComponents[component]
	reAllComponentsBracket := regexp.MustCompile(`allComponents\[(.*)\]`)
	ssrStr = reAllComponentsBracket.ReplaceAllString(ssrStr, "globalThis[${1}]")
	// Match: allComponents["layout_components_decrementer_svelte"]
	reAllComponentsBracketStr := regexp.MustCompile(`allComponents\[\"(.*)\"\]`)
	ssrStr = reAllComponentsBracketStr.ReplaceAllString(ssrStr, "${1}")

	paginatedContent, _ := getPagination()
	for _, pager := range paginatedContent {
		if "layout_content_"+pager.contentType+"_svelte" == componentSignature {
			for _, paginationVar := range pager.paginationVars {
				// Prefix var so it doesn't conflict with other variables.
				globalVar := "plenti_global_pager_" + paginationVar
				// Initialize var outside of function to set it as global.
				ssrStr = "var " + globalVar + ";\n" + ssrStr
				// Match where the pager var is set, like: let totalPages = Math.ceil(totalPosts / postsPerPage);
				reLocalVar := regexp.MustCompile(`((let\s|const\s|var\s)` + paginationVar + `.*;)`)
				// Create statement to assign local var to global var.
				makeGlobalVar := globalVar + " = " + paginationVar + ";"
				// Assign value to global var inside create_ssr_component() func, like: plenti_global_pager_totalPages = totalPages;
				ssrStr = reLocalVar.ReplaceAllString(ssrStr, "${1}\n"+makeGlobalVar)
				// Clear out styles for SSR since they are already pulled from client components.
				ssrStr = removeCSS(ssrStr)
			}
		}
	}

	// Add component to context so it can be used to render HTML in data_source.go.
	_, err = SSRctx.RunScript(ssrStr, "create_ssr")
	if err != nil {
		return fmt.Errorf("Could not add SSR Component for %s: %w%s", layoutPath, err, common.Caller())
	}

	return nil
}

func removeCSS(str string) string {
	// Match var css = { ... }
	reCSS := regexp.MustCompile(`var(\s)css(\s)=(\s)\{(.*\n){0,}\};`)
	// Delete these styles because they often break pagination SSR.
	return reCSS.ReplaceAllString(str, "")
}

func makeNameList(importNameSlice []string) []string {
	var namedImportNameStrs []string
	// Get just the name(s) of the variable(s).
	namedImportNameStr := strings.Trim(importNameSlice[1], "{ }")
	// Chech if there are multiple names separated by a comma.
	if strings.Contains(namedImportNameStr, ",") {
		// Break apart by comma and add to individual items to array.
		namedImportNameStrs = append(namedImportNameStrs, strings.Split(namedImportNameStr, ",")...)
		for i := range namedImportNameStrs {
			// Remove surrounding whitespace (this will be present if there was space after the comma).
			namedImportNameStrs[i] = strings.TrimSpace(namedImportNameStrs[i])
		}
	} else {
		// Only one name was used, so add it directly to the array.
		namedImportNameStrs = append(namedImportNameStrs, namedImportNameStr)
	}
	return namedImportNameStrs
}

type pager struct {
	contentType    string
	contentPath    string
	paginationVars []string
}

func getPagination() ([]pager, *regexp.Regexp) {
	// Get settings from config file.
	siteConfig, _ := readers.GetSiteConfig(".")
	// Setup regex to find pagination.
	rePaginate := regexp.MustCompile(`:paginate\((.*?)\)`)
	// Initialize new pager struct
	var pagers []pager
	// Check for pagination in plenti.json config file.
	for configContentType, slug := range siteConfig.Types {
		// Initialize list of all :paginate() vars in a given slug.
		replacements := []string{}
		// Find every instance of :paginate() in the slug.
		paginateReplacements := rePaginate.FindAllStringSubmatch(slug, -1)
		// Loop through all :paginate() replacements found in config file.
		for _, replacement := range paginateReplacements {
			// Add the variable name defined within the brackets to the list.
			replacements = append(replacements, replacement[1])
		}
		var pager pager
		pager.contentType = configContentType
		pager.contentPath = slug
		pager.paginationVars = replacements
		pagers = append(pagers, pager)
	}
	return pagers, rePaginate
}
