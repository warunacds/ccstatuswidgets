package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGet_ReturnsBodyOnSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello world"))
	}))
	defer srv.Close()

	c := New()
	body, err := c.Get(srv.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(body) != "hello world" {
		t.Fatalf("expected %q, got %q", "hello world", string(body))
	}
}

func TestGet_ReturnsErrorOnServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := New()
	_, err := c.Get(srv.URL)
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
	if err.Error() != "HTTP 500" {
		t.Fatalf("expected error %q, got %q", "HTTP 500", err.Error())
	}
}

func TestGet_ReturnsErrorOnConnectionTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := &Client{
		httpClient: &http.Client{Timeout: 50 * time.Millisecond},
		userAgent:  "ccw/0.1",
	}
	_, err := c.Get(srv.URL)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestGet_SetsUserAgentHeader(t *testing.T) {
	var receivedUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	c := New()
	_, err := c.Get(srv.URL)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if receivedUA != "ccw/0.1" {
		t.Fatalf("expected User-Agent %q, got %q", "ccw/0.1", receivedUA)
	}
}
