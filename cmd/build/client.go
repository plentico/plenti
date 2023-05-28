package build

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/plentico/plenti/readers"
	"github.com/spf13/afero"

	"rogchap.com/v8go"
)

// Global var for virtual SSR components
var SSRFs afero.Fs

// Client builds the SPA.
func Client(buildPath string, coreFS embed.FS, compilerFS embed.FS) error {

	defer Benchmark(time.Now(), "Compiling client SPA with Svelte")

	Log("\nCompiling client SPA with svelte")

	// Initialize var for virtual SSR components
	SSRFs = afero.NewMemMapFs()

	stylePath := buildPath + "/spa/bundle.css"
	allLayoutsPath := buildPath + "/spa/generated/layouts.js"
	// Initialize string for layouts.js component list.
	var allLayoutsStr string

	// Set up counter for logging output.
	compiledComponentCounter := 0

	// Get transformed svelte compiler code from embedded filesystem.
	compiler, err := compilerFS.ReadFile("compiler/compiler.js")
	if err != nil {
		return err
	}
	compilerStr := string(compiler)
	ctx := v8go.NewContext(nil)
	_, err = ctx.RunScript(compilerStr, "compile_svelte")
	if err != nil {
		return fmt.Errorf("Could not add svelte compiler: %w\n", err)

	}

	// Compile Svelte components from ejectable core
	fs.WalkDir(coreFS, "core", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Don't compile directories or non svelte files (.js files)
		if d.IsDir() || !strings.HasSuffix(path, ".svelte") {
			return nil
		}
		// Initialize var to hold actual component file contents
		var componentStr string
		// Check if the path has been ejected to the filesystem.
		_, err = os.Stat(path)
		// Check if the path has been ejected to the virtual filesystem for a theme build.
		if ThemeFs != nil {
			_, err = ThemeFs.Stat(path)
		}
		if err == nil {
			// The file has been ejected to the filesystem.
			component, err := getVirtualFileIfThemeBuild(path)
			if err != nil {
				return fmt.Errorf("can't read component file: %s %w\n", path, err)
			}
			componentStr = string(component)
		} else if os.IsNotExist(err) {
			// The file has not been ejected, use the embedded defaults.
			nonEjectedFS, err := fs.Sub(coreFS, ".")
			if err != nil {
				log.Fatal("Unable to get non ejected defaults: %w", err)
			}
			component, err := nonEjectedFS.Open(path)
			if err != nil {
				fmt.Printf("Could not open '%s' embeded file: %s", path, err)
			}
			componentBytes, err := ioutil.ReadAll(component)
			if err != nil {
				fmt.Printf("Could not read '%s' embeded file: %s", path, err)
			}
			componentStr = string(componentBytes)
		}
		destPath := buildPath + "/spa/" + strings.TrimSuffix(path, ".svelte") + ".js"
		err = (compileSvelte(ctx, path, componentStr, destPath, stylePath))
		if err != nil {
			fmt.Printf("Could not compile '%s' Svelte component: %s", path, err)
		}
		return nil
	})

	// Check if using a theme
	if ThemeFs != nil {
		// A theme is being used, so compile the files from the virtual fs
		if err := afero.Walk(ThemeFs, "layouts", func(layoutPath string, layoutFileInfo os.FileInfo, err error) error {
			if layoutFileInfo.IsDir() {
				return nil
			}
			err = copyNonSvelteFiles(layoutPath, buildPath)
			if err != nil {
				return err
			}
			compiledComponentCounter, allLayoutsStr, err = compileComponent(err, layoutPath, layoutFileInfo, buildPath, ctx, stylePath, allLayoutsStr, compiledComponentCounter)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			return fmt.Errorf("\nCould not get layout from virtual theme build %w", err)
		}
	} else {
		// A theme is NOT being used, so compile the components from the root project
		if err := filepath.Walk("layouts", func(layoutPath string, layoutFileInfo os.FileInfo, err error) error {
			if layoutFileInfo.IsDir() {
				return nil
			}
			err = copyNonSvelteFiles(layoutPath, buildPath)
			if err != nil {
				return err
			}
			compiledComponentCounter, allLayoutsStr, err = compileComponent(err, layoutPath, layoutFileInfo, buildPath, ctx, stylePath, allLayoutsStr, compiledComponentCounter)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			return fmt.Errorf("\nCould not get all layouts %w", err)
		}
	}

	// Write layouts.js to filesystem.
	err = ioutil.WriteFile(allLayoutsPath, []byte(allLayoutsStr), os.ModePerm)
	if err != nil {
		return fmt.Errorf("\nUnable to write layouts.js file: %w\n", err)
	}

	Log("Number of components compiled: " + strconv.Itoa(compiledComponentCounter))
	return nil
}

func copyNonSvelteFiles(layoutPath string, buildPath string) error {
	if filepath.Ext(layoutPath) != ".svelte" {
		from, err := os.Open(layoutPath)
		if err != nil {
			return fmt.Errorf("Could not open non-svelte layout %s for copying: %w\n", layoutPath, err)
		}
		defer from.Close()

		destPath := buildPath + "/spa/" + layoutPath
		// Create any sub directories need for filepath.
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return fmt.Errorf("can't make folders for '%s': %w\n", destPath, err)
		}

		to, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("Could not create non-svelte layout destination %s for copying: %w\n", destPath, err)
		}
		defer to.Close()

		_, err = io.Copy(to, from)
		if err != nil {
			return fmt.Errorf("Could not copy non-svelte layout from source %s to destination: %w\n", layoutPath, err)
		}
	}
	return nil
}

func compileComponent(err error, layoutPath string, layoutFileInfo os.FileInfo, buildPath string, ctx *v8go.Context, stylePath string, allLayoutsStr string, compiledComponentCounter int) (int, string, error) {
	if err != nil {
		return compiledComponentCounter, allLayoutsStr, fmt.Errorf("can't stat %s: %w", layoutPath, err)
	}
	// Create destination path.
	destFile := buildPath + "/spa/" + layoutPath
	// If the file is in .svelte format, compile it to .js
	if filepath.Ext(layoutPath) == ".svelte" {
		// Replace .svelte file extension with .js.
		destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".js"
		// Get component file contents
		component, err := getVirtualFileIfThemeBuild(layoutPath)
		if err != nil {
			return compiledComponentCounter, allLayoutsStr, fmt.Errorf("can't read component file: %s %w\n", layoutPath, err)
		}
		componentStr := string(component)
		// Actually compile component
		if err = compileSvelte(ctx, layoutPath, componentStr, destFile, stylePath); err != nil {
			return compiledComponentCounter, allLayoutsStr, fmt.Errorf("%w\n", err)
		}
		// Create entry for layouts.js.
		layoutSignature := strings.ReplaceAll(strings.ReplaceAll((layoutPath), "/", "_"), ".", "_")
		// Compose entry for layouts.js file.
		allLayoutsStr = allLayoutsStr + "export {default as " + layoutSignature + "} from '../" + layoutPath + "';\n"
		// Increment counter for each compiled component.
		compiledComponentCounter++
	}
	return compiledComponentCounter, allLayoutsStr, nil
}

func getVirtualFileIfThemeBuild(filename string) ([]byte, error) {
	var fileContents []byte
	var err error
	if ThemeFs != nil {
		fileContents, err = afero.ReadFile(ThemeFs, filename)
		if err != nil {
			return []byte{}, fmt.Errorf("Can't read %s from virtual theme: %w\n", filename, err)
		}
	} else {
		fileContents, err = ioutil.ReadFile(filename)
		if err != nil {
			return []byte{}, fmt.Errorf("Can't read %s from filesystem: %w\n", filename, err)
		}
	}
	return fileContents, nil
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
