package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

import weatherapps "github.com/rmbarboza/gofirstproject/weather_apps"

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello!"))
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok\n"))
}

func newMux(mw weatherapps.MultiWeatherProvider) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/hello", hello)

	mux.HandleFunc("/health", health)

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

	return mux
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	enverr := godotenv.Load()
	if enverr != nil {
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

	mux := newMux(mw)

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server returned:", err)
		}
	case <-signalCtx.Done():
		log.Println("shutdown requested")
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		begin := time.Now()
		err := server.Shutdown(shutdownCtx)
		log.Printf("Shutdown levou %v; erro: %v\n", time.Since(begin), err)
	}
}
