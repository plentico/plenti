package build

import (
	"testing"

	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

func TestHTMLMinifyPreservesLoopTemplateSiblingBoundaries(t *testing.T) {
	m := minify.New()
	m.Add("text/html", &html.Minifier{KeepEndTags: true})

	input := []byte(`<ul><li><a href="/elements/">1</a></li><template p-for="page of pages"><li p-text="page"></li></template><li p-for-key="s0:0">1</li></ul>`)
	got, err := minifyBytes(m, "text/html", input)
	if err != nil {
		t.Fatalf("minify HTML: %v", err)
	}

	want := `<ul><li><a href=/elements/>1</a></li><template p-for="page of pages"><li p-text=page></li></template><li p-for-key=s0:0>1</li></ul>`
	if string(got) != want {
		t.Fatalf("minified HTML changed loop sibling boundaries:\ngot:  %s\nwant: %s", got, want)
	}
}
