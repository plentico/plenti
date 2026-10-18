package build

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/plentico/pico/pkg/pico"
	"github.com/plentico/plenti/readers"
)

// envProps are sitewide values passed to every layout.
type envProps struct {
	Local       bool              `json:"local"`
	BaseURL     string            `json:"baseurl"`
	Fingerprint string            `json:"fingerprint"`
	SiteVars    map[string]any    `json:"sitevars"`
	Routes      map[string]string `json:"routes"`
	CMS         cmsEnv            `json:"cms"`
}

type cmsEnv struct {
	Provider         string `json:"provider"`
	Repo             string `json:"repo"`
	RedirectUrl      string `json:"redirect_url"`
	AppId            string `json:"app_id"`
	Branch           string `json:"branch"`
	AuthorizationUrl string `json:"authorization_url"`
	AccessTokenUrl   string `json:"access_token_url"`
	ApiBaseUrl       string `json:"api_base_url"`
}

func getEnvProps(siteConfig readers.SiteConfig) envProps {
	return envProps{
		Local:       Local,
		BaseURL:     siteConfig.BaseURL,
		Fingerprint: siteConfig.Fingerprint,
		SiteVars:    siteConfig.SiteVars,
		Routes:      siteConfig.Routes,
		CMS: cmsEnv{
			Provider:         siteConfig.CMS.Provider,
			Repo:             siteConfig.CMS.Repo,
			RedirectUrl:      siteConfig.CMS.RedirectUrl,
			AppId:            siteConfig.CMS.AppId,
			Branch:           siteConfig.CMS.Branch,
			AuthorizationUrl: siteConfig.CMS.AuthorizationUrl,
			AccessTokenUrl:   siteConfig.CMS.AccessTokenUrl,
			ApiBaseUrl:       siteConfig.CMS.ApiBaseUrl,
		},
	}
}

// getLayoutsDir returns the layouts directory for the current build
// (the materialized theme dir when a theme is used, otherwise the project).
func getLayoutsDir() string {
	if themeTmpDir != "" {
		return filepath.Join(themeTmpDir, "layouts")
	}
	return "layouts"
}

// Pages renders static HTML pages for every content node using Pico.
func Pages(buildPath string, siteConfig readers.SiteConfig, allContent []ContentNode) error {

	defer Benchmark(time.Now(), "Rendering pages with Pico")

	Log("\nRendering static pages:")

	layoutsDir := getLayoutsDir()
	entrypoint := siteConfig.EntryPointHTML
	rootLayout := filepath.Join(layoutsDir, entrypoint)
	if _, err := os.Stat(rootLayout); err != nil {
		return fmt.Errorf("can't find entrypoint layout %s: %w", rootLayout, err)
	}

	env := getEnvProps(siteConfig)
	envMap := structToMap(env)

	pageCounter := 0
	for _, node := range allContent {
		contentLayout := filepath.Join(layoutsDir, "content", node.Type+".pico")
		if _, err := os.Stat(contentLayout); err != nil {
			Log(fmt.Sprintf("Skipping route %s, no layout found at layouts/content/%s.pico", node.Path, node.Type))
			continue
		}

		props := map[string]any{
			"env": envMap,
			// Path to the content layout, relative to the global wrapper layout.
			"contentLayout": "../content/" + node.Type + ".pico",
			"content":       structToMap(node),
			"fields":        node.Fields,
		}
		for k, v := range node.Fields {
			props[k] = v
		}

		pageHTML, err := renderPage(rootLayout, props)
		if err != nil {
			return fmt.Errorf("can't render %s: %w", node.Path, err)
		}

		dest := filepath.Join(buildPath, node.Path, "index.html")
		if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
			return fmt.Errorf("can't create folder for %s: %w", node.Path, err)
		}
		if err := os.WriteFile(dest, []byte(pageHTML), 0644); err != nil {
			return fmt.Errorf("can't write page %s: %w", dest, err)
		}
		pageCounter++
	}

	// Render a 404 page if a 404 layout exists.
	notFoundLayout := filepath.Join(layoutsDir, "content", "404.pico")
	if _, err := os.Stat(notFoundLayout); err == nil {
		props := map[string]any{
			"env":           envMap,
			"contentLayout": "../content/404.pico",
			"content":       map[string]any{"type": "404", "path": "/404"},
			"fields":        map[string]any{},
		}
		pageHTML, err := renderPage(rootLayout, props)
		if err != nil {
			return fmt.Errorf("can't render 404 page: %w", err)
		}
		Path404 = "404.html"
		if err := os.WriteFile(filepath.Join(buildPath, Path404), []byte(pageHTML), 0644); err != nil {
			return fmt.Errorf("can't write 404 page: %w", err)
		}
	}

	Log(fmt.Sprintf("Number of pages rendered: %d", pageCounter))
	return nil
}

// structToMap converts a struct into a plain map via a JSON round trip so
// Pico can serialize values into the template scope and client payload.
func structToMap(v any) map[string]any {
	b, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}
	m := map[string]any{}
	if err := json.Unmarshal(b, &m); err != nil {
		return map[string]any{}
	}
	return m
}

// renderPage runs Pico and assembles the final HTML document with the
// scoped styles, component scripts and the Pattr runtime. Pico itself
// injects the p-root-data hydration payload into the head.
func renderPage(rootLayout string, props map[string]any) (string, error) {
	markup, script, style := pico.RenderRoot(rootLayout, props)

	var headInject strings.Builder
	if strings.TrimSpace(style) != "" {
		headInject.WriteString("<style>\n" + style + "\n</style>\n")
	}

	var bodyInject strings.Builder
	bodyInject.WriteString("<script src=\"/core/pattr.js\"></script>\n")
	if strings.TrimSpace(script) != "" {
		bodyInject.WriteString("<script>\n" + script + "\n</script>\n")
	}
	if Local {
		bodyInject.WriteString("<script src=\"/core/live-reload.js\"></script>\n")
	}

	html := markup
	if headInject.Len() > 0 {
		if i := strings.Index(html, "</head>"); i != -1 {
			html = html[:i] + headInject.String() + html[i:]
		}
	}
	if i := strings.Index(html, "</body>"); i != -1 {
		html = html[:i] + bodyInject.String() + html[i:]
	} else {
		html += bodyInject.String()
	}
	return html, nil
}
