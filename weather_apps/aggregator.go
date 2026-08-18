package weatherapps

import (
	"context"
	"errors"
)

type MultiWeatherProvider []WeatherProvider

func (w MultiWeatherProvider) Temperature(ctx context.Context, city string) (float64, error) {
	if len(w) == 0 {
		return 0, errors.New("empty provider list")
	}

	type weatherResult struct {
		temperature float64
		err         error
	}

	// Make a channel for weatherResult.
	// Each provider will push only one weatherResult.
	results := make(chan weatherResult, len(w))

	// For each provider, spawn a goroutine with an anonymous function.
	// That function will invoke the temperature method, and forward the response.
	for _, provider := range w {
		go func(p WeatherProvider) {
			temp, err := p.Temperature(ctx, city)
			results <- weatherResult{
				temperature: temp,
				err:         err,
			}
		}(provider)
	}

	sum := 0.0

	// Collect a temperature or an error from each provider.
	for i := 0; i < len(w); i++ {
		select {
		case res := <-results:
			if res.err != nil {
				return 0, res.err
			}
			sum += res.temperature
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	}

	// Return the average, same as before.
	return sum / float64(len(w)), nil
}
