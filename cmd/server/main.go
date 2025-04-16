package main

import (
	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/handler"
	"github.com/zubans/metrics/internal/logger"
	"github.com/zubans/metrics/internal/router"
	"github.com/zubans/metrics/internal/services"
	"github.com/zubans/metrics/internal/storage"
	"go.uber.org/zap"
	"log"
	"net/http"
)

var cfg = config.NewServerConfig()

func run(h http.Handler) error {
	if err := logger.Initialize(cfg.FlagLogLevel); err != nil {
		log.Fatalf("logger error: %v", err)
	}

	logger.Log.Info("Starting server on ", zap.String("address", cfg.RunAddr))

	return http.ListenAndServe(cfg.RunAddr, h)
}

func main() {
	log.Printf("here strat %v", "server")
	var memStorage = storage.NewMemStorage()
	var serv = services.NewMetricService(memStorage)
	memHandler := handler.NewHandler(serv)

	r := router.GetRouter(memHandler)
	if err := run(logger.RequestLogger(r)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
