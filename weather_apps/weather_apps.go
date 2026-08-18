package weatherapps

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

type WeatherProvider interface {
	Temperature(ctx context.Context, city string) (float64, error) // in Kelvin, naturally
}

type OpenWeatherMap struct {
	APIKey  string
	BaseURL string
}

type WeatherUnderground struct {
	APIKey  string
	BaseURL string
}

const (
	openWeatherMapBaseURL     = "http://api.openweathermap.org"
	weatherUndergroundBaseURL = "http://api.wunderground.com"
)

// Method for open weather map
func (w OpenWeatherMap) Temperature(ctx context.Context, city string) (float64, error) {
	baseURL := w.BaseURL
	if baseURL == "" {
		baseURL = openWeatherMapBaseURL
	}

	url := baseURL + "/data/2.5/weather?APPID=" + w.APIKey + "&q=" + city

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var d struct {
		Main struct {
			Kelvin float64 `json:"temp"`
		} `json:"main"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}

	log.Printf("openWeatherMap: %s: %.2f", city, d.Main.Kelvin)
	return d.Main.Kelvin, nil
}

// Method for weather underground
func (w WeatherUnderground) Temperature(ctx context.Context, city string) (float64, error) {
	baseURL := w.BaseURL
	if baseURL == "" {
		baseURL = weatherUndergroundBaseURL
	}

	url := baseURL + "/api/" + w.APIKey + "/conditions/q/" + city + ".json"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	var d struct {
		Observation struct {
			Celsius float64 `json:"temp_c"`
		} `json:"current_observation"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
		return 0, err
	}

	kelvin := d.Observation.Celsius + 273.15
	log.Printf("weatherUnderground: %s: %.2f", city, kelvin)
	return kelvin, nil
}
