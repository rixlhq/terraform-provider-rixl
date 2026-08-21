package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/rixlhq/rixl-go/sdk"
	"github.com/rixlhq/rixl-go/sdk/images"
	"github.com/rixlhq/rixl-go/sdk/videos"
)

func TestIsTransientWaitError(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, true},
		{"image 404", &images.ClientHttpError[struct{}]{StatusCode: http.StatusNotFound}, true},
		{"image 500", &images.ClientHttpError[struct{}]{StatusCode: http.StatusInternalServerError}, true},
		{"image 502", &images.ClientHttpError[struct{}]{StatusCode: http.StatusBadGateway}, true},
		{"image 400", &images.ClientHttpError[struct{}]{StatusCode: http.StatusBadRequest}, false},
		{"video 404", &videos.ClientHttpError[struct{}]{StatusCode: http.StatusNotFound}, true},
		{"video 503", &videos.ClientHttpError[struct{}]{StatusCode: http.StatusServiceUnavailable}, true},
		{"video 401", &videos.ClientHttpError[struct{}]{StatusCode: http.StatusUnauthorized}, false},
		{"network", errors.New("network error"), false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := isTransientWaitError(c.err)
			if got != c.want {
				t.Fatalf("isTransientWaitError(%v) = %v, want %v", c.err, got, c.want)
			}
		})
	}
}

type queuedTransport struct {
	responses []*http.Response
	index     int
}

func (q *queuedTransport) RoundTrip(_ *http.Request) (*http.Response, error) {
	if q.index >= len(q.responses) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader("{}")),
			Header:     make(http.Header),
		}, nil
	}
	resp := q.responses[q.index]
	q.index++
	return resp, nil
}

func newTestClient(responses []*http.Response) (*sdk.Client, error) {
	httpClient := &http.Client{
		Transport: &queuedTransport{responses: responses},
	}
	return sdk.New("", sdk.WithHTTPClient(httpClient))
}

func imageResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}
}

func videoResponse(status int, body string) *http.Response {
	return imageResponse(status, body)
}

func TestWaitImageRetriesNotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := newTestClient([]*http.Response{
		imageResponse(http.StatusNotFound, "{}"),                 //nolint:bodyclose
		imageResponse(http.StatusOK, `{"image":{"id":"img-1"}}`), //nolint:bodyclose
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	r := &imageResource{client: client}
	img, err := r.waitForImage(ctx, "img-1")
	if err != nil {
		t.Fatalf("waitForImage: %v", err)
	}
	if img == nil || img.ID == nil || *img.ID != "img-1" {
		t.Fatalf("unexpected image: %v", img)
	}
}

func TestWaitImageFailsFastOnClientError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := newTestClient([]*http.Response{
		imageResponse(http.StatusBadRequest, "{}"), //nolint:bodyclose
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	r := &imageResource{client: client}
	_, err = r.waitForImage(ctx, "img-1")
	if err == nil {
		t.Fatalf("expected error")
	}
	var httpErr *images.ClientHttpError[struct{}]
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 error, got %v", err)
	}
}

func TestWaitVideoRetriesNotFound(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := newTestClient([]*http.Response{
		videoResponse(http.StatusNotFound, "{}"),                                            //nolint:bodyclose
		videoResponse(http.StatusOK, `{"video":{"id":"vid-1","width":1920,"height":1080}}`), //nolint:bodyclose
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	r := &videoResource{client: client}
	vid, err := r.waitForVideo(ctx, "vid-1")
	if err != nil {
		t.Fatalf("waitForVideo: %v", err)
	}
	if vid == nil || vid.ID == nil || *vid.ID != "vid-1" {
		t.Fatalf("unexpected video: %v", vid)
	}
}

func TestWaitVideoFailsFastOnClientError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := newTestClient([]*http.Response{
		videoResponse(http.StatusForbidden, "{}"), //nolint:bodyclose
	})
	if err != nil {
		t.Fatalf("client: %v", err)
	}

	r := &videoResource{client: client}
	_, err = r.waitForVideo(ctx, "vid-1")
	if err == nil {
		t.Fatalf("expected error")
	}
	var httpErr *videos.ClientHttpError[struct{}]
	if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 error, got %v", err)
	}
}
