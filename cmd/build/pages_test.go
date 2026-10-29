package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plentico/plenti/readers"
)

func TestPagesProvidesAllContent(t *testing.T) {
	projectDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Fatal(err)
		}
	})

	rootLayout := filepath.Join("layouts", "global", "html.pico")
	if err := os.MkdirAll(filepath.Dir(rootLayout), 0755); err != nil {
		t.Fatal(err)
	}
	root := `---
prop content;
prop allContent;
let posts = allContent.filter(node => node.type === "posts");
---
<html><head></head><body>
{for let post in posts}
<article data-path="{post.path}">{post.fields.title}</article>
{/for}
</body></html>`
	if err := os.WriteFile(rootLayout, []byte(root), 0644); err != nil {
		t.Fatal(err)
	}

	pagesLayout := filepath.Join("layouts", "content", "pages.pico")
	if err := os.MkdirAll(filepath.Dir(pagesLayout), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pagesLayout, []byte("<main>{content.fields.title}</main>"), 0644); err != nil {
		t.Fatal(err)
	}

	content := []ContentNode{
		{Type: "pages", Path: "/", FilePath: "pages/index.json", FileName: "index", Fields: map[string]any{"title": "Home"}},
		{Type: "posts", Path: "/hello", FilePath: "posts/hello.json", FileName: "hello", Fields: map[string]any{"title": "Hello"}},
		{Type: "posts", Path: "/events", FilePath: "posts/events.json", FileName: "events", Fields: map[string]any{"title": "Events"}},
	}
	buildPath := filepath.Join(projectDir, "public")
	config := readers.SiteConfig{BaseLayout: "global/html.pico"}
	if err := Pages(buildPath, config, content); err != nil {
		t.Fatal(err)
	}

	page, err := os.ReadFile(filepath.Join(buildPath, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(page)
	for _, expected := range []string{
		`data-path="/hello"`,
		`data-path="/events"`,
		`>Hello</article>`,
		`>Events</article>`,
		`"allContent":[{`,
		`"filepath":"posts/hello.json"`,
	} {
		if !strings.Contains(html, expected) {
			t.Errorf("rendered page does not contain %q:\n%s", expected, html)
		}
	}
}
