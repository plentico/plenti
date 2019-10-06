/*
Copyright © 2019 Jantcu jim.fisk@jantcu.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"plenti/readers"
	"strings"

	"github.com/spf13/cobra"
)

var BuildDirFlag string

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Creates the static assets for your site",
	Long: `Build generates the actual HTML, JS, and CSS into a directory
of your choosing. The files that are created are all
you need to deploy for your website.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Create build directory based on config file.
		siteConfig := readers.GetSiteConfig()

		// Check if directory is overridden by flag.
		if BuildDirFlag != "" {
			// If dir flag exists, use it.
			siteConfig.BuildDir = BuildDirFlag
		}

		newpath := filepath.Join(".", siteConfig.BuildDir)
		err := os.MkdirAll(newpath, os.ModePerm)
		if err != nil {
			fmt.Printf("Unable to create \"%v\" build directory\n", siteConfig.BuildDir)
			log.Fatal(err)
		} else {
			fmt.Printf("Creating \"%v\" build directory\n", siteConfig.BuildDir)
		}

		// TODO: replace hardcoded scaffolding
		var publicHTML = map[string][]byte{
			"/index.html": []byte(`<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width">
    <title>Home | Plenti</title>
  </head>
  <body>
    <h1>Welcome to Plenti</h1>
    <p>Run <pre>npm install</pre> and <pre>npm run dev</pre> to get started.</p>
    <p><a href="/about">About us</a>.</p>
    <div id="app"></div>
    <script src="/dist/bundle.js"></script>
  </body>
</html>`),
			"/about/index.html": []byte(`<!DOCTYPE html>
<html>
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width">
    <title>About | Plenti</title>
  </head>
  <body>
    <h1>About page</h1>
    <p><a href="/">Go home</a>.</p>
    <div id="app"></div>
    <script src="/dist/bundle.js"></script>
  </body>
</html>`),
		}
		for file, content := range publicHTML {
			subDirs := strings.Split(file, "/")
			prevDir := newpath
			for _, subDir := range subDirs {
				// If a file extension exists, don't create directory
				if strings.Contains(subDir, ".") {
					break
				}
				os.MkdirAll(prevDir+subDir, os.ModePerm)
				prevDir = prevDir + "/" + subDir
			}
			err := ioutil.WriteFile(newpath+file, content, 0755)
			if err != nil {
				fmt.Printf("Unable to write file: %v", err)
			}
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
	buildCmd.Flags().StringVarP(&BuildDirFlag, "dir", "d", "", "Build directory to create")
}
