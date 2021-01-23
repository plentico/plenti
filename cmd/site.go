package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"plenti/generated"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// siteCmd represents the site command
var siteCmd = &cobra.Command{
	Use:   "site [name]",
	Short: "Creates default folders and files for a new site",
	Long: heredoc.Doc(`
	The project scaffolding follows this convention:
	  - plenti.json = sitewide configuration.
	  - assets/ = holds static files like images or videos.
	  - content/ = json files that hold site content.
	  - content/pages/ = regular site pages in json format.
	  - content/pages/_blueprint.json = template for the structure of a typical page.
	  - content/pages/about.json = an example page.
	  - content/pages/contact.json = another example page.
	  - layout/ =  the html structure of the site.
	  - layout/content/ = node level structure that has a route and correspond to content.
	  - layout/components/ = smaller reusable structures that can be used within larger ones.
	  - layout/global/ = base level html wrappers.
	  - node_modules/ = frontend libraries managed by npm.
	  - package.json = npm configuration file.
	`),
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
		// Create base directory for site
		newpath := strings.Trim(filepath.Join(".", args[0]), " /")

		if _, err := os.Stat(newpath); !os.IsNotExist(err) {

			confirmPrompt := promptui.Select{
				Label: fmt.Sprintf("%s exists. Overwrite?", newpath),
				Items: []string{"No", "Yes"},
			}
			_, rep, err := (confirmPrompt.Run())
			if err != nil {
				log.Fatal(err)
			}
			if rep == "No" {
				fmt.Println("Cancelled.")
				return
			}

		}

		if err := os.MkdirAll(newpath, os.ModePerm); err != nil {
			log.Fatalf("Unable to create directory %s: %v", newpath, err)
		}

		// Check for --bare flag.
		bareFlag, err := cmd.Flags().GetBool("bare")
		if err != nil {
			log.Fatalf("Unable to get 'bare' flag: %v", err)
		}

		// set to Defaults and overwrite if bareFlag is set
		scaffolding := generated.Defaults
		// Choose which scaffolding to use for new site.
		if bareFlag {
			scaffolding = generated.Defaults_bare
		}

		// Loop through generated file defaults to create site scaffolding
		for file, content := range scaffolding {
			// Create the directories needed for the current file
			pth := fmt.Sprintf("%s/%s", newpath, strings.Trim(filepath.Dir(file), "/"))
			if err := os.MkdirAll(pth, os.ModePerm); err != nil {
				log.Fatalf("Unable to create path(s) %s: %v", pth, err)
			}

			// Create the current default file
			if err := ioutil.WriteFile(fmt.Sprintf("%s/%s", newpath, file), content, 0755); err != nil {

				log.Fatalf("Unable to write file: %v", err)
			}
		}

		// Loop through generated node_modules npm pacakges to include in scaffolding
		for file, content := range generated.Defaults_node_modules {

			// Create the directories needed for the current file
			pth := fmt.Sprintf("%s/node_modules/%s", newpath, strings.Trim(filepath.Dir(file), "/"))
			if err := os.MkdirAll(pth, os.ModePerm); err != nil {
				log.Fatalf("Unable to create path(s) %s: %v", pth, err)

			}

			// Create the current default file
			if err = ioutil.WriteFile(fmt.Sprintf("%s/%s", pth, filepath.Base(file)), content, 0755); err != nil {
				log.Fatalf("Unable to write file: %v", err)
			}
		}

		fmt.Printf(heredoc.Docf(`
			Success: Created %q site.

			We suggest that you begin by typing:

			  cd %s
			  plenti serve
		`, newpath, newpath))

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
	siteCmd.Flags().BoolP("bare", "b", false, "Omit default content from site scaffolding")
}
