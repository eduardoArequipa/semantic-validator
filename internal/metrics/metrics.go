package metrics

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type Metrics struct {
	requests             atomic.Uint64
	requestLatencyNanos  atomic.Uint64
	jevRequests          atomic.Uint64
	jevErrors            atomic.Uint64
	cacheHits            atomic.Uint64
	cacheMisses          atomic.Uint64
}

type Snapshot struct {
	Requests            uint64
	RequestLatencyNanos uint64
	JevRequests         uint64
	JevErrors           uint64
	CacheHits           uint64
	CacheMisses         uint64
}

func New() *Metrics {
	return &Metrics{}
}

func (m *Metrics) ObserveRequest(duration time.Duration) {
	m.requests.Add(1)
	m.requestLatencyNanos.Add(uint64(duration))
}

func (m *Metrics) ObserveJevRequest(err error) {
	m.jevRequests.Add(1)
	if err != nil {
		m.jevErrors.Add(1)
	}
}

func (m *Metrics) ObserveCacheHit() {
	m.cacheHits.Add(1)
}

func (m *Metrics) ObserveCacheMiss() {
	m.cacheMisses.Add(1)
}

func (m *Metrics) Snapshot() Snapshot {
	return Snapshot{
		Requests:            m.requests.Load(),
		RequestLatencyNanos: m.requestLatencyNanos.Load(),
		JevRequests:         m.jevRequests.Load(),
		JevErrors:           m.jevErrors.Load(),
		CacheHits:           m.cacheHits.Load(),
		CacheMisses:         m.cacheMisses.Load(),
	}
}

func (m *Metrics) Handler(w http.ResponseWriter, _ *http.Request) {
	snapshot := m.Snapshot()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "semantic_validator_requests_total %d\n", snapshot.Requests)
	fmt.Fprintf(w, "semantic_validator_request_latency_seconds_sum %.9f\n", float64(snapshot.RequestLatencyNanos)/float64(time.Second))
	fmt.Fprintf(w, "semantic_validator_jev_requests_total %d\n", snapshot.JevRequests)
	fmt.Fprintf(w, "semantic_validator_jev_errors_total %d\n", snapshot.JevErrors)
	fmt.Fprintf(w, "semantic_validator_cache_hits_total %d\n", snapshot.CacheHits)
	fmt.Fprintf(w, "semantic_validator_cache_misses_total %d\n", snapshot.CacheMisses)
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		m.ObserveRequest(time.Since(started))
	})
}
