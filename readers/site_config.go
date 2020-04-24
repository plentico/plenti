package readers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

// SiteConfig is the site's configuration file values.
type SiteConfig struct {
	BuildDir string `json:"build"`
	Local    struct {
		Port int `json:"port"`
	} `json:"local"`
	Types map[string]string `json:"types"`
}

// GetSiteConfig reads the site's configuration file values.
func GetSiteConfig() SiteConfig {

	var siteConfig SiteConfig

	// Read site config file from the project
	configFile, _ := ioutil.ReadFile("plenti.json")
	err := json.Unmarshal(configFile, &siteConfig)
	if err != nil {
		fmt.Printf("Unable to read config file.\n")
		log.Fatal(err)
	}

	// If build directory is not set in config, use default
	if siteConfig.BuildDir == "" {
		siteConfig.BuildDir = "public"
	}

	// If local server port is not set in config, use default
	if siteConfig.Local.Port <= 0 {
		siteConfig.Local.Port = 3000
	}

	return siteConfig
}
