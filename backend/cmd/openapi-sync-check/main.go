package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type routeKey struct {
	Method string
	Path   string
}

var (
	apiRouteRe      = regexp.MustCompile(`\b(api|msg)\.(GET|POST|PUT|PATCH|DELETE)\("([^"]+)"`)
	healthRouteRe   = regexp.MustCompile(`\br\.(GET|HEAD)\("(/health)"`)
	openapiPathRe   = regexp.MustCompile(`^  (/[^\s:]+):\s*$`)
	openapiMethodRe = regexp.MustCompile(`^    (get|post|put|patch|delete|head):\s*$`)
)

func main() {
	mainGo := filepath.Clean("main.go")
	openAPI := filepath.Clean("../docs/openapi.yaml")

	declared, err := extractDeclaredRoutes(mainGo)
	if err != nil {
		fatal(err)
	}
	documented, err := extractDocumentedRoutes(openAPI)
	if err != nil {
		fatal(err)
	}

	var missing []routeKey
	for _, r := range declared {
		if _, ok := documented[r]; !ok {
			missing = append(missing, r)
		}
	}
	if len(missing) == 0 {
		fmt.Println("openapi sync check passed")
		return
	}
	sort.Slice(missing, func(i, j int) bool {
		if missing[i].Path == missing[j].Path {
			return missing[i].Method < missing[j].Method
		}
		return missing[i].Path < missing[j].Path
	})
	fmt.Println("openapi sync check failed: missing route/method entries in docs/openapi.yaml")
	for _, m := range missing {
		fmt.Printf("- %s %s\n", m.Method, m.Path)
	}
	os.Exit(1)
}

func extractDeclaredRoutes(path string) ([]routeKey, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	seen := map[routeKey]struct{}{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if m := apiRouteRe.FindStringSubmatch(line); len(m) == 4 {
			group := m[1]
			method := strings.ToLower(m[2])
			rawPath := m[3]
			var fullPath string
			switch group {
			case "api":
				fullPath = "/api" + convertGinPath(rawPath)
			case "msg":
				fullPath = "/api/messages" + convertGinPath(rawPath)
			}
			seen[routeKey{Method: method, Path: fullPath}] = struct{}{}
			continue
		}
		if m := healthRouteRe.FindStringSubmatch(line); len(m) == 3 {
			seen[routeKey{Method: strings.ToLower(m[1]), Path: m[2]}] = struct{}{}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	out := make([]routeKey, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out, nil
}

func extractDocumentedRoutes(path string) (map[routeKey]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := map[routeKey]struct{}{}
	sc := bufio.NewScanner(f)
	currentPath := ""
	for sc.Scan() {
		line := sc.Text()
		if m := openapiPathRe.FindStringSubmatch(line); len(m) == 2 {
			currentPath = m[1]
			continue
		}
		if currentPath == "" {
			continue
		}
		if m := openapiMethodRe.FindStringSubmatch(line); len(m) == 2 {
			out[routeKey{Method: m[1], Path: currentPath}] = struct{}{}
			continue
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func convertGinPath(p string) string {
	if strings.TrimSpace(p) == "" {
		return p
	}
	parts := strings.Split(p, "/")
	for i := range parts {
		if strings.HasPrefix(parts[i], ":") && len(parts[i]) > 1 {
			parts[i] = "{" + parts[i][1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}

func fatal(err error) {
	fmt.Println("openapi sync check failed:", err)
	os.Exit(1)
}
