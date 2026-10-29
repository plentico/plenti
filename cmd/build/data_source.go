package build

import (
	"bytes"

	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/plentico/plenti/readers"
)

// Local is set to true when using dev webserver, otherwise bools default to false.
var Local bool

// Path404 stores path to 404 html fallback for use in local webserver
var Path404 string

var (
	// Setup regex to find field name.
	reField = regexp.MustCompile(`:fields\((.*?)\)`)
	// Create regex for allowed characters when slugifying path.
	reSlugify = regexp.MustCompile(`[^a-z0-9./*]+`)
)

// ContentNode holds info related to a particular content node.
type ContentNode struct {
	Type     string         `json:"type"`
	Path     string         `json:"path"`
	FilePath string         `json:"filepath"`
	FileName string         `json:"filename"`
	Fields   map[string]any `json:"fields"`
	Pager    int            `json:"pager,omitempty"`
	// FieldOrder preserves the key order of the source JSON file (maps
	// marshal alphabetically) so the CMS can display and save fields in
	// the order the author wrote them.
	FieldOrder []string `json:"fieldOrder,omitempty"`
}

// jsonKeyOrder returns the top-level keys of a JSON object in document order.
func jsonKeyOrder(data []byte) []string {
	order := []string{}
	dec := json.NewDecoder(bytes.NewReader(data))
	// Consume opening '{'.
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return order
	}
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			break
		}
		if key, ok := t.(string); ok {
			order = append(order, key)
		}
		skipJSONValue(dec)
	}
	return order
}

// skipJSONValue consumes one JSON value (recursively for objects/arrays).
func skipJSONValue(dec *json.Decoder) {
	t, err := dec.Token()
	if err != nil {
		return
	}
	d, ok := t.(json.Delim)
	if !ok || (d != '{' && d != '[') {
		return
	}
	for dec.More() {
		if d == '{' {
			if _, err := dec.Token(); err != nil { // key
				return
			}
		}
		skipJSONValue(dec)
	}
	dec.Token() // closing delimiter
}

// DataSource builds a list of content nodes from the "content/" directory.
func DataSource(siteConfig readers.SiteConfig) ([]ContentNode, error) {

	defer Benchmark(time.Now(), "Creating data_source")

	Log("\nGathering data source from 'content/' folder")

	contentDir := "content"
	if ThemeFs != nil {
		// Materialize the merged theme+project filesystem to a temp dir so
		// Pico (which reads from disk) can access layouts and content.
		tmpDir, err := MaterializeThemeFs()
		if err != nil {
			return nil, err
		}
		contentDir = filepath.Join(tmpDir, "content")
	}

	allContent := []ContentNode{}

	err := filepath.WalkDir(contentDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("can't stat %s: %w", path, err)
		}
		if d.IsDir() {
			// Skip special folders (e.g. content/_components).
			if strings.HasPrefix(d.Name(), "_") {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".json" || strings.HasPrefix(d.Name(), "_") {
			return nil
		}

		relPath, err := filepath.Rel(contentDir, path)
		if err != nil {
			return fmt.Errorf("can't determine relative path for %s: %w", path, err)
		}
		relPath = filepath.ToSlash(relPath)

		// Read and parse the content file.
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("can't read content file %s: %w", path, err)
		}
		fields := map[string]any{}
		if err := json.Unmarshal(fileBytes, &fields); err != nil {
			return fmt.Errorf("can't parse content file %s: %w", path, err)
		}

		parts := strings.Split(strings.TrimSuffix(relPath, ".json"), "/")
		fileName := parts[len(parts)-1]

		// Determine content type. Nested under a folder = folder name is the type.
		// Top level json files are "single" types named after the file.
		contentType := fileName
		if len(parts) > 1 {
			contentType = parts[0]
		}

		node := ContentNode{
			Type:       contentType,
			FilePath:   relPath,
			FileName:   fileName,
			Fields:     fields,
			FieldOrder: jsonKeyOrder(fileBytes),
		}

		// Check for a route override for this content type.
		route, routeOverridden := siteConfig.Routes[contentType]
		if routeOverridden {
			node.Path = resolveRoute(route, fields)
		}
		if !routeOverridden {
			// Default path is the file path without the leading type folder
			// (for multi types) or the file name (for single types).
			if len(parts) > 1 {
				node.Path = "/" + strings.Join(parts[1:], "/")
			} else {
				node.Path = "/" + fileName
			}
		}
		// "index" files map to the root of their path.
		if strings.HasSuffix(node.Path, "/index") {
			node.Path = strings.TrimSuffix(node.Path, "index")
		} else if node.Path == "/index" {
			node.Path = "/"
		}

		allContent = append(allContent, node)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("can't walk content directory: %w", err)
	}

	Log(fmt.Sprintf("Number of content nodes found: %d", len(allContent)))
	return allContent, nil
}

// resolveRoute replaces :fields(...) tokens in a route with slugified field values.
func resolveRoute(route string, fields map[string]any) string {
	resolved := route
	for _, match := range reField.FindAllStringSubmatch(route, -1) {
		fieldName := match[1]
		fieldValue := ""
		if val, ok := fields[fieldName]; ok {
			fieldValue = fmt.Sprintf("%v", val)
		}
		resolved = strings.Replace(resolved, match[0], slugify(fieldValue), 1)
	}
	return resolved
}

// slugify creates a url-safe path segment.
func slugify(path string) string {
	return reSlugify.ReplaceAllString(strings.ToLower(path), "-")
}
