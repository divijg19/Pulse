package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func TestHealth(t *testing.T) {
	handler := NewHandler(testStaticFS())
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != "OK" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestRunMethodNotAllowed(t *testing.T) {
	handler := NewHandler(testStaticFS())
	req := httptest.NewRequest(http.MethodGet, "/run", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d (expected 405)", rec.Code)
	}
}

func TestRunRejectsInvalidPayload(t *testing.T) {
	handler := NewHandler(testStaticFS())
	req := httptest.NewRequest(http.MethodPost, "/run", strings.NewReader(`{"url":"","method":"GET","concurrency":1}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestRunAcceptsValidPayload(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Pulse-Test", "ok")
		_, _ = io.WriteString(w, "pong")
	}))
	defer upstream.Close()

	handler := NewHandler(testStaticFS())
	body := `{"url":"` + upstream.URL + `","method":"GET","headers":{},"body":"","concurrency":1}`
	req := httptest.NewRequest(http.MethodPost, "/run", strings.NewReader(body))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %q", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), upstream.URL) {
		t.Fatalf("response did not include accepted URL: %q", rec.Body.String())
	}
}

func TestPanicRecovery(t *testing.T) {
	panicking := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})
	handler := recoveryMiddleware(panicking)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d (expected 500)", rec.Code)
	}
}

func testStaticFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html": {Data: []byte("<html></html>")},
	}
}
