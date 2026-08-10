package weatherapps

import (
	"encoding/json"
	"log"
	"net/http"
)

type WeatherProvider interface {
	Temperature(city string) (float64, error) // in Kelvin, naturally
}

type OpenWeatherMap struct {
	APIKey string
}

type WeatherUnderground struct {
	APIKey string
}

// Method for open weather map
func (w OpenWeatherMap) Temperature(city string) (float64, error) {
	resp, err := http.Get("http://api.openweathermap.org/data/2.5/weather?APPID=" + w.APIKey + "&q=" + city)
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
func (w WeatherUnderground) Temperature(city string) (float64, error) {
	resp, err := http.Get("http://api.wunderground.com/api/" + w.APIKey + "/conditions/q/" + city + ".json")
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
