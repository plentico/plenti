package cmd

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"plenti/common"
	"plenti/generated"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

//go:embed defaults/*
var defaultsFS embed.FS

//go:embed defaults_bare/*
var defaultsBareFS embed.FS

//go:embed defaults_node_modules/*
var defaultsNodeModulesFS embed.FS

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
			common.CheckErr(fmt.Errorf("Unable to create directory %s: %w", newpath, err))
		}

		// Check for --bare flag.
		bareFlag, err := cmd.Flags().GetBool("bare")
		if err != nil {
			common.CheckErr(fmt.Errorf("Unable to get 'bare' flag: %w", err))
		}

		// set to Defaults and overwrite if bareFlag is set
		scaffolding, err := fs.Sub(defaultsFS, "defaults")
		if err != nil {
			common.CheckErr(fmt.Errorf("Unable to get defaults: %w", err))
		}
		// Choose which scaffolding to use for new site.
		if bareFlag {
			scaffolding, err = fs.Sub(defaultsBareFS, "defaults_bare")
			if err != nil {
				common.CheckErr(fmt.Errorf("Unable to get bare defaults: %w", err))
			}
		}

		// Loop through generated file defaults to create site scaffolding
		fs.WalkDir(scaffolding, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				// Create the directories needed for the current file
				if err := os.MkdirAll(newpath+"/"+path, os.ModePerm); err != nil {
					common.CheckErr(fmt.Errorf("Unable to create path(s) %s: %v", path, err))
				}
				return nil
			}
			content, _ := scaffolding.Open(path)
			contentBytes, err := ioutil.ReadAll(content)
			// Create the current default file
			if err := ioutil.WriteFile(newpath+"/"+path, contentBytes, 0755); err != nil {
				common.CheckErr(fmt.Errorf("Unable to write file: %w", err))
			}
			return nil
		})
		/*
			for file, content := range scaffolding {
				// Create the directories needed for the current file
				pth := fmt.Sprintf("%s/%s", newpath, strings.Trim(filepath.Dir(file), "/"))
				if err := os.MkdirAll(pth, os.ModePerm); err != nil {
					common.CheckErr(fmt.Errorf("Unable to create path(s) %s: %v", pth, err))
				}

				// Create the current default file
				if err := ioutil.WriteFile(fmt.Sprintf("%s/%s", newpath, file), content, 0755); err != nil {

					common.CheckErr(fmt.Errorf("Unable to write file: %w", err))
				}
			}
		*/

		// Loop through generated node_modules npm pacakges to include in scaffolding
		for file, content := range generated.Defaults_node_modules {

			// Create the directories needed for the current file
			pth := fmt.Sprintf("%s/node_modules/%s", newpath, strings.Trim(filepath.Dir(file), "/"))
			if err := os.MkdirAll(pth, os.ModePerm); err != nil {
				common.CheckErr(fmt.Errorf("Unable to create path(s) %s: %w", pth, err))

			}

			// Create the current default file
			if err = ioutil.WriteFile(fmt.Sprintf("%s/%s", pth, filepath.Base(file)), content, 0755); err != nil {
				common.CheckErr(fmt.Errorf("Unable to write file: %w", err))
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
