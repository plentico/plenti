// Package readers knows how to read different configuration files.
package readers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/MakeNowJust/heredoc/v2"
)

// SiteConfig is the site's configuration file values.
type SiteConfig struct {
	BuildDir    string                  `json:"build"`
	BaseURL     string                  `json:"baseurl"`
	Theme       string                  `json:"theme"`
	EntryPoint  string                  `json:"entrypoint"`
	LoadLayouts []string                `json:"loadLayouts"`
	ThemeConfig map[string]ThemeOptions `json:"theme_config"`
	Local       struct {
		Port int `json:"port"`
	} `json:"local"`
	Routes map[string]string `json:"routes"`
	CMS    struct {
		Repo        string `json:"repo"`
		RedirectUrl string `json:"redirect_url"`
		AppId       string `json:"app_id"`
		Branch      string `json:"branch"`
	} `json:"cms"`
}

// ThemeOptions is the theme configuration information.
type ThemeOptions struct {
	URL     string   `json:"url"`
	Commit  string   `json:"commit"`
	Exclude []string `json:"exclude,omitempty"`
}

// Create global var since cmd.ConfigFileFlag is a circular dependency.
var configFilePath string

// CheckConfigFileFlag sets global var to --config flag value (or defaults to plenti.json).
func CheckConfigFileFlag(flag string) {
	// If --config flag is passed by user, this will be set to its value.
	configFilePath = flag
}

// GetSiteConfig reads the site's configuration file values.
func GetSiteConfig(basePath string) (SiteConfig, string) {

	var siteConfig SiteConfig

	configPath := basePath + "/" + configFilePath

	// Read site config file from the project
	configFile, _ := ioutil.ReadFile(configPath)
	err := json.Unmarshal(configFile, &siteConfig)
	if err != nil {
		fmt.Println(heredoc.Docf(`

			Error: Unable to read plenti.json: %v ❌

			Are you in the folder for your project?

			Start by typing:

			  cd [your project name]
			  plenti serve
		`, err))
		os.Exit(1)
	}

	// If build directory is not set in config, use default
	if siteConfig.BuildDir == "" {
		siteConfig.BuildDir = "public"
	}

	// If local server port is not set in config, use default
	if siteConfig.Local.Port <= 0 {
		siteConfig.Local.Port = 3000
	}

	// If entrypoint is not set in config, use default
	if siteConfig.EntryPoint == "" {
		siteConfig.EntryPoint = "global/html.svelte"
	}

	// If loadComponents is not set in config, use default
	if len(siteConfig.LoadLayouts) == 0 {
		siteConfig.LoadLayouts = []string{"*"}
	}

	return siteConfig, configPath
}
