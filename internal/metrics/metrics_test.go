package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsHandler(t *testing.T) {
	metrics := New()
	metrics.ObserveRequest(10 * time.Millisecond)
	metrics.ObserveJevRequest(nil)
	metrics.ObserveCacheHit()
	metrics.ObserveCacheMiss()

	recorder := httptest.NewRecorder()
	metrics.Handler(recorder, httptest.NewRequest("GET", "/metrics", nil))
	body := recorder.Body.String()
	for _, expected := range []string{
		"semantic_validator_requests_total 1",
		"semantic_validator_jev_requests_total 1",
		"semantic_validator_cache_hits_total 1",
		"semantic_validator_cache_misses_total 1",
	} {
		if !strings.Contains(body, expected) {
			t.Fatalf("metrics missing %q: %s", expected, body)
		}
	}
}
