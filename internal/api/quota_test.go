package api

import (
	"github.com/eduardoArequipa/semantic-validator/internal/access"
	"github.com/eduardoArequipa/semantic-validator/internal/cache"
	"github.com/eduardoArequipa/semantic-validator/internal/rules"
	"github.com/eduardoArequipa/semantic-validator/internal/validation"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestIndividualQuotaBatchCacheAndRevocation(t *testing.T) {
	store, err := access.Open(filepath.Join(t.TempDir(), "keys.json"))
	if err != nil {
		t.Fatal(err)
	}
	id, token, err := store.Create("first", 3)
	if err != nil {
		t.Fatal(err)
	}
	_, other, err := store.Create("second", 3)
	if err != nil {
		t.Fatal(err)
	}
	service := validation.NewServiceWithCache(rules.DefaultRegistry(), handlerProvider{}, cache.NewMemoryWithMetrics(nil))
	mux := http.NewServeMux()
	mux.Handle("/v1/validate", NewValidateHandler(service))
	mux.Handle("/v1/check", NewCheckHandler(service))
	mux.Handle("/v1/validate/batch", NewBatchHandler(service))
	handler := store.Middleware(mux)
	call := func(key, path, body string, want int) *httptest.ResponseRecorder {
		t.Helper()
		r := httptest.NewRequest("POST", path, strings.NewReader(body))
		r.Header.Set("Authorization", "Bearer "+key)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != want {
			t.Fatalf("%s: status %d want %d: %s", path, w.Code, want, w.Body.String())
		}
		return w
	}
	const single = `{"rule":"person_name","value":"Jorge Eduardo"}`
	const batch = `{"items":[{"id":"a","rule":"person_name","value":"Ana"},{"id":"b","rule":"person_name","value":"Pedro"}]}`
	call(token, "/v1/validate", `{`, 400)
	call("wrong", "/v1/validate", single, 401)
	call(token, "/v1/validate", single, 200)
	w := call(token, "/v1/validate", single, 200)
	if w.Header().Get("X-Quota-Remaining") != "1" {
		t.Fatal(w.Header())
	}
	w = call(token, "/v1/validate/batch", batch, 429)
	if !strings.Contains(w.Body.String(), "quota_exceeded") || w.Header().Get("Retry-After") == "" {
		t.Fatal(w.Body.String())
	}
	call(token, "/v1/check", `{"value":"Quiero comprar","question":"¿Quiere comprar?"}`, 200)
	call(token, "/v1/validate", single, 429)
	call(other, "/v1/validate/batch", batch, 200)
	keys, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.ID == id && (key.Used != 3 || key.ProviderAttempts != 2) {
			t.Fatalf("usage %+v", key)
		}
	}
	if err := store.Revoke(id); err != nil {
		t.Fatal(err)
	}
	call(token, "/v1/validate", single, 401)
}
