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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"plenti/defaults"

	"github.com/spf13/cobra"
)

// siteCmd represents the site command
var siteCmd = &cobra.Command{
	Use:   "site [name]",
	Short: "Creates default folders and files for a new site",
	Long: `The project scaffolding follows this convention:
- config.yml = sitewide configuration
- content/ = json files that hold site content
- content/pages/ = regular site pages in json format
- content/pages/_archetype.json = template for the structure of a typical page
- content/pages/_index.json = the aggregate, or landing page
- content/pages/about.json = an example page
- content/pages/contact.json = another example page
- templates/ =  all react components
- templates/routed/ = page level react components that correspond to content
- templates/components/ = smaller react components that are used within larger ones
- templates/layouts/ = base level html wrappers
- node_modules/ = frontend libraries managed by npm
- package.json = npm configuration file
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a name argument")
		}
		if len(args) > 1 {
			return errors.New("names cannot have spaces")
		}
		if len(args) == 1 {
			return nil
		}
		return fmt.Errorf("invalid name specified: %s", args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Create directory structure
		newpath := filepath.Join(".", args[0])
		os.MkdirAll(newpath, os.ModePerm)
		dirs := []string{
			filepath.Join(newpath, "assets"),
			filepath.Join(newpath, "assets/css"),
			filepath.Join(newpath, "assets/js"),
			filepath.Join(newpath, "assets/img"),
			filepath.Join(newpath, "content"),
			filepath.Join(newpath, "content/pages"),
			filepath.Join(newpath, "templates"),
			filepath.Join(newpath, "templates/routed"),
			filepath.Join(newpath, "templates/components"),
			filepath.Join(newpath, "templates/layouts"),
		}
		for _, dir := range dirs {
			os.MkdirAll(dir, os.ModePerm)
		}

		// Get code templates from defaults dir
		allDefaults := []map[string][]byte{
			defaults.Assets,
			defaults.Config,
			defaults.Content,
			defaults.Templates,
			defaults.Vendor,
		}
		// Create files and populate default file code
		for _, defaultGroup := range allDefaults {
			for file, content := range defaultGroup {
				err := ioutil.WriteFile(newpath+file, content, 0755)
				if err != nil {
					fmt.Printf("Unable to write file: %v", err)
				}
			}
		}

		fmt.Printf("Created plenti site scaffolding in \"%v\" folder\n", newpath)

	},
}

func init() {
	newCmd.AddCommand(siteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// siteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// siteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
