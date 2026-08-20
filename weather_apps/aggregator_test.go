package weatherapps

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeWeatherProvider struct {
	temperature float64
	err         error
}

type blockingWeatherProvider struct {
	release          chan struct{}
	providerReturned chan struct{}
}

func (f fakeWeatherProvider) Temperature(_ context.Context, _ string) (float64, error) {
	return f.temperature, f.err
}

func (f blockingWeatherProvider) Temperature(_ context.Context, _ string) (float64, error) {
	defer close(f.providerReturned)

	<-f.release

	return 0, nil
}

func TestWeatherKelvinSuccess(t *testing.T) {
	city := "tokyo"
	want := 305.0

	ctx := context.Background()

	mw := MultiWeatherProvider{
		fakeWeatherProvider{temperature: 300},
		fakeWeatherProvider{temperature: 310},
	}

	temp, err := mw.Temperature(ctx, city)

	if err != nil {
		t.Fatalf("Temperature() returned error: %v", err)
	}

	if temp != want {
		t.Errorf("Temperature() = %.2f, want %.2f", temp, want)
	}
}

func TestWeatherKelvinReturnsErrorWhenProviderFails(t *testing.T) {
	city := "tokyo"
	want := 0.0
	expectedErr := errors.New("provider unavailable")

	ctx := context.Background()

	mw := MultiWeatherProvider{
		fakeWeatherProvider{temperature: 300},
		fakeWeatherProvider{err: expectedErr},
	}

	temp, err := mw.Temperature(ctx, city)

	if err == nil {
		t.Fatal("Temperature() returned nil error, want an error")
	}

	if temp != want {
		t.Errorf("Temperature() = %.2f, want %.2f", temp, want)
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Temperature() error = %v, want %v", err, expectedErr)
	}
}

func TestWeatherKelvinNoProvidersReturnsError(t *testing.T) {
	city := "tokyo"

	ctx := context.Background()

	mw := MultiWeatherProvider{}

	_, err := mw.Temperature(ctx, city)

	if err == nil {
		t.Fatal("Temperature() returned nil error, want an error")
	}
}

func TestTemperatureReturnsWhenContextIsCanceled(t *testing.T) {
	release := make(chan struct{})
	providerReturned := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mw := MultiWeatherProvider{
		blockingWeatherProvider{
			release:          release,
			providerReturned: providerReturned,
		},
	}

	result := make(chan error, 1)

	go func() {
		_, err := mw.Temperature(ctx, "tokyo")
		result <- err
	}()

	cancel()

	select {
	case err := <-result:
		// O agregador retornou.
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v, want %v", err, context.Canceled)
		}

	case <-time.After(time.Second):
		// O agregador não retornou em tempo razoável.
		t.Fatal("Temperature() did not return after context cancellation")
	}

	close(release)

	select {
	case <-providerReturned:
		// Provider returned
	case <-time.After(time.Second):
		// The provider didn't return in time limit
		t.Fatal("provider did not return after context cancellation")
	}
}

func TestTemperatureReturnsWhenContextDeadlineIsExceeded(t *testing.T) {
	release := make(chan struct{})
	providerReturned := make(chan struct{})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	mw := MultiWeatherProvider{
		blockingWeatherProvider{
			release:          release,
			providerReturned: providerReturned,
		},
	}

	result := make(chan error, 1)

	go func() {
		_, err := mw.Temperature(ctx, "tokyo")
		result <- err
	}()

	select {
	case err := <-result:
		// Aggregator returned
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("error = %v, want %v", err, context.DeadlineExceeded)
		}

	case <-time.After(time.Second):
		// The aggregator didn't return in time limit
		t.Fatal("Temperature() did not return after context deadline")
	}

	close(release)

	select {
	case <-providerReturned:
		// Provider returned
	case <-time.After(time.Second):
		// The provider didn't return in time limit
		t.Fatal("provider did not return after context deadline")
	}
}
