package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/plentico/plenti/cmd/build"
	"github.com/plentico/plenti/defaults"
	"github.com/plentico/plenti/readers"

	"github.com/spf13/cobra"
)

// OutputDirFlag allows users to override name of default build directory (public)
var OutputDirFlag string

// VerboseFlag provides users with additional logging information.
var VerboseFlag bool

// BenchmarkFlag provides users with build speed statistics to help identify bottlenecks.
var BenchmarkFlag bool

// MinifyFlag condenses the JavaScript output so it runs faster in the browser.
var MinifyFlag bool

// ConfigFileFlag allows you to point to a nonstandard sitewide configuration file for the build (instead of plenti.json).
var ConfigFileFlag string

func setBuildDir(siteConfig readers.SiteConfig) string {
	buildDir := siteConfig.BuildDir
	// Check if directory is overridden by flag.
	if OutputDirFlag != "" {
		// If dir flag exists, use it.
		buildDir = OutputDirFlag
	}
	return buildDir
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Creates the public assets for your site",
	Long: `Build generates the actual HTML, JS, and CSS into a directory
of your choosing. The files that are created are all
you need to deploy for your website.`,
	Run: func(cmd *cobra.Command, args []string) {
		Build()
	},
}

// Build creates the compiled app that gets deployed.
func Build() error {

	defer build.Benchmark(time.Now(), "Total build", true)

	build.CheckVerboseFlag(VerboseFlag)
	build.CheckBenchmarkFlag(BenchmarkFlag)
	build.CheckMinifyFlag(MinifyFlag)

	var err error
	// Handle panic when someone tries building outside of a valid Plenti site.
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Please create a valid Plenti project or fix your app structure before trying to run this command again.")
			fmt.Printf("Error: %v \n\n", r)
			debug.PrintStack()
			err = fmt.Errorf("panic recovered in Build: %v", r)
		}
	}()

	// Get settings from config file.
	siteConfig, _ := readers.GetSiteConfig(".")

	// Check flags and config for directory to build to.
	buildDir := setBuildDir(siteConfig)

	// Add core NPM dependencies if node_module folder doesn't already exist.
	err = build.NpmDefaults(defaults.NodeModulesFS)
	if err != nil {
		log.Fatal("\nError in NpmDefaults build step", err)
	}
	// TODO: ^ only adds node_modules to root project.
	// We should think of a way to honor theme dependecies,
	// which aren't usually tracked in git.

	// Get theme from plenti.json.
	theme := siteConfig.Theme
	// If a theme is set, run the nested build.
	if theme != "" {
		themeOptions := siteConfig.ThemeConfig[theme]
		// Recursively copy all nested themes to a temp folder for building.
		err = build.ThemesCopy("themes/"+theme, themeOptions)
		if err != nil {
			log.Fatal("\nError in ThemesCopy build step", err)
		}

		// Merge the current project files with the theme.
		err = build.ThemesMerge(buildDir)
		if err != nil {
			log.Fatal("\nError in ThemesMerge build step", err)
		}
	}

	// Get the full path for the build directory of the site.
	buildPath := filepath.Join(".", buildDir)

	// Clear out any previous build dir of the same name.
	if _, buildPathExistsErr := os.Stat(buildPath); buildPathExistsErr == nil {
		build.Log("Removing old '" + buildPath + "' build directory")
		if err = os.RemoveAll(buildPath); err != nil {
			log.Fatal("\nCan't remove \"%v\" folder from previous build", err, buildDir)
		}
	}

	// Create the buildPath directory.
	build.Log("Creating '" + buildDir + "' build directory")
	if err := os.MkdirAll(buildPath, os.ModePerm); err != nil {
		// bail on error in build
		log.Fatal("Unable to create \"%v\" build directory: %s", err, buildDir)
	}

	// Directly copy .js that don't need compiling to the build dir.
	err = build.EjectCopy(buildPath, defaults.CoreFS)
	if err != nil {
		log.Fatal("\nError in EjectCopy build step", err)
	}

	// Directly copy static files to the build dir.
	err = build.StaticCopy(buildPath)
	if err != nil {
		log.Fatal("\nError in StaticCopy build step", err)
	}

	// Directly copy media to the build dir.
	err = build.MediaCopy(buildPath)
	if err != nil {
		log.Fatal("\nError in MediaCopy build step", err)
	}

	// Prep the client SPA.
	err = build.Client(buildPath, defaults.CoreFS)
	if err != nil {
		log.Fatal("\nError in Client build step", err)
	}

	// Build JSON from "content/" directory.
	err = build.DataSource(buildPath, siteConfig)
	if err != nil {
		log.Fatal("\nError in DataSource build step", err)
	}

	// Run Gopack (custom Snowpack alternative) on app for ESM support.
	err = build.Gopack(buildPath, buildPath+"/spa/core/main.js")
	if err != nil {
		log.Fatal("\nError in Gopack build step", err)
	}

	// Run Gopack manually on dynamic imports
	err = build.GopackDynamic(buildPath)
	if err != nil {
		log.Fatal("\nError in GopackDynamic build step", err)
	}

	// only relates to defer recover
	return err

}

func init() {
	rootCmd.AddCommand(buildCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// buildCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// buildCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	buildCmd.Flags().StringVarP(&OutputDirFlag, "output", "o", "", "change name of the public build directory")
	buildCmd.Flags().BoolVarP(&VerboseFlag, "verbose", "v", false, "show log messages")
	buildCmd.Flags().BoolVarP(&BenchmarkFlag, "benchmark", "b", false, "display build time statistics")
	buildCmd.Flags().BoolVarP(&MinifyFlag, "minify", "m", true, "minify JS output for faster performance")
	buildCmd.Flags().StringVarP(&ConfigFileFlag, "config", "c", "plenti.json", "use a custom sitewide configuration file")
}
