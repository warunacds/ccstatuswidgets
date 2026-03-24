// Package cache provides a file-based TTL cache for widget output fallback.
package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Cache stores JSON files in a directory with per-key expiry.
type Cache struct {
	dir string
}

type entry struct {
	Data      json.RawMessage `json:"data"`
	ExpiresAt int64           `json:"expires_at"`
}

// New creates a cache backed by the given directory.
// The directory is created if it does not exist.
func New(dir string) *Cache {
	_ = os.MkdirAll(dir, 0o755)
	return &Cache{dir: dir}
}

// Get returns the cached data for key if it exists and has not expired.
// Returns (nil, false) if the key is missing or expired.
func (c *Cache) Get(key string) ([]byte, bool) {
	path := c.path(key)

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var e entry
	if err := json.Unmarshal(raw, &e); err != nil {
		return nil, false
	}

	if time.Now().Unix() >= e.ExpiresAt {
		_ = os.Remove(path)
		return nil, false
	}

	return []byte(e.Data), true
}

// Set writes value under key with the given TTL.
// The write is atomic: data is written to a temp file then renamed.
func (c *Cache) Set(key string, value []byte, ttl time.Duration) {
	e := entry{
		Data:      json.RawMessage(value),
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}

	raw, err := json.Marshal(e)
	if err != nil {
		return
	}

	path := c.path(key)

	tmp, err := os.CreateTemp(c.dir, ".tmp-*")
	if err != nil {
		return
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return
	}

	_ = os.Rename(tmpName, path)
}

func (c *Cache) path(key string) string {
	return filepath.Join(c.dir, key+".json")
}
