// Package access provides a single-host, persistent key and quota store.
package access

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

var ErrUnauthorized = errors.New("invalid or revoked API key")
var ErrQuota = errors.New("daily validation quota exceeded")

type Key struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Hash             string `json:"hash"`
	DailyLimit       int    `json:"daily_limit"`
	Revoked          bool   `json:"revoked"`
	Day              string `json:"day"`
	Used             int    `json:"used"`
	ProviderAttempts int    `json:"provider_attempts"`
}

type Store struct {
	path string
	mu   sync.Mutex
	now  func() time.Time
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, now: time.Now}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	err := s.transaction(func(keys map[string]*Key) error { return nil }, false)
	return s, err
}

// A separate lock file coordinates the server and administration CLI across
// atomic data-file replacements. This store targets Linux, on a local disk.
func (s *Store) transaction(fn func(map[string]*Key) error, write bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	lock, err := os.OpenFile(s.path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer lock.Close()
	if err = syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	keys := make(map[string]*Key)
	data, err := os.ReadFile(s.path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err == nil {
		if err = json.Unmarshal(data, &keys); err != nil {
			return fmt.Errorf("invalid access store")
		}
		if keys == nil {
			return errors.New("invalid access store")
		}
		for id, key := range keys {
			if key == nil || key.ID != id || key.DailyLimit <= 0 || key.Used < 0 || key.ProviderAttempts < 0 || len(key.Hash) != 64 {
				return errors.New("invalid access store")
			}
		}
	}
	if err = fn(keys); err != nil || !write {
		return err
	}
	data, err = json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".access-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), s.path); err != nil {
		return err
	}
	dir, err := os.Open(filepath.Dir(s.path))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func digest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *Store) Create(name string, limit int) (string, string, error) {
	if name == "" || limit <= 0 {
		return "", "", errors.New("name and positive daily limit required")
	}
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return "", "", err
	}
	token := "sv_live_" + hex.EncodeToString(secret[:])
	id := hex.EncodeToString(secret[:8])
	err := s.transaction(func(keys map[string]*Key) error {
		if _, exists := keys[id]; exists {
			return errors.New("key ID collision")
		}
		keys[id] = &Key{ID: id, Name: name, Hash: digest(token), DailyLimit: limit}
		return nil
	}, true)
	return id, token, err
}

func (s *Store) Revoke(id string) error {
	return s.transaction(func(keys map[string]*Key) error {
		key, ok := keys[id]
		if !ok {
			return errors.New("key not found")
		}
		key.Revoked = true
		return nil
	}, true)
}

func (s *Store) reset(key *Key) {
	day := s.now().UTC().Format("2006-01-02")
	if key.Day != day {
		key.Day = day
		key.Used = 0
		key.ProviderAttempts = 0
	}
}

func (s *Store) List() ([]Key, error) {
	result := []Key{}
	err := s.transaction(func(keys map[string]*Key) error {
		for _, key := range keys {
			copy := *key
			s.reset(&copy)
			copy.Hash = ""
			result = append(result, copy)
		}
		return nil
	}, false)
	return result, err
}

func (s *Store) Authenticate(token string) (string, error) {
	id := ""
	hash := digest(token)
	err := s.transaction(func(keys map[string]*Key) error {
		for _, key := range keys {
			if key.Hash == hash && !key.Revoked {
				id = key.ID
				return nil
			}
		}
		return ErrUnauthorized
	}, false)
	return id, err
}

func (s *Store) Reserve(id string, units int) (Key, error) {
	var result Key
	err := s.transaction(func(keys map[string]*Key) error {
		key, ok := keys[id]
		if !ok || key.Revoked {
			return ErrUnauthorized
		}
		s.reset(key)
		if units < 1 || units > key.DailyLimit-key.Used {
			return ErrQuota
		}
		key.Used += units
		result = *key
		result.Hash = ""
		return nil
	}, true)
	return result, err
}

func (s *Store) RecordProviderAttempt(id string) error {
	return s.transaction(func(keys map[string]*Key) error {
		key, ok := keys[id]
		if !ok || key.Revoked {
			return ErrUnauthorized
		}
		s.reset(key)
		key.ProviderAttempts++
		return nil
	}, true)
}
