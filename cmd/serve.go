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
	"log"
	"net/http"
	"os"
	"strconv"

	"plenti/readers"

	"github.com/spf13/cobra"
)

var PortFlag int

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

		siteConfig := readers.GetSiteConfig()

		// Check that the build directory exists
		if _, err := os.Stat(siteConfig.BuildDir); os.IsNotExist(err) {
			fmt.Printf("The \"%v\" build directory does not exist, check your config.json file.\n", siteConfig.BuildDir)
			log.Fatal(err)
		} else {
			fmt.Printf("Serving site from your \"%v\" directory.\n", siteConfig.BuildDir)
		}
		// Point to folder containing the built site
		fs := http.FileServer(http.Dir(siteConfig.BuildDir))
		http.Handle("/", fs)

		// Check if port is overridden by flag
		if PortFlag > 0 {
			// If dir flag exists, use it
			siteConfig.Local.Port = PortFlag
		}
		// Start the webserver
		fmt.Printf("Visit your site at http://localhost:%v/\n", siteConfig.Local.Port)
		err := http.ListenAndServe(":"+strconv.Itoa(siteConfig.Local.Port), nil)
		if err != nil {
			log.Fatal(err)
		}
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
	serveCmd.Flags().IntVarP(&PortFlag, "port", "p", 0, "Port for local server")
}
