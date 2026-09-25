package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eduardoArequipa/semantic-validator/internal/access"
	"github.com/eduardoArequipa/semantic-validator/internal/api"
	"github.com/eduardoArequipa/semantic-validator/internal/auth"
	"github.com/eduardoArequipa/semantic-validator/internal/cache"
	"github.com/eduardoArequipa/semantic-validator/internal/config"
	"github.com/eduardoArequipa/semantic-validator/internal/metrics"
	"github.com/eduardoArequipa/semantic-validator/internal/providers/jev"
	"github.com/eduardoArequipa/semantic-validator/internal/ratelimit"
	"github.com/eduardoArequipa/semantic-validator/internal/rules"
	"github.com/eduardoArequipa/semantic-validator/internal/validation"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	handler, err := buildHandler(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.DocsOnly {
		log.Print("hosted validation paused; serving documentation only")
	}
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("server shutdown: %v", err)
		}
	}()

	log.Printf("semantic-validator listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("server error: %v", err)
		os.Exit(1)
	}
}

func buildHandler(cfg config.Config) (http.Handler, error) {
	observability := metrics.New()
	mux := http.NewServeMux()
	if cfg.DocsOnly {
		mux.HandleFunc("/v1", api.PausedHandler)
		mux.HandleFunc("/v1/", api.PausedHandler)
		mux.HandleFunc("/health", api.DocsOnlyHealthHandler)
	} else {
		httpClient := &http.Client{Timeout: cfg.JevTimeout}
		provider := jev.NewClientWithMetrics(httpClient, cfg.JevBaseURL, cfg.JevAPIKey, cfg.JevModel, observability)
		service := validation.NewServiceWithCache(rules.DefaultRegistry(), provider, cache.NewMemoryWithMetrics(observability))
		limiter := ratelimit.NewWithTrustedProxy(60, time.Minute, cfg.TrustProxyHeaders)

		protectedMux := http.NewServeMux()
		protectedMux.Handle("/v1/validate", api.NewValidateHandler(service))
		protectedMux.Handle("/v1/check", api.NewCheckHandler(service))
		protectedMux.Handle("/v1/validate/batch", api.NewBatchHandler(service))

		var protected http.Handler = auth.Middleware(cfg.SemanticValidatorKey, limiter.Middleware(protectedMux))
		if cfg.APIKeysFile != "" {
			store, err := access.Open(cfg.APIKeysFile)
			if err != nil {
				return nil, fmt.Errorf("access store unavailable: %w", err)
			}
			protected = store.Middleware(limiter.Middleware(protectedMux))
		}
		mux.Handle("/v1/", protected)
		mux.HandleFunc("/health", api.HealthHandler)
	}
	mux.HandleFunc("/metrics", observability.Handler)
	mux.HandleFunc("/openapi.yaml", api.OpenAPIHandler)
	mux.Handle("/docs/", api.DocsHandler())
	mux.Handle("/", api.HomeHandler())
	return observability.Middleware(mux), nil
}
