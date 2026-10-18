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

	// Copy core assets (pattr.js, live-reload.js, cms files) into the build,
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

	// Render the CMS admin shell with Pico.
	adminLayout := filepath.Join(coreDest, "cms", "admin.pico")
	if _, err := os.Stat(adminLayout); err == nil {
		props := map[string]any{"env": structToMap(getEnvProps(siteConfig))}
		pageHTML, err := renderPage(adminLayout, props)
		if err != nil {
			return fmt.Errorf("can't render CMS admin: %w", err)
		}
		adminDest := filepath.Join(buildPath, "admin", "index.html")
		if err := os.MkdirAll(filepath.Dir(adminDest), os.ModePerm); err != nil {
			return err
		}
		if err := os.WriteFile(adminDest, []byte(pageHTML), 0644); err != nil {
			return fmt.Errorf("can't write CMS admin page: %w", err)
		}
	}

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
type cmsData struct {
	Env      envProps                  `json:"env"`
	Content  []ContentNode             `json:"content"`
	Schemas  map[string]map[string]any `json:"schemas"`
	Defaults map[string]map[string]any `json:"defaults"`
	Media    []string                  `json:"media"`
}

func writeCmsData(buildPath string, siteConfig readers.SiteConfig, allContent []ContentNode, mediaIndex []string) error {
	contentDir := "content"
	if themeTmpDir != "" {
		contentDir = filepath.Join(themeTmpDir, "content")
	}

	data := cmsData{
		Env:      getEnvProps(siteConfig),
		Content:  allContent,
		Schemas:  map[string]map[string]any{},
		Defaults: map[string]map[string]any{},
		Media:    mediaIndex,
	}

	filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		name := d.Name()
		if name != "_schema.json" && name != "_defaults.json" {
			return nil
		}
		rel, err := filepath.Rel(contentDir, path)
		if err != nil {
			return nil
		}
		typeName := filepath.ToSlash(filepath.Dir(rel))
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		parsed := map[string]any{}
		if err := json.Unmarshal(fileBytes, &parsed); err != nil {
			return nil
		}
		if name == "_schema.json" {
			data.Schemas[typeName] = parsed
		} else {
			data.Defaults[typeName] = parsed
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
