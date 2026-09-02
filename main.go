package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

import weatherapps "github.com/rmbarboza/gofirstproject/weather_apps"

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello!"))
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

	mw := weatherapps.MultiWeatherProvider{
		weatherapps.OpenWeatherMap{APIKey: openWeatherMapApiKey},
		weatherapps.WeatherUnderground{APIKey: "your-key-here"},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/hello", hello)

	mux.HandleFunc("/weather/", func(w http.ResponseWriter, r *http.Request) {
		begin := time.Now()
		city := strings.SplitN(r.URL.Path, "/", 3)[2]

		temp, err := mw.Temperature(r.Context(), city)
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

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	err = server.ListenAndServe()

	if !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
