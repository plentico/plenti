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
	Theme       string                  `json:"theme"`
	ThemeConfig map[string]ThemeOptions `json:"theme_config"`
	Local       struct {
		Port int `json:"port"`
	} `json:"local"`
	Routes map[string]string `json:"routes"`
}

// ThemeOptions is the theme configuration information.
type ThemeOptions struct {
	URL     string   `json:"url"`
	Commit  string   `json:"commit"`
	Exclude []string `json:"exclude,omitempty"`
}

// GetSiteConfig reads the site's configuration file values.
func GetSiteConfig(basePath string) (SiteConfig, string) {

	var siteConfig SiteConfig

	configPath := basePath + "/plenti.json"

	// Read site config file from the project
	configFile, _ := ioutil.ReadFile(configPath)
	err := json.Unmarshal(configFile, &siteConfig)
	if err != nil {
		fmt.Println(heredoc.Docf(`

			Error: Unable to read site config file %v.

			Are you in the site folder that was created?.

			We suggest that you begin by typing:

			  cd [project name]
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

	return siteConfig, configPath
}
