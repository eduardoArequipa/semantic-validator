package api

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

// Embed only the SDK guides and installable SDK packages. The old beta site is not served.
//
//go:embed docs/byok.html docs/byok-en.html docs/style.css docs/downloads/LICENSE docs/downloads/semantic-validator-sdk-0.1.0.tgz docs/downloads/semantic-validator-python-0.1.0.tar.gz docs/downloads/semantic-validator-go-0.1.0.tar.gz docs/downloads/semantic-validator-java-0.1.0.tar.gz docs/downloads/semantic-validator-sdk-0.2.0.tgz docs/downloads/semantic-validator-python-0.2.0.tar.gz docs/downloads/semantic-validator-go-0.2.0.tar.gz docs/downloads/semantic-validator-java-0.2.0.tar.gz
var documentation embed.FS

// HomeHandler sends visitors straight to the SDK documentation.
func HomeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		http.Redirect(w, r, "/docs/byok.html", http.StatusFound)
	})
}

// DocsHandler serves only the embedded, public documentation bundle.
func DocsHandler() http.Handler {
	files, err := fs.Sub(documentation, "docs")
	if err != nil {
		panic(err)
	}
	server := http.StripPrefix("/docs/", http.FileServer(http.FS(files)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		path := strings.TrimPrefix(r.URL.Path, "/docs/")
		switch path {
		case "", "index.html", "byok.html":
			if path != "byok.html" {
				http.Redirect(w, r, "/docs/byok.html", http.StatusFound)
				return
			}
		case "en.html", "byok-en.html":
			if path != "byok-en.html" {
				http.Redirect(w, r, "/docs/byok-en.html", http.StatusFound)
				return
			}
		case "style.css", "downloads/LICENSE",
			"downloads/semantic-validator-sdk-0.1.0.tgz",
			"downloads/semantic-validator-python-0.1.0.tar.gz",
			"downloads/semantic-validator-go-0.1.0.tar.gz",
			"downloads/semantic-validator-java-0.1.0.tar.gz",
			"downloads/semantic-validator-sdk-0.2.0.tgz",
			"downloads/semantic-validator-python-0.2.0.tar.gz",
			"downloads/semantic-validator-go-0.2.0.tar.gz",
			"downloads/semantic-validator-java-0.2.0.tar.gz":
		default:
			http.NotFound(w, r)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'none'; style-src 'self'; img-src 'self'; connect-src 'none'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'")
		server.ServeHTTP(w, r)
	})
}
