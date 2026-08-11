package weatherapps

import "testing"

type fakeWeatherProvider struct {
	temperature float64
	err         error
}

func (f fakeWeatherProvider) Temperature(_ string) (float64, error) {
	return f.temperature, f.err
}

func TestWeatherKelvin(t *testing.T) {
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
