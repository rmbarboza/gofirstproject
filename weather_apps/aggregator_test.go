package weatherapps

import (
	"errors"
	"testing"
)

type fakeWeatherProvider struct {
	temperature float64
	err         error
}

func (f fakeWeatherProvider) Temperature(_ string) (float64, error) {
	return f.temperature, f.err
}

func TestWeatherKelvinSuccess(t *testing.T) {
	city := "tokyo"
	want := 305.0

	mw := MultiWeatherProvider{
		fakeWeatherProvider{temperature: 300},
		fakeWeatherProvider{temperature: 310},
	}

	temp, err := mw.Temperature(city)

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

	mw := MultiWeatherProvider{
		fakeWeatherProvider{temperature: 300},
		fakeWeatherProvider{err: expectedErr},
	}

	temp, err := mw.Temperature(city)

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

	mw := MultiWeatherProvider{}

	_, err := mw.Temperature(city)

	if err == nil {
		t.Fatal("Temperature() returned nil error, want an error")
	}
}
