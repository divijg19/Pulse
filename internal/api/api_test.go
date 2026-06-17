package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/divijg19/Pulse/internal/model"
	"github.com/divijg19/Pulse/internal/stream"
)

type flushRecorder struct {
	*httptest.ResponseRecorder
}

func (f *flushRecorder) Flush() {}

func TestHandleRun_MethodNotAllowed(t *testing.T) {
	hub := stream.NewHub()
	handler := &RunHandler{Hub: hub}

	req := httptest.NewRequest(http.MethodGet, "/run", nil)
	rec := httptest.NewRecorder()
	handler.HandleRun(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d (expected %d)", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandleRun_InvalidBody(t *testing.T) {
	hub := stream.NewHub()
	handler := &RunHandler{Hub: hub}

	t.Run("malformed json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/run", strings.NewReader(`{bad`))
		rec := httptest.NewRecorder()
		handler.HandleRun(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d (expected %d)", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("unknown fields", func(t *testing.T) {
		body := `{"url":"https://example.com","method":"GET","concurrency":1,"unknown":"field"}`
		req := httptest.NewRequest(http.MethodPost, "/run", strings.NewReader(body))
		rec := httptest.NewRecorder()
		handler.HandleRun(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d (expected %d)", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		body := `{"url":"","method":"GET","concurrency":1}`
		req := httptest.NewRequest(http.MethodPost, "/run", strings.NewReader(body))
		rec := httptest.NewRecorder()
		handler.HandleRun(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d (expected %d)", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestHandleRun_OversizedBody(t *testing.T) {
	hub := stream.NewHub()
	handler := &RunHandler{Hub: hub}

	padding := maxRunRequestBodyBytes + 1 - len(`{"k":""}`)
	body := strings.NewReader(`{"k":"` + strings.Repeat("v", padding) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/run", body)
	rec := httptest.NewRecorder()
	handler.HandleRun(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d (expected %d): body = %q", rec.Code, http.StatusRequestEntityTooLarge, rec.Body.String())
	}
}

func TestHandleRun_Valid(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	hub := stream.NewHub()
	handler := &RunHandler{Hub: hub}

	body := `{"url":"` + upstream.URL + `","method":"GET","headers":{},"body":"","concurrency":1}`
	req := httptest.NewRequest(http.MethodPost, "/run", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler.HandleRun(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d (expected %d): body = %q", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), upstream.URL) {
		t.Fatalf("response does not include URL: %q", rec.Body.String())
	}
}

func TestStreamHandler_Headers(t *testing.T) {
	hub := stream.NewHub()
	handler := &StreamHandler{Hub: hub}

	rec := &flushRecorder{httptest.NewRecorder()}
	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()
	handler.ServeHTTP(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("Content-Type = %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-cache" {
		t.Fatalf("Cache-Control = %q", cc)
	}
}

func TestStreamHandler_Integration(t *testing.T) {
	hub := stream.NewHub()
	handler := &StreamHandler{Hub: hub}

	rec := &flushRecorder{httptest.NewRecorder()}
	req := httptest.NewRequest(http.MethodGet, "/stream", nil)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(rec, req)
		close(done)
	}()

	time.Sleep(5 * time.Millisecond)

	hub.Broadcast(model.Event{Type: "result", Data: model.Result{Status: 200, Latency: 100 * time.Millisecond}})

	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("handler did not exit after cancellation")
	}

	body := strings.TrimRight(rec.Body.String(), "\n")
	lines := strings.Split(body, "\n")

	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d: %q", len(lines), body)
	}
	if !strings.HasPrefix(lines[0], "event: result") {
		t.Fatalf("line 0 = %q", lines[0])
	}
	if !strings.Contains(lines[1], `"status":200`) {
		t.Fatalf("line 1 = %q", lines[1])
	}
	if !strings.Contains(lines[1], `"latencyNs"`) {
		t.Fatalf("line 1 missing latencyNs: %q", lines[1])
	}
}
