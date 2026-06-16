package engine

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/stream"
)

func TestExecuteSingle_Happy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	result := ExecuteSingle(context.Background(), srv.URL, "GET", nil, "")
	if result.Status != http.StatusOK {
		t.Fatalf("status = %d", result.Status)
	}
	if result.Error != "" {
		t.Fatalf("error = %q", result.Error)
	}
	if result.RequestURL != srv.URL {
		t.Fatalf("RequestURL = %q", result.RequestURL)
	}
	if result.RequestMethod != "GET" {
		t.Fatalf("RequestMethod = %q", result.RequestMethod)
	}
}

func TestExecuteSingle_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result := ExecuteSingle(ctx, srv.URL, "GET", nil, "")
	if result.Status != 0 {
		t.Fatalf("status = %d", result.Status)
	}
	if result.Error == "" {
		t.Fatal("expected error on timeout")
	}
}

func TestExecuteSingle_Error(t *testing.T) {
	result := ExecuteSingle(context.Background(), "http://127.0.0.1:1", "GET", nil, "")
	if result.Status != 0 {
		t.Fatalf("status = %d", result.Status)
	}
	if result.Error == "" {
		t.Fatal("expected error for invalid address")
	}
}

func TestExecuteSingle_WithHeadersAndBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Error("expected POST method")
		}
		if r.Header.Get("X-Custom") != "test-value" {
			t.Error("expected X-Custom header")
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"key":"value"}` {
			t.Errorf("body = %q", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	headers := map[string]string{"X-Custom": "test-value"}
	result := ExecuteSingle(context.Background(), srv.URL, "POST", headers, `{"key":"value"}`)
	if result.Status != http.StatusOK {
		t.Fatalf("status = %d", result.Status)
	}
	if result.Error != "" {
		t.Fatalf("error = %q", result.Error)
	}
}

func TestExecuteSingle_RequestMethodURL_OnTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	result := ExecuteSingle(ctx, srv.URL, "POST", nil, "")
	if result.RequestMethod != "POST" {
		t.Fatalf("RequestMethod = %q", result.RequestMethod)
	}
	if result.RequestURL != srv.URL {
		t.Fatalf("RequestURL = %q", result.RequestURL)
	}
}

func TestExecuteSingle_RequestMethodURL_OnConnectError(t *testing.T) {
	result := ExecuteSingle(context.Background(), "http://127.0.0.1:1", "DELETE", nil, "")
	if result.RequestMethod != "DELETE" {
		t.Fatalf("RequestMethod = %q", result.RequestMethod)
	}
	if result.RequestURL != "http://127.0.0.1:1" {
		t.Fatalf("RequestURL = %q", result.RequestURL)
	}
}

func TestExecuteConcurrent_CancelMidExecution(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	hub := stream.NewHub()
	eventCh := make(chan model.Event, 10)
	hub.Add(eventCh)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		ExecuteConcurrent(ctx, model.RunRequest{
			URL: srv.URL, Method: "GET", Concurrency: 10,
		}, hub)
		close(done)
	}()

	// Let at least one result through
	select {
	case <-eventCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for first result")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecuteConcurrent did not return after cancellation")
	}
}

func TestExecuteConcurrent_Cancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	hub := stream.NewHub()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	go func() {
		ExecuteConcurrent(ctx, model.RunRequest{
			URL: srv.URL, Method: "GET", Concurrency: 10,
		}, hub)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ExecuteConcurrent did not return after cancellation")
	}
}
