package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

// CommitFlag targets a specific commit hash when running the "git clone" operation.
var CommitFlag string

// themeCmd represents the theme command
var themeCmd = &cobra.Command{
	Use:   "theme [url]",
	Short: "Downloads parent theme to inherit content, layouts, and assets from",
	Long: `Themes allow you to leverage an existing Plenti site as a starting point for your own site.

To use https://plenti.co as a theme for example, run: plenti new theme git@github.com:plentico/plenti.co.git
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires a url argument")
		}
		if len(args) > 1 {
			return errors.New("urls cannot have spaces")
		}
		if len(args) == 1 {
			return nil
		}
		return fmt.Errorf("invalid url specified: %s", args[0])
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get the repo URL passed via the CLI.
		url := args[0]

		// Get the last part of the URL to isolate the repository name.
		parts := strings.Split(url, "/")
		repoName := parts[len(parts)-1]

		themeDir := "themes/" + repoName

		// Run the "git clone" operation.
		r, err := git.PlainClone(themeDir, false, &git.CloneOptions{
			URL:      url,
			Progress: os.Stdout,
		})
		if err != nil {
			fmt.Printf("Can't clone theme repository: %v\n", err)
		}

		// Get the latest commit hash from the repo.
		ref, _ := r.Head()
		commitObj, _ := r.CommitObject(ref.Hash())
		commitHash := commitObj.Hash
		fmt.Println(commitHash)

	},
}

func init() {
	newCmd.AddCommand(themeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// typeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// typeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	themeCmd.Flags().StringVarP(&CommitFlag, "commit", "c", "", "pull a specific commit hash for the theme")
}
