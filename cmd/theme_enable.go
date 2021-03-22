package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/plentico/plenti/common"
	"github.com/plentico/plenti/readers"
	"github.com/plentico/plenti/writers"

	"github.com/spf13/cobra"
)

// themeEnableCmd represents the theme command
var themeEnableCmd = &cobra.Command{
	Use:   "enable [theme]",
	Short: "Use a specific theme as a starting point for your project",
	Long: `Enabling a theme adds a "theme" entry to plenti.json. Once
this has been added, builds will inherit assets, content,
and layout from the theme you enabled.
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

		if _, err := os.Stat("themes/" + repoName); !os.IsNotExist(err) {
			siteConfig.Theme = repoName

			// Update the config file on the filesystem.
			common.CheckErr(writers.SetSiteConfig(siteConfig, configPath))
		} else {
			fmt.Printf("Could not locate '%v' theme: %v\n", repoName, err)
		}

	},
}

func init() {
	themeCmd.AddCommand(themeEnableCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// typeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// typeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
