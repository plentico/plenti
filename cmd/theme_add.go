package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/plentico/plenti/common"
	"github.com/plentico/plenti/readers"
	"github.com/plentico/plenti/writers"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"
)

// CommitFlag targets a specific commit hash when running the "git clone" operation.
var CommitFlag string

// themeAddCmd represents the theme command
var themeAddCmd = &cobra.Command{
	Use:   "add [url]",
	Short: "Downloads parent theme to inherit content, layouts, and assets from",
	Long: `Themes allow you to leverage an existing Plenti site as a starting point for your own site.

To use https://plenti.co as a theme for example, run: plenti new theme git@github.com:plentico/plenti.co
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

		// Get just the repository name for the URL.
		repoName := getRepoName(url)

		// Set the directory to clone code into.
		themeDir := "themes/" + repoName

		// Clone the theme.
		repo := addTheme(themeDir, url, repoName)
		// Get the hash the represents the version of the theme.
		commitHash := getCommitHash(repo)
		// Update the plenti.json config file with theme info.
		setThemeConfig(".", url, commitHash, repoName)
		// Remove .git folder from theme to avoid submodules.
		cleanThemeGit(themeDir)
	},
}

func getRepoName(url string) string {
	// Get the last part of the git URL to isolate the repository name.
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}

func addTheme(themeDir string, url string, repoName string) *git.Repository {
	// Run the "git clone" operation.
	repo, err := git.PlainClone(themeDir, false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		common.CheckErr(fmt.Errorf("Can't clone theme repository: %w", err))
	}
	return repo
}

func getCommitHash(repo *git.Repository) string {
	// Get the latest commit hash from the repo.
	ref, err := repo.Head()
	if err != nil {
		common.CheckErr(fmt.Errorf("Can't get HEAD: %w", err))

	}
	commitObj, err := repo.CommitObject(ref.Hash())
	if err != nil {
		common.CheckErr(fmt.Errorf("Can't get Commit from hash: %w", err))

	}
	commitHash := commitObj.Hash.String()

	// Check if a --commit flag was used.
	if CommitFlag != "" {
		worktree, worktreeErr := repo.Worktree()
		if worktreeErr != nil {
			common.CheckErr(fmt.Errorf("Can't get worktree: %w", worktreeErr))

		}
		// Resolve commit in case short hash is used instead of full hash.
		resolvedCommitHash, resolveErr := repo.ResolveRevision(plumbing.Revision(CommitFlag))
		if resolveErr != nil {
			common.CheckErr(fmt.Errorf("Can't resolve commit hash: %w", resolveErr))

		}
		// Git checkout the commit hash that was sent via the flag.
		if checkoutErr := worktree.Checkout(&git.CheckoutOptions{
			Hash: *resolvedCommitHash,
		}); checkoutErr != nil {
			common.CheckErr(fmt.Errorf("Can't get commit: %w", checkoutErr))
		}
		// The --commit flag could be checkout out, so the hash is valid.
		commitHash = CommitFlag
	}
	return commitHash
}

func setThemeConfig(configLocation string, url string, commitHash string, repoName string) {
	// Get the current site configuration file values.
	siteConfig, configPath := readers.GetSiteConfig(configLocation)
	// Update the sitConfig struct with new values.
	themeOptions := new(readers.ThemeOptions)
	themeOptions.URL = url
	themeOptions.Commit = commitHash
	themeOptions.Exclude = siteConfig.ThemeConfig[repoName].Exclude
	if siteConfig.ThemeConfig == nil {
		siteConfig.ThemeConfig = make(map[string]readers.ThemeOptions)
	}
	siteConfig.ThemeConfig[repoName] = *themeOptions

	// Update the config file on the filesystem.
	common.CheckErr(writers.SetSiteConfig(siteConfig, configPath))
}

func cleanThemeGit(themeDir string) {
	// Remove the theme's .git/ folder to avoid submodule issues.
	if err := os.RemoveAll(themeDir + "/.git"); err != nil {
		common.CheckErr(fmt.Errorf("Could not delete .git folder for theme: %w", err))
	}
}

func init() {
	themeCmd.AddCommand(themeAddCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// typeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// typeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	themeAddCmd.Flags().StringVarP(&CommitFlag, "commit", "c", "", "pull a specific commit hash for the theme")
}
