package cache

import (
	"sync"
	"testing"
	"time"
)

func TestSetThenGetReturnsData(t *testing.T) {
	c := New(t.TempDir())
	c.Set("greeting", []byte(`"hello"`), 5*time.Minute)

	data, ok := c.Get("greeting")
	if !ok {
		t.Fatal("expected ok=true, got false")
	}
	if string(data) != `"hello"` {
		t.Fatalf("expected %q, got %q", `"hello"`, string(data))
	}
}

func TestGetExpiredKeyReturnsFalse(t *testing.T) {
	c := New(t.TempDir())
	c.Set("short", []byte(`"bye"`), 1*time.Millisecond)

	time.Sleep(5 * time.Millisecond)

	_, ok := c.Get("short")
	if ok {
		t.Fatal("expected ok=false for expired key")
	}
}

func TestGetMissingKeyReturnsFalse(t *testing.T) {
	c := New(t.TempDir())

	_, ok := c.Get("nonexistent")
	if ok {
		t.Fatal("expected ok=false for missing key")
	}
}

func TestSetOverwritesExisting(t *testing.T) {
	c := New(t.TempDir())
	c.Set("key", []byte(`"v1"`), 5*time.Minute)
	c.Set("key", []byte(`"v2"`), 5*time.Minute)

	data, ok := c.Get("key")
	if !ok {
		t.Fatal("expected ok=true, got false")
	}
	if string(data) != `"v2"` {
		t.Fatalf("expected %q, got %q", `"v2"`, string(data))
	}
}

func TestConcurrentAccessSafety(t *testing.T) {
	c := New(t.TempDir())
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Set("shared", []byte(`"data"`), 5*time.Minute)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.Get("shared")
		}()
	}

	wg.Wait()

	// After all goroutines finish, key should be readable
	data, ok := c.Get("shared")
	if !ok {
		t.Fatal("expected ok=true after concurrent writes")
	}
	if string(data) != `"data"` {
		t.Fatalf("expected %q, got %q", `"data"`, string(data))
	}
}
