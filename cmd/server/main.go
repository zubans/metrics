package main

import (
	"context"
	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/handler"
	"github.com/zubans/metrics/internal/logger"
	"github.com/zubans/metrics/internal/middlewares"
	"github.com/zubans/metrics/internal/router"
	"github.com/zubans/metrics/internal/services"
	"github.com/zubans/metrics/internal/storage"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

var cfg = config.NewServerConfig()

func run(h http.Handler) error {

	if err := logger.Initialize(cfg.FlagLogLevel); err != nil {
		log.Printf("logger error: %v", err)
	}

	logger.Log.Info("Starting server on ", zap.String("address", cfg.RunAddr))

	return http.ListenAndServe(cfg.RunAddr, h)
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			logger.Log.Info("CRITICAL panic occurred", zap.Any("error", r))
			log.Printf("CRITICAL error %v", r)

		}
	}()

	var memStorage = storage.NewMemStorage()
	var dump = storage.New(memStorage, *cfg)

	var actualStorage services.MetricStorage

	if cfg.DBCfg != "" {
		err := storage.InitDB(cfg.DBCfg, "./migrations")
		if err != nil {
			logger.Log.Info("error init DB", zap.Any("error", err))
		}
		actualStorage = storage.NewDB(storage.DB)
	} else {
		if cfg.StoreInterval == 0 {
			actualStorage = storage.NewAutoDump(memStorage, dump)
		} else {
			actualStorage = memStorage
			go func() {
				ticker := time.NewTicker(cfg.StoreInterval)
				defer ticker.Stop()

				for range ticker.C {
					if err := dump.SaveMetricToFile(context.Background()); err != nil {
						logger.Log.Info("error save to file", zap.Any("error", err))
					} else {
						logger.Log.Info("metrics saved to file successfully")
					}
				}
			}()
		}
	}

	if cfg.Restore {
		err := dump.LoadMetricsFromFile()
		if err != nil {
			logger.Log.Info("error load from file", zap.Any("error", err))
		}
	}

	var serv = services.NewMetricService(actualStorage)
	var memHandler = handler.NewHandler(serv)
	r := router.GetRouter(memHandler)

	if err := run(middlewares.RequestLogger(r)); err != nil {
		log.Printf("Server failed to start: %v", err)
	}

	defer func() {
		logger.Log.Info("Saving metrics before shutdown...")
		if err := dump.SaveMetricToFile(context.Background()); err != nil {
			logger.Log.Info("failed to save metrics: ", zap.Any("error", err))
		} else {
			logger.Log.Info("Metrics saved.")
		}
	}()
}
