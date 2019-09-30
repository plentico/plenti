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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

type SiteConfig struct {
	Build string `json:"build"`
}

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Lightweight webserver for local development",
	Long: `Serve will run "plenti build" automatically to create
a compiled version of your site. This defaults to
folder named "public" but you can adjust this in
your site config.

You can also set a different port in your site config file.`,
	Run: func(cmd *cobra.Command, args []string) {

		// Read site config file from the project
		configFile, _ := ioutil.ReadFile("config.json")
		var siteConfig SiteConfig
		err := json.Unmarshal(configFile, &siteConfig)
		if err != nil {
			fmt.Printf("Unable to read config file: %v", err)
		}

		buildDir := "public"
		// Attempt to set build directory from config file
		if siteConfig.Build != "" {
			buildDir = siteConfig.Build
		}

		// Check that the build directory exists
		if _, err := os.Stat(buildDir); os.IsNotExist(err) {
			fmt.Printf("The \"%v\" build directory does not exist, check your config.json file.\n", buildDir)
			log.Fatal(err)
		} else {
			fmt.Printf("Serving site from your \"%v\" directory.\n", buildDir)
		}
		// Point to folder containing the built site
		fs := http.FileServer(http.Dir(buildDir))
		http.Handle("/", fs)

		// Start the webserver
		fmt.Printf("Visit your site at http://localhost:3000/\n")
		http.ListenAndServe(":3000", nil)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
