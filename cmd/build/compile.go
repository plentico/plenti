package build

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	// Match: var Compname = create_ssr_component(
	reSSRComp = regexp.MustCompile(`(var\s)[A-Za-z0-9_-]*(\s=\screate_ssr_component\()`)
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

type OutputCode struct {
	JS  string
	CSS string
}

func compileSvelte(ctx *v8go.Context, SSRctx *v8go.Context, layoutPath string,
	componentStr string, destFile string, stylePath string) error {

	// Create any sub directories need for filepath.
	if err := os.MkdirAll(filepath.Dir(destFile), os.ModePerm); err != nil {
		return fmt.Errorf("can't make path: %s %w\n", layoutPath, err)
	}

	// Compile component with Svelte.
	scriptDOM := fmt.Sprintf(`;__svelte__.compile({ "path": %q, "code": %q, "target": "dom", "css": false, hydratable: true })`, layoutPath, componentStr)
	resultDOM, err := ctx.RunScript(scriptDOM, "compile_svelte")
	if err != nil {
		return fmt.Errorf("\nDOM: Can't compile component file %s\n%w", layoutPath, err)
	}
	// Get the JS code from the compiled result.
	outDOM := new(OutputCode)
	if err := json.Unmarshal([]byte(resultDOM.String()), outDOM); err != nil {
		return fmt.Errorf("Could not unmarshal DOM output: %w\n", err)
	}
	jsCode := outDOM.JS
	jsBytes := []byte(jsCode)
	err = os.WriteFile(destFile, jsBytes, 0755)
	if err != nil {
		return fmt.Errorf("Unable to write compiled client file: %w\n", err)
	}

	// Get the CSS code from the compiled result.
	cssCode := outDOM.CSS
	cssStr := strings.TrimSpace(cssCode)
	// If there is CSS, write it into the bundle.css file.
	if cssStr != "null" {
		cssFile, err := os.OpenFile(stylePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("Could not open bundle.css for writing: %w\n", err)
		}
		defer cssFile.Close()
		if _, err := cssFile.WriteString(cssStr); err != nil {
			return fmt.Errorf("could not write to cssStr: %w\n", err)
		}
	}

	// Get Server Side Rendered (SSR) JS.
	scriptSSR := fmt.Sprintf(`;__svelte__.compile({ "path": %q, "code": %q, "target": "ssr", "css": false })`, layoutPath, componentStr)
	resultSSR, err := ctx.RunScript(scriptSSR, "compile_svelte")
	if err != nil {
		return fmt.Errorf("\nSSR: Can't compile component file %s\n%w", layoutPath, err)
	}
	outSSR := new(OutputCode)
	if err := json.Unmarshal([]byte(resultSSR.String()), outSSR); err != nil {
		return fmt.Errorf("Could not unmarshal SSR output: %w\n", err)
	}
	ssrJsCode := outSSR.JS

	// Remove static import statements.
	ssrStr := reStaticImport.ReplaceAllString(ssrJsCode, `/*$0*/`)
	// Remove static export statements.
	ssrStr = reStaticExport.ReplaceAllString(ssrStr, `/*$0*/`)

	ssrStr = reConst.ReplaceAllString(ssrStr, "var ")
	// Create custom variable name for component based on the file path for the layout.
	componentSignature := strings.ReplaceAll(strings.ReplaceAll(layoutPath, "/", "_"), ".", "_")
	// Use signature instead of specific component name (e.g. var Html = create_ssr_component(($$result, $$props, $$bindings, slots) => {)
	ssrStr = reSSRComp.ReplaceAllString(ssrStr, "${1}"+componentSignature+"${2}")

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
			if currentNamedImport != "" && importSignature != "" && strings.HasSuffix(importSignature, "_svelte") {
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
		return fmt.Errorf("Could not add SSR Component for %s: %w\n", layoutPath, err)
	}

	return nil
}
