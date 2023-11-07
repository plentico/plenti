package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/livebud/npm"
	"github.com/plentico/plenti/readers"

	"github.com/spf13/cobra"
)

// npmInstallCmd represents the npm install command
var npmInstallCmd = &cobra.Command{
	Use:   "npm install",
	Short: "Install packages from NPM",
	Long: `Download packages from NPM without requiring NodeJS/NPM.
This is especially helpful for lighter containers when building in CI.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("must run the \"install\" command")
		}
		if len(args) > 1 {
			return errors.New("must only run the \"install\" command")
		}
		if len(args) == 1 && args[0] == "install" {
			return nil
		}
		return fmt.Errorf("must run \"install\" - invalid command: %s", args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get the current site NPM configuration file values.
		npmConfig := readers.GetNpmConfig("package.json")
		modules := npmConfig.Dependencies

		var flattenedModules []string
		for module, version := range modules {
			// TODO: should get these values from package-lock.json and get exact version
			// instead of processing and manually removing the ^ from the beginning of version
			flattenedModules = append(flattenedModules, module+"@"+version[1:])
		}

		ctx := context.Background()
		dir := "."
		err := npm.Install(ctx, dir, flattenedModules...)
		if err != nil {
			log.Fatal("Could not install NPM dependencies\n", err)
		}

	},
}

func init() {
	rootCmd.AddCommand(npmInstallCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// typeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// typeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
