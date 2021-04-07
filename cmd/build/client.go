package build

import (
	"embed"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
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
	if common.UseMemFS {
		// clear styles as we append bytes.
		common.Set(stylePath, "", &common.FData{})
	}
	layoutPath := buildPath + "/spa/ejected/layouts.js"

	// Initialize string for layouts.js component list.
	var allLayoutsStr string

	// Set up counter for logging output.
	compiledComponentCounter := 0

	// Get svelte compiler code from node_modules.
	compiler, err := ioutil.ReadFile(
		fmt.Sprintf("%snode_modules/svelte/compiler.js", tempBuildDir),
	)
	if err != nil {
		return fmt.Errorf("Can't read %s/node_modules/svelte/compiler.js: %w%s\n", tempBuildDir, err, common.Caller())

	}
	// Remove reference to 'self' that breaks v8go on line 19 of node_modules/svelte/compiler.js.
	compilerStr := strings.Replace(string(compiler), "self.performance.now();", "'';", 1)
	// Remove 'require' that breaks v8go on line 22647 of node_modules/svelte/compiler.js.
	compilerStr = strings.Replace(compilerStr, "const Url$1 = (typeof URL !== 'undefined' ? URL : require('url').URL);", "", 1)
	ctx, err := v8go.NewContext(nil)
	if err != nil {
		return fmt.Errorf("Could not create Isolate: %w%s\n", err, common.Caller())

	}
	_, err = ctx.RunScript(compilerStr, "compile_svelte")
	if err != nil {
		return fmt.Errorf("Could not add svelte compiler: %w%s\n", err, common.Caller())

	}

	SSRctx, err = v8go.NewContext(nil)
	if err != nil {
		return fmt.Errorf("Could not create Isolate: %w%s\n", err, common.Caller())

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
			return fmt.Errorf("Can't read %s: %w%s\n", svelteLib, err, common.Caller())

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
		// 	fmt.Println(fmt.Errorf("Could not add create_ssr_component() func from svelte/internal for file %s: %w%s\n", svelteLib, err, common.Caller()))

		// }

	}

	routerPath := "ejected/router.svelte"
	ejectedRouterPath := tempBuildDir + routerPath
	var componentStr string
	if _, err := os.Stat(ejectedRouterPath); err == nil {
		// The router has been ejected to the filesystem.
		component, err := ioutil.ReadFile(ejectedRouterPath)
		if err != nil {
			return fmt.Errorf("can't read component file: %s %w%s\n", ejectedRouterPath, err, common.Caller())
		}
		componentStr = string(component)
	} else if os.IsNotExist(err) {
		// The router has not been ejected, use the embedded defaults.
		ejected, err := fs.Sub(defaultsEjectedFS, "defaults")
		if err != nil {
			common.CheckErr(fmt.Errorf("Unable to get ejected defaults: %w", err))
		}
		routerComp, err := ejected.Open(routerPath)
		if err != nil {
			log.Fatal(err)
		}
		routerCompBytes, err := ioutil.ReadAll(routerComp)
		if err != nil {
			log.Fatal(err)
		}
		componentStr = string(routerCompBytes)
	}
	// Compile router separately since it's ejected from core.
	if err = (compileSvelte(ctx, SSRctx, routerPath, componentStr, buildPath+"/spa/ejected/router.js", stylePath, tempBuildDir)); err != nil {
		return err
	}

	// Go through all file paths in the "/layouts" folder.
	err = filepath.WalkDir(tempBuildDir+"layouts", func(layoutPath string, layoutFileInfo fs.DirEntry, err error) error {

		if err != nil {
			return fmt.Errorf("can't stat %s: %w", layoutPath, err)
		}
		// Create destination path.
		destFile := buildPath + "/spa" + strings.TrimPrefix(layoutPath, tempBuildDir+"layouts")
		// Make sure path is a directory
		if layoutFileInfo.IsDir() {
			// Create any sub directories need for filepath.
			if err = os.MkdirAll(destFile, os.ModePerm); err != nil {
				return fmt.Errorf("can't make path: %s %w%s\n", layoutPath, err, common.Caller())
			}
		} else {
			// If the file is in .svelte format, compile it to .js
			if filepath.Ext(layoutPath) == ".svelte" {

				// Replace .svelte file extension with .js.
				destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".js"

				component, err := ioutil.ReadFile(layoutPath)
				if err != nil {
					return fmt.Errorf("can't read component file: %s %w%s\n", layoutPath, err, common.Caller())
				}
				componentStr := string(component)

				if err = compileSvelte(ctx, SSRctx, layoutPath, componentStr, destFile, stylePath, tempBuildDir); err != nil {
					return fmt.Errorf("%w%s\n", err, common.Caller())
				}

				// Remove temporary theme build directory.
				destLayoutPath := strings.TrimPrefix(layoutPath, tempBuildDir)
				// Create entry for layouts.js.
				layoutSignature := strings.ReplaceAll(strings.ReplaceAll((destLayoutPath), "/", "_"), ".", "_")
				// Remove layouts directory.
				destLayoutPath = strings.TrimPrefix(destLayoutPath, "layouts/")
				// Compose entry for layouts.js file.
				allLayoutsStr = allLayoutsStr + "export {default as " + layoutSignature + "} from '../" + destLayoutPath + "';\n"

				compiledComponentCounter++

			}
		}
		return nil
	})
	// TODO: return file names here amd anywhere possible
	if err != nil {
		return err

	}

	if common.UseMemFS {
		b := []byte(allLayoutsStr)
		common.Set(layoutPath, "", &common.FData{B: b})
		return nil

	}
	// Write layouts.js to filesystem.
	err = ioutil.WriteFile(layoutPath, []byte(allLayoutsStr), os.ModePerm)
	if err != nil {
		return fmt.Errorf("Unable to write layouts.js file: %w%s\n", err, common.Caller())

	}

	Log("Number of components compiled: " + strconv.Itoa(compiledComponentCounter))
	return nil
}

func removeCSS(str string) string {

	// Delete these styles because they often break pagination SSR.
	return reCSSCli.ReplaceAllString(str, "")
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
	for configContentType, slug := range siteConfig.Routes {
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
