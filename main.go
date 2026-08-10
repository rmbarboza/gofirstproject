package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

import weatherapps "github.com/rmbarboza/gofirstproject/weather_apps"

type multiWeatherProvider []weatherapps.WeatherProvider

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello!"))
}

func (w multiWeatherProvider) temperature(city string) (float64, error) {
	// Make a channel for temperatures, and a channel for errors.
	// Each provider will push a value into only one.
	temps := make(chan float64, len(w))
	errs := make(chan error, len(w))

	// For each provider, spawn a goroutine with an anonymous function.
	// That function will invoke the temperature method, and forward the response.
	for _, provider := range w {
		go func(p weatherapps.WeatherProvider) {
			k, err := p.Temperature(city)
			if err != nil {
				errs <- err
				return
			}
			temps <- k
		}(provider)
	}

	sum := 0.0

	// Collect a temperature or an error from each provider.
	for i := 0; i < len(w); i++ {
		select {
		case temp := <-temps:
			sum += temp
		case err := <-errs:
			return 0, err
		}
	}

	// Return the average, same as before.
	return sum / float64(len(w)), nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from system env")
	}

	// Read the key
	openWeatherMapApiKey := os.Getenv("OPEN_WEATHER_MAP_APPID")
	if openWeatherMapApiKey == "" {
		log.Fatal("OPEN_WEATHER_MAP_APPID is not set")
	}

	mw := multiWeatherProvider{
		weatherapps.OpenWeatherMap{APIKey: openWeatherMapApiKey},
		weatherapps.WeatherUnderground{APIKey: "your-key-here"},
	}

	http.HandleFunc("/hello", hello)

	http.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		city := strings.SplitN(r.URL.Path, "/", 3)[2]

		temp, err := mw.temperature(city)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"city": city,
			"temp": temp,
			"took": time.Since(begin).String(),
		})
	})

	http.ListenAndServe(":8080", nil)
}
