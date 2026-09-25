package access

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPersistenceConcurrencyAndRevocation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	a, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	id, token, err := a.Create("tester", 10)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := b.Authenticate(token); err != nil || got != id {
		t.Fatal("authentication failed", err)
	}
	if _, err := b.Authenticate("wrong"); !errors.Is(err, ErrUnauthorized) {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	var accepted atomic.Int32
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			s := a
			if i%2 == 0 {
				s = b
			}
			_, err := s.Reserve(id, 1)
			if err == nil {
				accepted.Add(1)
			} else if !errors.Is(err, ErrQuota) {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	if accepted.Load() != 10 {
		t.Fatalf("accepted %d, want 10", accepted.Load())
	}
	if err := b.RecordProviderAttempt(id); err != nil {
		t.Fatal(err)
	}
	c, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	keys, err := c.List()
	if err != nil || len(keys) != 1 || keys[0].Used != 10 || keys[0].ProviderAttempts != 1 {
		t.Fatalf("usage %+v err %v", keys, err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), token) {
		t.Fatal("plaintext key persisted")
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatal("insecure permissions", err)
	}
	if err := c.Revoke(id); err != nil {
		t.Fatal(err)
	}
	if _, err := a.Authenticate(token); !errors.Is(err, ErrUnauthorized) {
		t.Fatal("revoked key accepted", err)
	}
	if _, err := a.Reserve(id, 1); !errors.Is(err, ErrUnauthorized) {
		t.Fatal(err)
	}
}

func TestDailyResetAndAtomicBatch(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "keys.json"))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 23, 23, 59, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	id, _, err := s.Create("tester", 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Reserve(id, 2); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Reserve(id, 2); !errors.Is(err, ErrQuota) {
		t.Fatal(err)
	}
	if key, err := s.Reserve(id, 1); err != nil || key.Used != 3 {
		t.Fatal("partial batch charged", err)
	}
	now = now.Add(2 * time.Minute)
	if key, err := s.Reserve(id, 3); err != nil || key.Used != 3 {
		t.Fatal("day did not reset", err)
	}
}

func TestCorruptStoreFailsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	id, token, err := s.Create("tester", 5)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Authenticate(token); err == nil {
		t.Fatal("authenticated with corrupt store")
	}
	if _, err := s.Reserve(id, 1); err == nil {
		t.Fatal("reserved with corrupt store")
	}
	if err := s.RecordProviderAttempt(id); err == nil {
		t.Fatal("recorded against corrupt store")
	}
	if _, err := Open(path); err == nil {
		t.Fatal("opened corrupt store")
	}
}
