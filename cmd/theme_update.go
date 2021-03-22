package cmd

import (
	"errors"
	"fmt"

	"github.com/plentico/plenti/readers"

	"github.com/spf13/cobra"
)

// themeUpdateCmd represents the theme command
var themeUpdateCmd = &cobra.Command{
	Use:   "update [theme]",
	Short: "Get a newer version of an existing theme",
	Long: `Updating themes uses go-git behind the scenes to check
if there are newer versions for the theme available. Plenti
manages the git commits behind the scenes so you don't have
to worry about tracking code with git submodules.

You will need a valid theme_config in your plenti.json
file in order to pull updates.
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

		// Get the repo URL passed via the CLI.
		repoName := args[0]

		// Get the current site configuration file values.
		siteConfig, _ := readers.GetSiteConfig(".")

		// Get the corresponding git remote for the theme.
		url := siteConfig.ThemeConfig[repoName].URL

		// Check that we were able to get the URL from the config file.
		if url == "" {
			fmt.Println("Could not find URL for theme, fix theme_config in plenti.json")
			return

		}
		// Run "theme add" to get new version of the theme.
		themeAddCmd.Run(cmd, []string{url})

	},
}

func init() {
	themeCmd.AddCommand(themeUpdateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// typeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// typeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	themeUpdateCmd.Flags().StringVarP(&CommitFlag, "commit", "c", "", "pull a specific commit hash for the theme")
}
