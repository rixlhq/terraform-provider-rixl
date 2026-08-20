package rixlclient

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientAuthHeaders(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := New("api-key", "bearer-token", srv.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if _, err := c.Get(t.Context(), "/feeds", nil); err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got.Get("X-Api-Key") != "api-key" {
		t.Errorf("X-Api-Key header = %q, want %q", got.Get("X-Api-Key"), "api-key")
	}
	if got.Get("Authorization") != "Bearer bearer-token" {
		t.Errorf("Authorization header = %q, want %q", got.Get("Authorization"), "Bearer bearer-token")
	}
}

func TestClientGetQuery(t *testing.T) {
	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := New("key", "", srv.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	query := make(map[string][]string)
	query["pagination.limit"] = []string{"10"}
	query["pagination.offset"] = []string{"5"}
	if _, err := c.Get(t.Context(), "/feeds", query); err != nil {
		t.Fatalf("Get: %v", err)
	}

	if !strings.Contains(gotURL, "pagination.limit=10") {
		t.Errorf("URL %q missing pagination.limit", gotURL)
	}
	if !strings.Contains(gotURL, "pagination.offset=5") {
		t.Errorf("URL %q missing pagination.offset", gotURL)
	}
}

func TestClientPostBody(t *testing.T) {
	var body []byte
	var contentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType = r.Header.Get("Content-Type")
		body, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c, err := New("key", "", srv.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if _, err := c.Post(t.Context(), "/feeds", []byte(`{"name":"test"}`)); err != nil {
		t.Fatalf("Post: %v", err)
	}

	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", contentType)
	}
	if string(body) != `{"name":"test"}` {
		t.Errorf("body = %s, want %s", body, `{"name":"test"}`)
	}
}

func TestClientNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	}))
	defer srv.Close()

	c, err := New("key", "", srv.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = c.Get(t.Context(), "/missing", nil)
	if !IsNotFound(err) {
		t.Fatalf("expected NotFound error, got %v", err)
	}
}

func TestClientError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusBadRequest)
	}))
	defer srv.Close()

	c, err := New("key", "", srv.URL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = c.Get(t.Context(), "/bad", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if IsNotFound(err) {
		t.Fatal("expected non-NotFound error")
	}
}

func TestNewDefaultBaseURL(t *testing.T) {
	c, err := New("key", "", "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
}
