package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/plentico/plenti/cmd/build"
	"github.com/plentico/plenti/common"
	"github.com/plentico/plenti/readers"

	"github.com/spf13/cobra"
)

// OutputDirFlag allows users to override name of default build directory (public)
var OutputDirFlag string

// VerboseFlag provides users with additional logging information.
var VerboseFlag bool

// BenchmarkFlag provides users with build speed statistics to help identify bottlenecks.
var BenchmarkFlag bool

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
	Short: "Creates the static assets for your site",
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

	themeBuildDir := ""

	// Add core NPM dependencies if node_module folder doesn't already exist.
	if err = common.CheckErr(build.NpmDefaults(defaultsNodeModulesFS)); err != nil {
		return err
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
		if err = common.CheckErr(err); err != nil {
			return err
		}

		// Merge the current project files with the theme.
		if err = common.CheckErr(build.ThemesMerge(buildDir)); err != nil {
			return err
		}

	}

	// Get the full path for the build directory of the site.
	buildPath := filepath.Join(".", buildDir)

	// Clear out any previous build dir of the same name.
	if !common.UseMemFS {
		if _, buildPathExistsErr := os.Stat(buildPath); buildPathExistsErr == nil {
			build.Log("Removing old '" + buildPath + "' build directory")
			if err = common.CheckErr(os.RemoveAll(buildPath)); err != nil {
				return err
			}

		}

		// Create the buildPath directory.
		if err := os.MkdirAll(buildPath, os.ModePerm); err != nil {
			// bail on error in build
			if err = common.CheckErr(fmt.Errorf("Unable to create \"%v\" build directory: %s", err, buildDir)); err != nil {
				return err
			}

		}
	}
	build.Log("Creating '" + buildDir + "' build directory")

	// Directly copy .js that don't need compiling to the build dir.
	if err = common.CheckErr(build.EjectCopy(buildPath, defaultsEjectedFS)); err != nil {
		return err
	}

	// Directly copy static assets to the build dir.
	if err = common.CheckErr(build.AssetsCopy(buildPath)); err != nil {
		return err
	}

	// Prep the client SPA.
	err = build.Client(buildPath, defaultsEjectedFS)
	if err = common.CheckErr(err); err != nil {
		return err
	}

	// Build JSON from "content/" directory.
	err = build.DataSource(buildPath, siteConfig, themeBuildDir)
	if err = common.CheckErr(err); err != nil {
		return err
	}

	// Run Gopack (custom Snowpack alternative) for ESM support.
	if err = common.CheckErr(build.Gopack(buildPath)); err != nil {
		return err
	}

	// If using themes, clean up the virtual filesystem afterwards.
	/*
		if build.AppFs != nil {
			build.AppFs.RemoveAll(".")
		}
	*/

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
}
