package build

import (
	"regexp"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/js"
)

func loadJSMinifier() *minify.M {
	m := minify.New()
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	return m
}
func minifyJS(m *minify.M, jsBytes []byte) []byte {
	jsBytes, err := m.Bytes("text/javascript", jsBytes)
	if err != nil {
		panic(err)
	}
	return jsBytes
}
