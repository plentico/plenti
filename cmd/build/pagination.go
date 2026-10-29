package build

import (
	"fmt"
	"math"
	"path"
	"regexp"
	"strings"

	"github.com/plentico/plenti/readers"
)

var rePaginate = regexp.MustCompile(`:paginate\(([^)]+)\)`)

type paginationRoute struct {
	route    string
	variable string
}

func paginationForType(siteConfig readers.SiteConfig, contentType string) paginationRoute {
	route := siteConfig.Routes[contentType]
	match := rePaginate.FindStringSubmatch(route)
	if len(match) == 2 {
		return paginationRoute{route: route, variable: strings.TrimSpace(match[1])}
	}
	return paginationRoute{}
}

func (p paginationRoute) path(page int) string {
	replacement := ""
	if page > 1 {
		replacement = fmt.Sprint(page)
	}
	route := rePaginate.ReplaceAllString(p.route, replacement)
	return path.Clean("/" + strings.TrimPrefix(route, "/"))
}

func pageCount(value any, variable string) (int, error) {
	var count float64
	switch value := value.(type) {
	case int:
		count = float64(value)
	case int64:
		count = float64(value)
	case float64:
		count = value
	default:
		return 0, fmt.Errorf("%q must evaluate to a positive integer, got %T", variable, value)
	}
	if math.IsNaN(count) || math.IsInf(count, 0) || count < 1 || count != math.Trunc(count) {
		return 0, fmt.Errorf("%q must evaluate to a positive integer, got %v", variable, value)
	}
	if count > float64(math.MaxInt) {
		return 0, fmt.Errorf("%q is too large: %v", variable, value)
	}
	return int(count), nil
}
