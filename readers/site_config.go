package readers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

type SiteConfig struct {
	BuildDir string `json:"build"`
	Local    struct {
		Port int `json:"port"`
	} `json:"local"`
}

func GetSiteConfig() SiteConfig {

	var siteConfig SiteConfig

	// Read site config file from the project
	configFile, _ := ioutil.ReadFile("config.json")
	err := json.Unmarshal(configFile, &siteConfig)
	if err != nil {
		fmt.Printf("Unable to read config file.\n")
		log.Fatal(err)
	}

	// If build directory is not set in config, use default
	if siteConfig.BuildDir == "" {
		siteConfig.BuildDir = "public"
	}

	return siteConfig
}
