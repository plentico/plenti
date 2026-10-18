package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"github.com/plentico/plenti/cmd/build"
	"github.com/plentico/plenti/readers"

	"github.com/spf13/cobra"
)

// OutputDirFlag allows users to override name of default build directory (public)
var OutputDirFlag string

// VerboseFlag provides users with additional logging information.
var VerboseFlag bool

// BenchmarkFlag provides users with build speed statistics to help identify bottlenecks.
var BenchmarkFlag bool

// MinifyFlag condenses the HTML/CSS/JS output so it loads faster in the browser.
var MinifyFlag bool

// ConfigFileFlag allows you to point to a nonstandard sitewide configuration file for the build (instead of plenti.json).
var ConfigFileFlag string

// Building keeps track of whether the build is in process or not
var Building bool

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
	Long: `Build generates the actual static HTML, JS, and CSS into a directory
of your choosing. The files that are created are all
you need to deploy for your website.`,
	Run: func(cmd *cobra.Command, args []string) {
		Build()
	},
}

// Build creates the static site that gets deployed.
func Build() error {

	Building = true

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

	// Get theme from plenti.json.
	theme := siteConfig.Theme
	// If a theme is set, run the nested build.
	if theme != "" {
		themeOptions := siteConfig.ThemeConfig[theme]
		// Recursively copy all nested themes to a virtual filesystem for building.
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
	defer build.ThemesClean()

	// Get the full path for the build directory of the site.
	buildPath := buildDir

	// Clear out any previous build dir of the same name.
	if _, buildPathExistsErr := os.Stat(buildPath); buildPathExistsErr == nil {
		build.Log("Removing old '" + buildPath + "' build directory")
		if err = os.RemoveAll(buildPath); err != nil {
			log.Fatalf("\nCan't remove %q folder from previous build: %s", buildDir, err)
		}
	}

	// Create the buildPath directory.
	build.Log("Creating '" + buildDir + "' build directory")
	if err := os.MkdirAll(buildPath, os.ModePerm); err != nil {
		// bail on error in build
		log.Fatalf("Unable to create %q build directory: %s", buildDir, err)
	}

	// Directly copy static files to the build dir.
	err = build.StaticCopy(buildPath)
	if err != nil {
		log.Fatal("\nError in StaticCopy build step", err)
	}

	// Directly copy media to the build dir.
	mediaIndex, err := build.MediaCopy(buildPath)
	if err != nil {
		log.Fatal("\nError in MediaCopy build step", err)
	}

	// Build content list from "content/" directory.
	allContent, err := build.DataSource(siteConfig)
	if err != nil {
		log.Fatal("\nError in DataSource build step", err)
	}

	// Render static HTML pages with Pico.
	err = build.Pages(buildPath, siteConfig, allContent)
	if err != nil {
		log.Fatal("\nError in Pages build step", err)
	}

	// Build the CMS admin into the public dir.
	err = build.Cms(buildPath, siteConfig, allContent, mediaIndex)
	if err != nil {
		log.Fatal("\nError in Cms build step", err)
	}

	// Run Minification
	err = build.Minify(buildPath)
	if err != nil {
		log.Fatal("\nError in Minify build step", err)
	}

	Building = false

	// only relates to defer recover
	return err

}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().StringVarP(&OutputDirFlag, "output", "o", "", "change name of the public build directory")
	buildCmd.Flags().BoolVarP(&VerboseFlag, "verbose", "v", false, "show log messages")
	buildCmd.Flags().BoolVarP(&BenchmarkFlag, "profile", "P", false, "display build time statistics")
	buildCmd.Flags().BoolVarP(&MinifyFlag, "minify", "m", true, "minify HTML/CSS/JS output for faster performance")
	buildCmd.Flags().StringVarP(&ConfigFileFlag, "config", "c", "plenti.json", "use a custom sitewide configuration file")
}
