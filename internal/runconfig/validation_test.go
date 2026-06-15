package runconfig

import (
	"testing"

	"github.com/divijg19/Pulse/internal/model"
)

func TestValidateNormalizesRequest(t *testing.T) {
	req, err := Validate(model.RunRequest{
		URL:         " https://example.com/path ",
		Method:      "post",
		Concurrency: 3,
	})
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if req.URL != "https://example.com/path" {
		t.Fatalf("URL was not trimmed: %q", req.URL)
	}
	if req.Method != "POST" {
		t.Fatalf("method was not normalized: %q", req.Method)
	}
	if req.Headers == nil {
		t.Fatal("headers should be initialized")
	}
}

func TestValidateRejectsInvalidRequests(t *testing.T) {
	tests := []model.RunRequest{
		{Method: "GET", Concurrency: 1},
		{URL: "/relative", Method: "GET", Concurrency: 1},
		{URL: "ftp://example.com", Method: "GET", Concurrency: 1},
		{URL: "https://example.com", Method: "TRACE", Concurrency: 1},
		{URL: "https://example.com", Method: "GET", Concurrency: 0},
		{URL: "https://example.com", Method: "GET", Concurrency: 101},
	}

	for _, test := range tests {
		if _, err := Validate(test); err == nil {
			t.Fatalf("Validate(%+v) returned nil error", test)
		}
	}
}

func TestClampConcurrency(t *testing.T) {
	if got := ClampConcurrency(-1); got != MinConcurrency {
		t.Fatalf("low value clamp = %d", got)
	}
	if got := ClampConcurrency(500); got != MaxConcurrency {
		t.Fatalf("high value clamp = %d", got)
	}
	if got := ClampConcurrency(42); got != 42 {
		t.Fatalf("in-range clamp = %d", got)
	}
}
