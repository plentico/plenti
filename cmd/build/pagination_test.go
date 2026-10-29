package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/plentico/plenti/readers"
)

func TestPagesPagination(t *testing.T) {
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
	contentLayout := filepath.Join("layouts", "content", "pages.pico")
	for _, dir := range []string{filepath.Dir(rootLayout), filepath.Dir(contentLayout)} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	root := `---
prop content;
---
<html><head></head><body><main data-path="{content.path}" data-pager="{content.pager}"></main></body></html>`
	if err := os.WriteFile(rootLayout, []byte(root), 0644); err != nil {
		t.Fatal(err)
	}
	contentTemplate := `---
prop content;
prop allContent;
let postsPerPage = 2;
let posts = allContent.filter(node => node.type === "posts");
let totalPages = Math.ceil(posts.length / postsPerPage);
---
<div>{content.pager}</div>`
	if err := os.WriteFile(contentLayout, []byte(contentTemplate), 0644); err != nil {
		t.Fatal(err)
	}

	allContent := []ContentNode{{Type: "pages", Path: "/blog", Fields: map[string]any{}}}
	for i := 1; i <= 5; i++ {
		allContent = append(allContent, ContentNode{
			Type: "posts", Path: fmt.Sprintf("/post-%d", i),
			Fields: map[string]any{"title": fmt.Sprintf("Post %d", i)},
		})
	}
	config := readers.SiteConfig{
		BaseLayout: "global/html.pico",
		Routes:     map[string]string{"pages": "blog/:paginate(totalPages)"},
	}
	buildPath := filepath.Join(projectDir, "public")
	if err := Pages(buildPath, config, allContent); err != nil {
		t.Fatal(err)
	}

	for page := 1; page <= 3; page++ {
		route := "blog"
		if page > 1 {
			route = fmt.Sprintf("blog/%d", page)
		}
		pageBytes, err := os.ReadFile(filepath.Join(buildPath, route, "index.html"))
		if err != nil {
			t.Fatal(err)
		}
		html := string(pageBytes)
		expected := fmt.Sprintf(`data-pager="%d"`, page)
		if !strings.Contains(html, expected) {
			t.Errorf("page %d does not contain %q:\n%s", page, expected, html)
		}
		if !strings.Contains(html, "p-external-props") || strings.Contains(html, `"allContent":[{`) {
			t.Errorf("page %d did not externalize its allContent hydration data", page)
		}
	}
	shared, err := os.ReadFile(filepath.Join(buildPath, "core", "allContent.json"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(shared), `"type":"posts"`) < 5 {
		t.Error("shared allContent resource does not retain the complete collection")
	}
	if _, err := os.Stat(filepath.Join(buildPath, "blog", "1", "index.html")); !os.IsNotExist(err) {
		t.Error("page one should only be rendered at its clean base route")
	}
}

func TestPageCountRejectsInvalidValues(t *testing.T) {
	for _, value := range []any{nil, "3", 0, -1, 1.5} {
		if _, err := pageCount(value, "totalPages"); err == nil {
			t.Errorf("expected page count %#v to fail", value)
		}
	}
	for _, value := range []any{1, int64(2), float64(3)} {
		if _, err := pageCount(value, "totalPages"); err != nil {
			t.Errorf("expected page count %#v to succeed: %v", value, err)
		}
	}
}
