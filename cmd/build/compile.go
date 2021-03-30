package build

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/plentico/plenti/common"
	"rogchap.com/v8go"
)

var (
	// Regex match static import statements.
	reStaticImport = regexp.MustCompile(`import\s((.*)\sfrom(.*);|(((.*)\n){0,})\}\sfrom(.*);)`)

	// Regex match static export statements.
	reStaticExport = regexp.MustCompile(`export\s(.*);`)
	// Replace import references with variable signatures.
	reStaticImportPath = regexp.MustCompile(`(?:'|").*(?:'|")`)
	reStaticImportName = regexp.MustCompile(`import\s(.*?)\sfrom`)
	// Use var instead of const so it can be redeclared multiple times.
	reConst = regexp.MustCompile(`(?m)^const\s`)
	// Only use comp signatures inside JS template literal placeholders.
	reTemplatePlaceholder = regexp.MustCompile(`(?s)\$\{validate_component\(.*\)\}`)
	// Only add named imports to create_ssr_component().
	reCreateFunc = regexp.MustCompile(`(create_ssr_component\(\(.*\)\s=>\s\{)`)
	// Match: allLayouts.layouts_components_grid_svelte
	reAllLayoutsDot = regexp.MustCompile(`allLayouts\.(layouts_.*_svelte)`)
	// Match: allLayouts[component]
	reAllLayoutsBracket = regexp.MustCompile(`allLayouts\[(.*)\]`)
	// Match: allLayouts["layouts_components_decrementer_svelte"]
	reAllLayoutsBracketStr = regexp.MustCompile(`allLayouts\[\"(.*)\"\]`)
	// Match var css = { ... }
	reCSSCli = regexp.MustCompile(`var(\s)css(\s)=(\s)\{(.*\n){0,}\};`)
)

func compileSvelte(ctx *v8go.Context, SSRctx *v8go.Context, layoutPath string,
	componentStr string, destFile string, stylePath string, tempBuildDir string) error {

	layoutFD := common.GetOrSet(layoutPath)

	// will break router if no layout check and can't use HasPrefix as themes breaks that
	if common.UseMemFS {
		if strings.Contains(layoutPath, "layouts/") {
			// if hash is the same skip. common.Get(layoutPath).Hash will be 0 initially
			if layoutFD.Hash > 0 && layoutFD.Hash == common.CRC32Hasher([]byte(componentStr)) {
				// Add the orig css as we clear the bundle each build.
				//  May be a better way to avoid also but cheap enough vs compiling anyway.
				val := common.Get(stylePath)
				// layoutFD.CSS is the specific component css already compiled
				// .B may seem counterintuitive for style but we use .B for all in server.
				val.B = append(val.B, layoutFD.CSS...)
				// need this or complains about missing layout_xxxxx. Some way around?
				_, err := SSRctx.RunScript(string(layoutFD.SSR), "create_ssr")
				if err != nil {
					return fmt.Errorf("Could not add SSR Component: %w%s\n", err, common.Caller())
				}
				/// we don't go further so set js etc.. as processed
				// important for not doing work for no reason and stops regex import logc breaking as paths will have already be changed...
				common.Get(destFile).Processed = true

				return nil
			}

		}
		// add or update hash to other also, not just layouts
		layoutFD.Hash = common.CRC32Hasher([]byte(componentStr))

	}

	// Compile component with Svelte.
	_, err := ctx.RunScript("var { js, css } = svelte.compile(`"+componentStr+"`, {css: false, hydratable: true});", "compile_svelte")
	if err != nil {
		return fmt.Errorf("can't compile component file %s with Svelte: %w%s\n", layoutPath, err, common.Caller())
	}
	// Get the JS code from the compiled result.
	jsCode, err := ctx.RunScript("js.code;", "compile_svelte")
	if err != nil {
		return fmt.Errorf("V8go could not execute js.code for %s: %w%s\n", layoutPath, err, common.Caller())
	}
	jsBytes := []byte(jsCode.String())
	if common.UseMemFS {
		common.Set(destFile, layoutPath, &common.FData{Hash: common.CRC32Hasher(jsBytes), B: jsBytes})

	} else {
		err = os.WriteFile(destFile, jsBytes, 0755)
		if err != nil {
			return fmt.Errorf("Unable to write compiled client file: %w%s\n", err, common.Caller())
		}
	}

	// Get the CSS code from the compiled result.
	cssCode, err := ctx.RunScript("css.code;", "compile_svelte")
	if err != nil {
		return fmt.Errorf("V8go could not execute css.code  for %s: %w%s\n", layoutPath, err, common.Caller())
	}
	cssStr := strings.TrimSpace(cssCode.String())
	// If there is CSS, write it into the bundle.css file.
	if cssStr != "null" {
		if common.UseMemFS {
			// ok to append as created on build
			val := common.Get(stylePath)

			val.B = append(val.B, []byte(cssCode.String())...) // will reuse just layout/component css when no change
			// could use pointers but this is ok
			layoutFD.CSS = []byte(cssCode.String())

		} else {
			cssFile, err := os.OpenFile(stylePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("Could not open bundle.css for writing: %w%s\n", err, common.Caller())
			}
			defer cssFile.Close()
			if _, err := cssFile.WriteString(cssStr); err != nil {
				return fmt.Errorf("could not write to cssStr: %w%s\n", err, common.Caller())
			}
		}
	}

	// Get Server Side Rendered (SSR) JS.
	_, ssrCompileErr := ctx.RunScript("var { js: ssrJs, css: ssrCss } = svelte.compile(`"+componentStr+"`, {generate: 'ssr'});", "compile_svelte")
	if ssrCompileErr != nil {
		return fmt.Errorf("V8go could not compile ssrJs.code for %s: %w%s\n", layoutPath, ssrCompileErr, common.Caller())
	}
	ssrJsCode, err := ctx.RunScript("ssrJs.code;", "compile_svelte")
	if err != nil {
		return fmt.Errorf("V8go could not get ssrJs.code value for %s: %w%s\n", layoutPath, err, common.Caller())
	}

	// Remove static import statements.
	ssrStr := reStaticImport.ReplaceAllString(ssrJsCode.String(), `/*$0*/`)
	// Remove static export statements.
	ssrStr = reStaticExport.ReplaceAllString(ssrStr, `/*$0*/`)

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

				// Entry should be block scoped, like: let count = layouts_scripts_stores_svelte_count;
				blockScopedVar := "\n let " + currentNamedImport + " = " + importSignature + "_" + currentNamedImport + ";"
				// Add block scoped var inside create_ssr_component.
				ssrStr = reCreateFunc.ReplaceAllString(ssrStr, "${1}"+blockScopedVar)
			}
		}
	}

	// Remove allLayouts object (leaving just componentSignature) for SSR.
	// Match: allLayouts.layouts_components_grid_svelte
	ssrStr = reAllLayoutsDot.ReplaceAllString(ssrStr, "${1}")
	// Match: allLayouts[component]
	ssrStr = reAllLayoutsBracket.ReplaceAllString(ssrStr, "globalThis[${1}]")
	// Match: allLayouts["layouts_components_decrementer_svelte"]
	ssrStr = reAllLayoutsBracketStr.ReplaceAllString(ssrStr, "${1}")

	paginatedContent, _ := getPagination()
	for _, pager := range paginatedContent {
		if "layouts_content_"+pager.contentType+"_svelte" == componentSignature {
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
		return fmt.Errorf("Could not add SSR Component for %s: %w%s\n", layoutPath, err, common.Caller())
	}
	// again store for no change
	layoutFD.SSR = []byte(ssrStr)
	return nil
}
