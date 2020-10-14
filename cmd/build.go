package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"plenti/cmd/build"
	"plenti/readers"
	"time"

	"github.com/spf13/cobra"
)

// BuildDirFlag allows users to override name of default build directory (public)
var BuildDirFlag string

// VerboseFlag provides users with additional logging information.
var VerboseFlag bool

// BenchmarkFlag provides users with build speed statistics to help identify bottlenecks.
var BenchmarkFlag bool

// NodeJSFlag let you use your systems NodeJS to build the site instead of core build.
var NodeJSFlag bool

func setBuildDir(siteConfig readers.SiteConfig) string {
	var buildDir string
	// Check if directory is overridden by flag.
	if BuildDirFlag != "" {
		// If dir flag exists, use it.
		buildDir = BuildDirFlag
	} else {
		buildDir = siteConfig.BuildDir
	}
	return buildDir
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Creates the static assets for your site",
	Long: `Build generates the actual HTML, JS, and CSS into a directory
of your choosing. The files that are created are all
you need to deploy for your website.`,
	Run: func(cmd *cobra.Command, args []string) {
		Build()
	},
}

// Build creates the compiled app that gets deployed.
func Build() {

	build.CheckVerboseFlag(VerboseFlag)
	build.CheckBenchmarkFlag(BenchmarkFlag)
	defer build.Benchmark(time.Now(), "Total build", true)

	// Handle panic when someone tries building outside of a valid Plenti site.
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("\nIt doesn't look like you're inside a valid Plenti project.")
			fmt.Println("Please create a valid project or fix your app structure before trying to run this command again.")
			fmt.Printf("Error: %v \n\n", r)
		}
	}()

	// Get settings from config file.
	siteConfig := readers.GetSiteConfig()

	// Check flags and config for directory to build to.
	buildDir := setBuildDir(siteConfig)

	// Get the full path for the build directory of the site.
	buildPath := filepath.Join(".", buildDir)

	// Clear out any previous build dir of the same name.
	if _, buildPathExistsErr := os.Stat(buildPath); buildPathExistsErr == nil {
		deleteBuildErr := os.RemoveAll(buildPath)
		build.Log("Removing old '" + buildPath + "' build directory")
		if deleteBuildErr != nil {
			fmt.Println(deleteBuildErr)
			return
		}
	}

	// Create the buildPath directory.
	err := os.MkdirAll(buildPath, os.ModePerm)
	if err != nil {
		fmt.Printf("Unable to create \"%v\" build directory: %s\n", buildDir, err)
	} else {
		build.Log("Creating '" + buildDir + "' build directory")
	}

	// Add core NPM dependencies if node_module folder doesn't already exist.
	build.NpmDefaults()

	// Write ejectable core files to filesystem before building.
	tempFiles := build.EjectTemp()

	// Directly copy .js that don't need compiling to the build dir.
	build.EjectCopy(buildPath)

	// Bundle the JavaScript dependencies needed for the build.
	//bundledContent := build.Bundle()

	// Directly copy static assets to the build dir.
	build.AssetsCopy(buildPath)

	// Run the build.js script using user local NodeJS.
	if NodeJSFlag {
		clientBuildStr := build.NodeClient(buildPath)
		staticBuildStr, allNodesStr := build.NodeDataSource(buildPath, siteConfig)
		build.NodeExec(clientBuildStr, staticBuildStr, allNodesStr)
	} else {

		// Prep the client SPA.
		build.Client(buildPath)

		// Build JSON from "content/" directory.
		build.DataSource(buildPath, siteConfig)

	}

	// Run Gopack (custom Snowpack alternative) for ESM support.
	build.Gopack(buildPath)

	// Delete any ejectable files that the user didn't manually eject.
	build.EjectClean(tempFiles)

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
	buildCmd.Flags().StringVarP(&BuildDirFlag, "dir", "d", "", "change name of the build directory")
	buildCmd.Flags().BoolVarP(&VerboseFlag, "verbose", "v", false, "show log messages")
	buildCmd.Flags().BoolVarP(&BenchmarkFlag, "benchmark", "b", false, "display build time statistics")
	buildCmd.Flags().BoolVarP(&NodeJSFlag, "nodejs", "n", false, "use system nodejs for build with ejectable build.js script")
}
