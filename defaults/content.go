package defaults

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
)

func getFileContents() []byte {
	absPath, _ := filepath.Abs("index.json")
	index, err := ioutil.ReadFile(absPath)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(index)
	return index
}

//var index, _ = ioutil.ReadFile("../defaults/content/_index.json")
var blueprint, _ = ioutil.ReadFile("./content/_blueprint.json.go")
var about, _ = ioutil.ReadFile("./content/about.json.go")
var contact, _ = ioutil.ReadFile("./content/contact.json.go")

// Content : default types of content
var Content = map[string][]byte{
	"/content/_index.json":           getFileContents(),
	"/content/pages/_blueprint.json": blueprint,
	"/content/pages/about.json":      about,
	"/content/pages/contact.json":    contact,
}
