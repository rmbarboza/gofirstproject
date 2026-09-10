package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	wantCode := http.StatusOK
	wantBody := "ok\n"

	mux := newMux(nil)
	mux.ServeHTTP(rec, req)

	if rec.Code != wantCode {
		t.Fatalf("GET /health returned code: %v; want %v", rec.Code, wantCode)
	}

	if rec.Body.String() != wantBody {
		t.Fatalf("GET /health returned body %q; want %q", rec.Body.String(), wantBody)
	}
}
