package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"plenti/readers"
	"strings"

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

		// Create the build directory for the site.
		buildPath := filepath.Join(".", buildDir)
		err := os.MkdirAll(buildPath, os.ModePerm)
		if err != nil {
			fmt.Printf("Unable to create \"%v\" build directory: %s\n", buildDir, err)
		} else {
			fmt.Printf("Creating \"%v\" build directory\n", buildDir)
		}

		var layoutFiles []string
		layoutFilesErr := filepath.Walk("layout", func(path string, info os.FileInfo, err error) error {
			layoutFiles = append(layoutFiles, path)
			return nil
		})
		if layoutFilesErr != nil {
			fmt.Printf("Could not get layout file: %s", layoutFilesErr)
		}

		for _, layoutFile := range layoutFiles {
			// Create destination path.
			destFile := buildPath + strings.Replace(layoutFile, "layout", "/spa", 1)
			// Make sure path is a directory
			fileInfo, _ := os.Stat(layoutFile)
			if fileInfo.IsDir() {
				// Create any sub directories need for filepath.
				os.MkdirAll(destFile, os.ModePerm)
			}
			// If the file is already .js just copy it straight over to build dir.
			if filepath.Ext(layoutFile) == ".js" {
				from, err := os.Open(layoutFile)
				if err != nil {
					fmt.Printf("Could not open source .js file for copying: %s\n", err)
				}
				defer from.Close()

				to, err := os.Create(destFile)
				if err != nil {
					fmt.Printf("Could not create destination .js file for copying: %s\n", err)
				}
				defer to.Close()

				_, fileCopyErr := io.Copy(to, from)
				if err != nil {
					fmt.Printf("Could not copy .js from source to destination: %s\n", fileCopyErr)
				}
			}
			if filepath.Ext(layoutFile) == ".svelte" {
				fileContentByte, readFileErr := ioutil.ReadFile(layoutFile)
				if readFileErr != nil {
					fmt.Printf("Could not read contents of svelte source file: %s\n", readFileErr)
				}
				fileContentStr := string(fileContentByte)
				//fmt.Printf("contents: %s\n", fileContentStr)
				output, buildErr := exec.Command("node", "layout/ejected/build_client.js", fileContentStr).Output()
				//fmt.Printf("svelte output: %s\n", output)
				if buildErr != nil {
					fmt.Printf("Could not compile svelte: %s", buildErr)
				}
				destFile = strings.TrimSuffix(destFile, filepath.Ext(destFile)) + ".js"
				err := ioutil.WriteFile(destFile, output, 0755)
				if err != nil {
					fmt.Printf("Unable to write file: %v", err)
				}
			}

		}

		_, snowpackErr := exec.Command("npx", "snowpack", "--include", "'public/spa/**/*'", "--dest", "'public/spa/web_modules'").Output()
		if snowpackErr != nil {
			fmt.Printf("Snowpack failed to build npm dependencies: %s\n", snowpackErr)
		}

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
