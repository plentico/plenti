package build

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/plentico/plenti/readers"
)

func TestWriteCmsDataLoadsCollectionAndSingleTypeMetadata(t *testing.T) {
	projectDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(previousDir) })

	files := map[string]string{
		"content/pages/_defaults.json":   `{"title":"Default page"}`,
		"content/pages/_fieldtypes.json": `{"title":"text"}`,
		"content/index.json":             `{"title":"Home"}`,
		"content/index._fieldtypes.json": `{"title":"textarea"}`,
		"content/_index.json":            `{"title":"Special home"}`,
		"content/_index.fieldtypes.json": `{"title":"markdown"}`,
	}
	for name, body := range files {
		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(name, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}

	buildDir := filepath.Join(projectDir, "public")
	if err := writeCmsData(buildDir, readers.SiteConfig{}, nil, nil); err != nil {
		t.Fatal(err)
	}
	fileBytes, err := os.ReadFile(filepath.Join(buildDir, "admin", "data.json"))
	if err != nil {
		t.Fatal(err)
	}
	var data cmsData
	if err := json.Unmarshal(fileBytes, &data); err != nil {
		t.Fatal(err)
	}

	for typeName, want := range map[string]string{
		"pages":  "text",
		"index":  "textarea",
		"_index": "markdown",
	} {
		if got := data.Schemas[typeName]["title"]; got != want {
			t.Errorf("%s schema title = %v, want %q", typeName, got, want)
		}
	}
	for typeName, want := range map[string]string{
		"pages": "Default page",
	} {
		var fields map[string]any
		if err := json.Unmarshal(data.Defaults[typeName], &fields); err != nil {
			t.Fatal(err)
		}
		if got := fields["title"]; got != want {
			t.Errorf("%s default title = %v, want %q", typeName, got, want)
		}
	}
	if len(data.Defaults) != 1 {
		t.Errorf("got %d defaults, want only the collection defaults", len(data.Defaults))
	}
}
