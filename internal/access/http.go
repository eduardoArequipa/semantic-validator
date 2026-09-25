package access

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type identity struct {
	store *Store
	id    string
}
type contextKey struct{}

var ErrAccounting = errors.New("usage accounting unavailable")

func reply(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func failure(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrUnauthorized):
		w.Header().Set("WWW-Authenticate", "Bearer")
		reply(w, 401, "unauthorized", "valid Semantic Validator API key required")
	case errors.Is(err, ErrQuota):
		now := time.Now().UTC()
		reset := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC)
		w.Header().Set("Retry-After", strconv.Itoa(int(reset.Sub(now).Seconds())+1))
		reply(w, 429, "quota_exceeded", "daily validation quota exceeded")
	default:
		reply(w, 503, "usage_unavailable", "access and usage storage unavailable")
	}
}

func (s *Store) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			failure(w, ErrUnauthorized)
			return
		}
		id, err := s.Authenticate(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			failure(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), contextKey{}, identity{s, id})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ReserveRequest is called after decoding, before any batch work starts.
// Legacy single-key mode has no identity and retains its existing behavior.
func ReserveRequest(w http.ResponseWriter, r *http.Request, units int) bool {
	actor, ok := r.Context().Value(contextKey{}).(identity)
	if !ok {
		return true
	}
	key, err := actor.store.Reserve(actor.id, units)
	if err != nil {
		failure(w, err)
		return false
	}
	w.Header().Set("X-Quota-Limit", strconv.Itoa(key.DailyLimit))
	w.Header().Set("X-Quota-Remaining", strconv.Itoa(key.DailyLimit-key.Used))
	return true
}

// CountProviderAttempt persists an attempted inference before contacting Jev.
func CountProviderAttempt(ctx context.Context) error {
	actor, ok := ctx.Value(contextKey{}).(identity)
	if !ok {
		return nil
	}
	if err := actor.store.RecordProviderAttempt(actor.id); err != nil {
		return ErrAccounting
	}
	return nil
}
