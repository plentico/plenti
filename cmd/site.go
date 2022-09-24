package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/plentico/plenti/common"
	"github.com/plentico/plenti/readers"
	"github.com/plentico/plenti/writers"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// ThemeFlag starts the project using a particular theme
var themeFlag string

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
	  - content/pages/_defaults.json = template for the starting content of a page.
	  - content/pages/_schema.json = defines the field types for the CMS editor.
	  - content/pages/about.json = an example page.
	  - content/pages/contact.json = another example page.
	  - layouts/ =  the html structure of the site.
	  - layouts/content/ = node level structure that has a route and correspond to content.
	  - layouts/components/ = smaller reusable structures that can be used within larger ones.
	  - layouts/global/ = base level html wrappers.
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
		// Get user specified name of site
		projectDir := strings.Trim(filepath.Join(".", args[0]), " /")
		// Check if a folder already exists on the filesystem with that name
		if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
			// If folder already exists, ask if user wants to replace it
			confirmPrompt := promptui.Select{
				Label: fmt.Sprintf("%s exists. Overwrite?", projectDir),
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
		// Create base directory for site
		if err := os.MkdirAll(projectDir, os.ModePerm); err != nil {
			common.CheckErr(fmt.Errorf("Unable to create directory %s: %w", projectDir, err))
		}

		// Check for --theme flag.
		if themeFlag != "" {
			// Create empty base folders for the project
			defaultDirs := []string{"assets", "content", "layouts"}
			for _, dir := range defaultDirs {
				if err := os.MkdirAll(projectDir+"/"+dir, os.ModePerm); err != nil {
					common.CheckErr(fmt.Errorf("Unable to create directory %s: %w", projectDir+"/"+dir, err))
				}
			}

			// Create NPM depedencies.
			addNodeModules(projectDir)

			// Create default project files.
			writeDefaultFile("plenti.json", projectDir)
			writeDefaultFile("package.json", projectDir)
			writeDefaultFile(".gitignore", projectDir)

			// Download and set up theme.
			repoName := getRepoName(themeFlag)
			themeDir := projectDir + "/themes/" + repoName
			repo := addTheme(themeDir, themeFlag, repoName)
			commitHash := getCommitHash(repo)
			setThemeConfig(projectDir, themeFlag, commitHash, repoName)

			// Enable the theme.
			enableTheme(themeDir, projectDir, repoName)

			// Get siteConfig for project.
			siteConfig, configPath := readers.GetSiteConfig(projectDir)
			// Get siteConfig for theme.
			themeSiteConfig, _ := readers.GetSiteConfig(themeDir)
			// Copy over the route overrides from the theme.
			siteConfig.Routes = themeSiteConfig.Routes
			// Save the siteConfig for the project with the updated routes.
			common.CheckErr(writers.SetSiteConfig(siteConfig, configPath))

			return

		}

		// Check for --bare flag.
		bareFlag, err := cmd.Flags().GetBool("bare")
		if err != nil {
			common.CheckErr(fmt.Errorf("Unable to get 'bare' flag: %w", err))
		}

		// set to Defaults and overwrite if bareFlag is set
		scaffolding, err := fs.Sub(defaultsLearnerFS, "defaults/starters/learner")
		if err != nil {
			common.CheckErr(fmt.Errorf("Unable to get learner defaults: %w", err))
		}
		// Choose which scaffolding to use for new site.
		if bareFlag {
			scaffolding, err = fs.Sub(defaultsBareFS, "defaults/starters/bare")
			if err != nil {
				common.CheckErr(fmt.Errorf("Unable to get bare defaults: %w", err))
			}
		}
		// Loop through site defaults to create site scaffolding
		writeScaffolding(scaffolding, projectDir)

		// Create NPM depedencies.
		addNodeModules(projectDir)

		fmt.Printf(heredoc.Docf(`
			Success: Created %q site.

			We suggest that you begin by typing:

			  cd %s
			  plenti serve
		`, projectDir, projectDir))

	},
}

func writeDefaultFile(filename string, projectDir string) {
	defaultFile, err := defaultsBareFS.ReadFile("defaults/starters/bare/" + filename)
	if err != nil {
		fmt.Printf("Can't read default file '%s': %s\n", filename, err)
	}
	// Create the current default file
	if err := ioutil.WriteFile(projectDir+"/"+filename, defaultFile, os.ModePerm); err != nil {
		common.CheckErr(fmt.Errorf("Unable to write file '%s': %w\n", filename, err))
	}
}

func writeScaffolding(defaults fs.FS, projectDir string) {
	fs.WalkDir(defaults, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Create the directories needed for the current file
			if err := os.MkdirAll(projectDir+"/"+path, os.ModePerm); err != nil {
				common.CheckErr(fmt.Errorf("Unable to create path(s) %s: %v", path, err))
			}
			return nil
		}
		content, _ := defaults.Open(path)
		contentBytes, err := ioutil.ReadAll(content)
		// Create the current default file
		if err := ioutil.WriteFile(projectDir+"/"+path, contentBytes, 0755); err != nil {
			common.CheckErr(fmt.Errorf("Unable to write file: %w", err))
		}
		return nil
	})

}

func addNodeModules(projectDir string) {
	nodeModules, err := fs.Sub(defaultsNodeModulesFS, "defaults")
	if err != nil {
		common.CheckErr(fmt.Errorf("Unable to get node_modules defaults: %w", err))
	}
	// Loop through node_modules npm pacakges to include in scaffolding
	writeScaffolding(nodeModules, projectDir)
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
	siteCmd.Flags().StringVarP(&themeFlag, "theme", "t", "", "start your project by inheriting from an existing theme")
}
