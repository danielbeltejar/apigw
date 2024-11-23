package main

import (
	"net/http"
	"os"

	"apigw/internal/gateway"
	"apigw/internal/metrics"
	"apigw/pkg/utils"
)

func main() {
	logger := utils.NewLogger()
	logger.Info("Starting API Gateway...")

	// Initialize metrics
	metrics.InitMetrics()

	// Load gateway configuration
	gw := gateway.NewGateway(logger)
	configFile := os.Getenv("CONFIG_PATH")
	if configFile == "" {
		configFile = "config/config.yaml"
	}
	gw.LoadConfig(configFile)

	// Register routes
	for _, route := range gw.Routes {
		logger.Infof("Registered route: %s", route.Pattern)
	}

	// Health check
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Metrics endpoint
	http.Handle("/metrics", metrics.Handler())

	// Run health check server
	go func() {
		logger.Info("Health check server listening on port 8081...")
		if err := http.ListenAndServe(":8081", nil); err != nil {
			logger.Fatalf("Failed to start health check server: %v", err)
		}
	}()

	// Run API Gateway
	logger.Info("API Gateway listening on port 8080...")
	if err := http.ListenAndServe(":8080", gw); err != nil {
		logger.Fatalf("Failed to start API Gateway: %v", err)
	}
}
