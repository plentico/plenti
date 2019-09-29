package defaults

var Config = map[string][]byte{
	"/config.json": []byte(`{
	"baseurl": "http://example.org/",
	"title": "My New Plenti Site",
	"types": {
		"pages": "/:filename"
	}
}`),
}
