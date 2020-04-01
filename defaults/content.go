package defaults

// Content : default types
var Content = map[string][]byte{
	"/content/pages/_blueprint.json": []byte(`{
	"title": "",
	"desc": "",
	"author": ""
}`),
	"/content/_index.json": []byte(`{
	"title": "My Site Homepage",
	"intro": {
		"slogan": "Welcome to a faster way to web",
		"color": "red"
	}
}`),
	"/content/pages/about.json": []byte(`{
	"title": "About Me",
	"desc": "Tell us about yourself",
	"author": "Your name"
}`),
	"/content/pages/contact.json": []byte(`{
	"title": "Contact",
	"desc": "Maybe add a <a href='https://plentiform.com'>plentiform</a>?",
	"author": "Your name"
}`),
}
