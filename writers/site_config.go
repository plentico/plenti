package writers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"plenti/readers"
)

// SetSiteConfig writes values to the site's configuration file.
func SetSiteConfig(siteConfig readers.SiteConfig, configPath string) {

	//result, err := json.Marshal(siteConfig)
	result, err := json.MarshalIndent(siteConfig, "", "\t")
	if err != nil {
		fmt.Printf("Unable to marshal JSON: %s\n", err)
	}

	// Write values to site config file for the project.
	writeErr := ioutil.WriteFile(configPath, result, os.ModePerm)
	if writeErr != nil {
		fmt.Printf("Unable to write to config file: %s\n", writeErr)
	}
}
