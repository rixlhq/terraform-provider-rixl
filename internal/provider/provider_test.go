package provider

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rixlhq/rixl-go/sdk"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestUserAgentTransportDoesNotRewriteURL(t *testing.T) {
	originalURL := "https://presigned.example.com/upload?signature=abc"

	var got *http.Request
	inner := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		got = req
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
			Header:     make(http.Header),
		}, nil
	})

	client := &http.Client{Transport: &userAgentTransport{inner: inner, ua: "terraform-provider-rixl"}}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, originalURL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if got == nil {
		t.Fatalf("request was not sent through transport")
	}
	if got.Header.Get("User-Agent") != "terraform-provider-rixl" {
		t.Fatalf("expected User-Agent header, got %q", got.Header.Get("User-Agent"))
	}
	if got.URL.String() != originalURL {
		t.Fatalf("presigned URL was rewritten: got %q, want %q", got.URL.String(), originalURL)
	}
}

func TestBaseURLRewriter(t *testing.T) {
	var gotPath, gotHost string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotHost = r.Host
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	opt, err := baseURLRewriter(srv.URL)
	if err != nil {
		t.Fatalf("baseURLRewriter: %v", err)
	}

	client, err := sdk.New("test-key", sdk.WithHTTPClient(newHTTPClient()), opt)
	if err != nil {
		t.Fatalf("sdk.New: %v", err)
	}

	_, _ = client.APIKeys.ListApiKeys(context.Background(), "org-1", nil)

	wantPath := "/organizations/org-1/api-keys/v1"
	if gotPath != wantPath {
		t.Fatalf("request path mismatch: got %q, want %q", gotPath, wantPath)
	}

	srvHost, _ := strings.CutPrefix(srv.URL, "http://")
	if gotHost != srvHost {
		t.Fatalf("request host mismatch: got %q, want %q", gotHost, srvHost)
	}
}

func TestAuthEditor(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	baseOpt, err := baseURLRewriter(srv.URL)
	if err != nil {
		t.Fatalf("baseURLRewriter: %v", err)
	}

	client, err := sdk.New("", sdk.WithHTTPClient(newHTTPClient()), baseOpt, authEditor("my-token"))
	if err != nil {
		t.Fatalf("sdk.New: %v", err)
	}

	_, _ = client.APIKeys.ListApiKeys(context.Background(), "org", nil)

	wantAuth := "Bearer my-token"
	if gotAuth != wantAuth {
		t.Fatalf("Authorization header mismatch: got %q, want %q", gotAuth, wantAuth)
	}
}

func TestUploadFileURLNotRewritten(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "upload.bin")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client, err := sdk.New("test-key", sdk.WithHTTPClient(newHTTPClient()))
	if err != nil {
		t.Fatalf("sdk.New: %v", err)
	}

	uploadURL := srv.URL + "/upload/123"
	if err := client.UploadFile(context.Background(), uploadURL, path); err != nil {
		t.Fatalf("UploadFile: %v", err)
	}

	wantPath := "/upload/123"
	if gotPath != wantPath {
		t.Fatalf("upload URL was rewritten: got %q, want %q", gotPath, wantPath)
	}
}
