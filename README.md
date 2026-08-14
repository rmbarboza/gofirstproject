# gofirstproject
First little project in Go

This project gets temperature from multiple weather data sources, calculates the average in Kelvin and exports data in json format.

## Current aggregation policy

1. Every provider is queried concurrently.
2. Temperature average is returned only when all providers respond successfully.
3. Any provider error invalidates all successful results.
4. An empty provider list returns an error.
5. Timeout and cancellation are not implemented yet.

## Running locally

* Requirements: Go 1.26.5 .
* Configure .env file with the API KEY used for each provider. This name is used in the provider implementation and will vary for each provider.
* Command to start: go run .
* URL: http://localhost:8080/weather/{city}

## Validation

```
go test ./...
go vet ./...
```

