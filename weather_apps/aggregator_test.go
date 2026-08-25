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
	type testCase struct {
		name      string
		providers MultiWeatherProvider
		want      float64
	}

	city := "tokyo"

	ctx := context.Background()

	tests := []testCase{
		{
			name: "one provider",
			providers: MultiWeatherProvider{
				fakeWeatherProvider{temperature: 300},
			},
			want: 300,
		},
		{
			name: "two providers",
			providers: MultiWeatherProvider{
				fakeWeatherProvider{temperature: 300},
				fakeWeatherProvider{temperature: 310},
			},
			want: 305,
		},
		{
			name: "three providers",
			providers: MultiWeatherProvider{
				fakeWeatherProvider{temperature: 270},
				fakeWeatherProvider{temperature: 280},
				fakeWeatherProvider{temperature: 290},
			},
			want: 280,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mw := tc.providers

			temp, err := mw.Temperature(ctx, city)

			if err != nil {
				t.Fatalf("Temperature() returned error: %v", err)
			}

			if temp != tc.want {
				t.Errorf("Temperature() = %.2f, want %.2f", temp, tc.want)
			}
		})
	}
}

func TestWeatherKelvinReturnsErrorWhenProviderFails(t *testing.T) {
	type testCase struct {
		name      string
		providers MultiWeatherProvider
	}

	city := "tokyo"
	want := 0.0
	expectedErr := errors.New("provider unavailable")

	ctx := context.Background()

	tests := []testCase{
		{
			name: "only provider fails",
			providers: MultiWeatherProvider{
				fakeWeatherProvider{err: expectedErr},
			},
		},
		{
			name: "failing provider listed first",
			providers: MultiWeatherProvider{
				fakeWeatherProvider{err: expectedErr},
				fakeWeatherProvider{temperature: 300},
			},
		},
		{
			name: "failing provider listed last",
			providers: MultiWeatherProvider{
				fakeWeatherProvider{temperature: 300},
				fakeWeatherProvider{err: expectedErr},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mw := tc.providers

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
		})
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
