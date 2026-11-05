package build

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/plentico/plenti/defaults"
	"github.com/plentico/plenti/readers"
)

// Cms builds the admin CMS into the public build directory. Core files are
// embedded, but users can override any of them by ejecting to core/ in their
// project (see `plenti eject`).
func Cms(buildPath string, siteConfig readers.SiteConfig, allContent []ContentNode, mediaIndex []string) error {

	defer Benchmark(time.Now(), "Building CMS")

	Log("\nBuilding CMS admin:")

	coreDest := filepath.Join(buildPath, "core")

	// Copy core assets (pattr.js, cms files) into the build,
	// preferring versions the user has ejected into their project.
	err := fs.WalkDir(defaults.CoreFS, "core", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		dest := filepath.Join(buildPath, path)

		var content []byte
		// Check for an ejected override in the user's project.
		if _, statErr := os.Stat(path); statErr == nil {
			content, err = os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("can't read ejected core file %s: %w", path, err)
			}
		} else {
			content, err = defaults.CoreFS.ReadFile(path)
			if err != nil {
				return fmt.Errorf("can't read embedded core file %s: %w", path, err)
			}
		}

		if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
			return fmt.Errorf("can't create folder for %s: %w", dest, err)
		}
		if err := os.WriteFile(dest, content, 0644); err != nil {
			return fmt.Errorf("can't write core file %s: %w", dest, err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("can't copy core files: %w", err)
	}

	// The CMS is embedded in every page (admin bar); there is no standalone
	// /admin page anymore. Remove the shell template from the build output.
	os.Remove(filepath.Join(coreDest, "cms", "admin.pico"))

	// Write the data payload the CMS editor consumes (content, schemas, env).
	if err := writeCmsData(buildPath, siteConfig, allContent, mediaIndex); err != nil {
		return err
	}

	// Remove .pico sources from the public build (only used at build time).
	filepath.WalkDir(coreDest, func(path string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".pico") {
			os.Remove(path)
		}
		return nil
	})

	return nil
}

// cmsData is the payload served to the CMS editor at /admin/data.json.
type cmsContentNode struct {
	ContentNode
	SourceFields json.RawMessage `json:"sourceFields,omitempty"`
}

type cmsData struct {
	Env               envProps                   `json:"env"`
	Content           []cmsContentNode           `json:"content"`
	Schemas           map[string]map[string]any  `json:"schemas"`
	Defaults          map[string]json.RawMessage `json:"defaults"`
	Components        map[string]json.RawMessage `json:"components"`
	ComponentSchemas  map[string]map[string]any  `json:"componentSchemas"`
	ComponentPreviews map[string]string          `json:"componentPreviews"`
	Media             []string                   `json:"media"`
}

func writeCmsData(buildPath string, siteConfig readers.SiteConfig, allContent []ContentNode, mediaIndex []string) error {
	contentDir := "content"
	if themeTmpDir != "" {
		contentDir = filepath.Join(themeTmpDir, "content")
	}

	data := cmsData{
		Env:               getEnvProps(siteConfig),
		Content:           make([]cmsContentNode, len(allContent)),
		Schemas:           map[string]map[string]any{},
		Defaults:          map[string]json.RawMessage{},
		Components:        map[string]json.RawMessage{},
		ComponentSchemas:  map[string]map[string]any{},
		ComponentPreviews: map[string]string{},
		Media:             mediaIndex,
	}
	for i, item := range allContent {
		data.Content[i].ContentNode = item
		if source, err := os.ReadFile(filepath.Join(contentDir, filepath.FromSlash(item.FilePath))); err == nil && json.Valid(source) {
			data.Content[i].SourceFields = json.RawMessage(source)
		}
	}

	filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		name := d.Name()
		rel, err := filepath.Rel(contentDir, path)
		if err != nil {
			return nil
		}
		typeName := filepath.ToSlash(filepath.Dir(rel))
		if strings.HasPrefix(typeName, "_components/") && (name == "_fieldtypes.json" || name == "_defaults.json") {
			componentName := strings.TrimPrefix(typeName, "_components/")
			fileBytes, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			if name == "_defaults.json" {
				data.Components[componentName] = json.RawMessage(fileBytes)
				var props map[string]any
				if json.Unmarshal(fileBytes, &props) == nil {
					props["env"] = structToMap(data.Env)
					props["content"] = props
					layout := filepath.Join(getLayoutsDir(), "components", componentName+".pico")
					if _, err := os.Stat(layout); err == nil {
						if preview, renderErr := renderPage(layout, props); renderErr == nil {
							data.ComponentPreviews[componentName] = preview
						}
					}
				}
			} else {
				parsed := map[string]any{}
				if json.Unmarshal(fileBytes, &parsed) == nil {
					data.ComponentSchemas[componentName] = parsed
				}
			}
			return nil
		}
		kind := ""
		if name == "_fieldtypes.json" || name == "_defaults.json" {
			if typeName == "." {
				return nil
			}
			kind = strings.TrimSuffix(strings.TrimPrefix(name, "_"), ".json")
		} else if typeName == "." {
			// Single types have field types, but no defaults because they cannot
			// be duplicated: index.json -> index._fieldtypes.json. The homepage
			// convention is _index.json -> _index.fieldtypes.json.
			suffix := "._fieldtypes.json"
			if name == "_index.fieldtypes.json" {
				suffix = ".fieldtypes.json"
			}
			if !strings.HasSuffix(name, suffix) {
				return nil
			}
			typeName = strings.TrimSuffix(name, suffix)
			kind = "schema"
		} else {
			return nil
		}
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		parsed := map[string]any{}
		if err := json.Unmarshal(fileBytes, &parsed); err != nil {
			return nil
		}
		if kind == "schema" || kind == "fieldtypes" {
			data.Schemas[typeName] = parsed
		} else {
			// Keep the raw bytes so the CMS sees fields in the file's original
			// key order (a map round-trip would sort them alphabetically).
			data.Defaults[typeName] = json.RawMessage(fileBytes)
		}
		return nil
	})

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("can't marshal CMS data: %w", err)
	}
	dest := filepath.Join(buildPath, "admin", "data.json")
	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}
	if err := os.WriteFile(dest, dataJSON, 0644); err != nil {
		return fmt.Errorf("can't write CMS data: %w", err)
	}
	return nil
}
