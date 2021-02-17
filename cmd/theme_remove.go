package cmd

import (
	"errors"
	"fmt"
	"os"
	"plenti/common"
	"plenti/readers"
	"plenti/writers"

	"github.com/spf13/cobra"
)

// themeRemoveCmd represents the theme command
var themeRemoveCmd = &cobra.Command{
	Use:   "remove [theme]",
	Short: "Completely delete all references to a theme",
	Long: `This removes the theme specific "theme_config" 
entry in plenti.json and deletes the corresponding
theme folder within the "themes/" directory.
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

		// Check if the theme_config entry exists for this theme.
		if _, ok := siteConfig.ThemeConfig[repoName]; ok {
			// Remove the corresponding theme_config entry.
			delete(siteConfig.ThemeConfig, repoName)
			// Update the config file on the filesystem.
			common.CheckErr(writers.SetSiteConfig(siteConfig, configPath))
			// Delete the corresponding theme folder.
			if err := os.RemoveAll("themes/" + repoName); err != nil {
				common.CheckErr(fmt.Errorf("Could not delete theme folder: %w", err))
			}

		}
		fmt.Printf("Could not find %v theme_config in plenti.json\n", repoName)

	},
}

func init() {
	themeCmd.AddCommand(themeRemoveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// typeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// typeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
