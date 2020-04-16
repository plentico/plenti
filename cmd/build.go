package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"plenti/cmd/build"
	"plenti/readers"

	"github.com/spf13/cobra"
)

// BuildDirFlag allows users to override name of default build directory (public)
var BuildDirFlag string

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

		// Get settings from config file.
		siteConfig := readers.GetSiteConfig()

		// Check flags and config for directory to build to.
		buildDir := setBuildDir(siteConfig)

		// Get the full path for the build directory of the site.
		buildPath := filepath.Join(".", buildDir)

		// Create the buildPath directory.
		err := os.MkdirAll(buildPath, os.ModePerm)
		if err != nil {
			fmt.Printf("Unable to create \"%v\" build directory: %s\n", buildDir, err)
		} else {
			fmt.Printf("Creating \"%v\" build directory\n", buildDir)
		}

		// Build the client SPA.
		build.Client(buildPath)

	},
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
}
