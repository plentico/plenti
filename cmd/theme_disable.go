package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/plentico/plenti/readers"
	"github.com/plentico/plenti/writers"

	"github.com/spf13/cobra"
)

// themeEnableCmd represents the theme command
var themeDisableCmd = &cobra.Command{
	Use:   "disable [theme]",
	Short: "Stop actively using a specific theme in your project",
	Long: `Disabling a theme removes the "theme" entry in plenti.json. Your
will no longer inherit assets, content, and layouts from this theme.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a theme argument")
		}
		if len(args) > 1 {
			return errors.New("theme cannot have spaces")
		}
		if len(args) == 1 {
			return nil
		}
		return fmt.Errorf("invalid theme specified: %s", args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {

		// Get the theme name passed via the CLI.
		repoName := args[0]

		// Get the current site configuration file values.
		siteConfig, configPath := readers.GetSiteConfig(".")

		if siteConfig.Theme == "" {
			fmt.Println("No theme is currently enabled.")
			return
		}

		if siteConfig.Theme != repoName {
			fmt.Printf("Theme '%v' is not currently enabled. The enabled theme is: %v\n", repoName, siteConfig.Theme)
			return
		}
		siteConfig.Theme = ""
		// Update the config file on the filesystem.
		err := writers.SetSiteConfig(siteConfig, configPath)
		if err != nil {
			log.Fatal("Could not update site config file %w", err)
		}

	},
}

func init() {
	themeCmd.AddCommand(themeDisableCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// typeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// typeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
