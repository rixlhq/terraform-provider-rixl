package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/rixlhq/rixl-go/sdk"
)

func newHTTPClient(ver ...string) *http.Client {
	var baseTransport http.RoundTripper
	if dt, ok := http.DefaultTransport.(*http.Transport); ok {
		cloned := dt.Clone()
		cloned.DialContext = (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext
		baseTransport = cloned
	} else {
		baseTransport = http.DefaultTransport
	}

	ua := userAgent
	if len(ver) > 0 && ver[0] != "" {
		ua = fmt.Sprintf("%s/%s", userAgent, ver[0])
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: &userAgentTransport{inner: baseTransport, ua: ua},
	}
}

type userAgentTransport struct {
	inner http.RoundTripper
	ua    string
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent", t.ua)
	return t.inner.RoundTrip(req)
}

func baseURLRewriter(base string) (sdk.Option, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("parse base_url: %w", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, errors.New("base_url must include scheme and host")
	}

	return sdk.WithRequestEditor(func(_ context.Context, req *http.Request) error {
		req.URL.Scheme = baseURL.Scheme
		req.URL.Host = baseURL.Host
		if baseURL.Path != "" {
			prefix, _ := strings.CutSuffix(baseURL.Path, "/")
			req.URL.Path = prefix + req.URL.Path
			if req.URL.RawPath != "" {
				req.URL.RawPath = prefix + req.URL.RawPath
			}
		}
		if req.Host != "" {
			req.Host = baseURL.Host
		}
		return nil
	}), nil
}

func authEditor(token string) sdk.Option {
	return sdk.WithRequestEditor(func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	})
}
