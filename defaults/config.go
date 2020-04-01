package defaults

// Config : sitewide configuration file defaults
var Config = map[string][]byte{
	"/plenti.json": []byte(`{
	"baseurl": "http://example.org/",
	"title": "My New Plenti Site",
	"types": {
		"pages": "/:filename"
	},
	"build": "public",
	"local": {
		"port": 3000
	}
}`),
}
