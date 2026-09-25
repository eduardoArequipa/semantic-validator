package api

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

func TestDocumentationOnlySite(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/docs/", DocsHandler())
	mux.Handle("/", HomeHandler())
	mux.HandleFunc("/openapi.yaml", OpenAPIHandler)
	get := func(path string) *httptest.ResponseRecorder {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		return w
	}
	for path, target := range map[string]string{
		"/":                "/docs/byok.html",
		"/docs":            "/docs/",
		"/docs/":           "/docs/byok.html",
		"/docs/index.html": "/docs/byok.html",
		"/docs/en.html":    "/docs/byok-en.html",
	} {
		w := get(path)
		if w.Code != http.StatusFound && !(path == "/docs" && w.Code == http.StatusMovedPermanently) {
			t.Errorf("%s: expected redirect, got %d", path, w.Code)
		}
		if got := w.Header().Get("Location"); got != target {
			t.Errorf("%s: redirect to %q, want %q", path, got, target)
		}
	}
	for _, path := range []string{"/docs/byok.html", "/docs/byok-en.html"} {
		page := get(path)
		if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "TYPESAFE_API_KEY") {
			t.Fatalf("missing SDK guide: %s", path)
		}
		for _, expected := range []string{"typescript", "python", "go", "java", "DirectValidator", "http://localhost:8080"} {
			if !strings.Contains(page.Body.String(), expected) {
				t.Errorf("%s: missing %s", path, expected)
			}
		}
		for _, expected := range []string{
			"product_description", "validator.check(", "client.Check(",
			"uncertain", "semantic-validator/sdk/python", "semantic-validator/sdk/java",
			"sdk/go@v0.3.0",
		} {
			if !strings.Contains(page.Body.String(), expected) {
				t.Errorf("%s: incomplete usage guide, missing %s", path, expected)
			}
		}
		if !strings.Contains(page.Header().Get("Content-Security-Policy"), "connect-src 'none'") {
			t.Errorf("%s: docs must not call the API", path)
		}
		for _, match := range regexp.MustCompile(`(?:href|src)="([^"]+)"`).FindAllStringSubmatch(page.Body.String(), -1) {
			link, err := url.Parse(match[1])
			if err != nil {
				t.Fatal(err)
			}
			if link.IsAbs() {
				if link.Scheme != "https" {
					t.Errorf("%s: external link must use HTTPS: %s", path, match[1])
				}
				continue
			}
			resolved := link.Path
			if resolved == "" {
				resolved = path
			}
			if !strings.HasPrefix(resolved, "/") {
				resolved = "/docs/" + resolved
			}
			target := get(resolved)
			if target.Code != http.StatusOK && !(resolved == "/" && target.Code == http.StatusFound) {
				t.Errorf("%s: broken link %s (%d)", path, match[1], target.Code)
			}
			if link.Fragment != "" && !strings.Contains(target.Body.String(), `id="`+link.Fragment+`"`) {
				t.Errorf("%s: missing anchor %s", path, match[1])
			}
			if strings.HasSuffix(resolved, ".tar.gz") || strings.HasSuffix(resolved, ".tgz") {
				checkSDKArchive(t, resolved, target.Body.Bytes())
			}
		}
	}
	for _, name := range []string{"sdk-0.1.0.tgz", "python-0.1.0.tar.gz", "go-0.1.0.tar.gz", "java-0.1.0.tar.gz"} {
		if w := get("/docs/downloads/semantic-validator-" + name); w.Code != http.StatusOK {
			t.Errorf("legacy SDK download unavailable: %s (%d)", name, w.Code)
		}
	}
	for _, path := range []string{
		"/docs/home.html", "/docs/demo.html", "/docs/demo.js", "/docs/demo.css",
		"/docs/app.js", "/docs/home.css", "/docs/downloads/semantic-validator-demo-typescript-0.1.0.tar.gz",
		"/docs/.env", "/docs/downloads/", "/docs/missing", "/docs/%2e%2e/.env",
	} {
		if w := get(path); w.Code == http.StatusOK {
			t.Errorf("old or private page still available: %s", path)
		}
	}
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/docs/byok.html", nil))
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST docs: %d", w.Code)
	}
}

func checkSDKArchive(t *testing.T, path string, data []byte) {
	t.Helper()
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	archive := tar.NewReader(gz)
	hasLicense := false
	for {
		entry, err := archive.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasSuffix(entry.Name, "/LICENSE") || entry.Name == "LICENSE" {
			hasLicense = true
		}
		for _, segment := range strings.Split(entry.Name, "/") {
			if segment == ".env" || segment == "node_modules" || segment == "__pycache__" || segment == ".." {
				t.Errorf("private archive entry %s", entry.Name)
			}
		}
	}
	if !hasLicense {
		t.Errorf("archive lacks license: %s", path)
	}
}
