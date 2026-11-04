package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/plentico/plenti/readers"
)

func TestDataSourceTrueIndexFiles(t *testing.T) {
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

	files := []string{
		"content/_index.json",
		"content/pages/_index.json",
		"content/pages/index.json",
		"content/events/_index.json",
		"content/events/past/_index.json",
		"content/events/workshop.json",
		"content/events/_defaults.json",
		"content/pages/index._fieldtypes.json",
		"content/_index.fieldtypes.json",
	}
	for _, name := range files {
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(`{"title":"Test"}`), 0644); err != nil {
			t.Fatal(err)
		}
	}

	content, err := DataSource(readers.SiteConfig{})
	if err != nil {
		t.Fatal(err)
	}
	paths := make(map[string]string, len(content))
	for _, node := range content {
		paths[node.FilePath] = node.Path
	}
	for file, want := range map[string]string{
		"_index.json":             "/",
		"pages/_index.json":       "/pages",
		"pages/index.json":        "/index",
		"events/_index.json":      "/events",
		"events/past/_index.json": "/events/past",
		"events/workshop.json":    "/workshop",
	} {
		if got := paths[file]; got != want {
			t.Errorf("%s path = %q, want %q", file, got, want)
		}
	}
	if len(content) != len(files)-3 {
		t.Errorf("got %d content nodes, want %d; metadata files must be excluded", len(content), len(files)-2)
	}
}

func TestDataSourceFileRouteOverridesTypeRoute(t *testing.T) {
	projectDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previousDir) })

	for name, body := range map[string]string{
		"content/references.json":         `{"title":"References"}`,
		"content/references/example.json": `{"title":"Example","slug":"example"}`,
	} {
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	config := readers.SiteConfig{Routes: map[string]string{
		"references.json": "/references/page/:paginate(totalPages)",
		"references":      "/references/:fields(slug)",
	}}
	content, err := DataSource(config)
	if err != nil {
		t.Fatal(err)
	}
	paths := map[string]string{}
	for _, node := range content {
		paths[node.FilePath] = node.Path
	}
	if paths["references.json"] != "/references/page/:paginate(totalPages)" {
		t.Errorf("aggregate path = %q, want unresolved file-specific route", paths["references.json"])
	}
	if paths["references/example.json"] != "/references/example" {
		t.Errorf("detail path = %q, want type route", paths["references/example.json"])
	}
}

func TestDataSourceRouteOverrideTakesPrecedenceOverTrueIndex(t *testing.T) {
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

	name := "content/events/_index.json"
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(`{"title":"All Events"}`), 0644); err != nil {
		t.Fatal(err)
	}
	config := readers.SiteConfig{Routes: map[string]string{"events": "programs/:fields(title)"}}
	content, err := DataSource(config)
	if err != nil {
		t.Fatal(err)
	}
	if len(content) != 1 || content[0].Path != "programs/all-events" {
		t.Fatalf("route override was not preserved: %+v", content)
	}
}
