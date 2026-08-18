package weatherapps

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func noContentHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func TestHTTPTestServerReturnsControlledStatus(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(noContentHandler))
	defer server.Close()

	resp, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestOpenWeatherMapReturnsWhenContextIsCanceled(t *testing.T) {

	requestStarted := make(chan struct{})
	errorChan := make(chan error, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(requestStarted)
		<-r.Context().Done()
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	provider := OpenWeatherMap{
		APIKey:  "test-key",
		BaseURL: server.URL,
	}

	go func() {
		_, err := provider.Temperature(ctx, "tokyo")
		errorChan <- err
	}()

	select {
	case <-requestStarted:
		// Chegou ao servidor.
	case <-time.After(time.Second):
		t.Fatal("request did not reach the test server")
	}

	cancel()

	select {
	case err := <-errorChan:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want %v", err, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Fatal("OpenWeatherMap did not return after context cancellation")
	}
}
