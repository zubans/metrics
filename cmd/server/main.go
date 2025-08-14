package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/cryptoutil"
	"github.com/zubans/metrics/internal/handler"
	"github.com/zubans/metrics/internal/logger"
	"github.com/zubans/metrics/internal/middlewares"
	"github.com/zubans/metrics/internal/router"
	"github.com/zubans/metrics/internal/services"
	"github.com/zubans/metrics/internal/storage"
	"go.uber.org/zap"
)

var cfg = config.NewServerConfig()

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

	var baseRouter = router.GetRouter(memHandler)
	var r = baseRouter
	if cfg.CryptoKey != "" {
		priv, err := cryptoutil.LoadPrivateKey(cfg.CryptoKey)
		if err != nil {
			logger.Log.Info("failed to load private key", zap.Any("error", err))
		} else {
			decrypt := func(env *cryptoutil.Envelope) ([]byte, error) {
				return cryptoutil.DecryptHybrid(priv, env)
			}
			r = middlewares.DecryptRequestMiddleware(decrypt)(baseRouter)
		}
	}

	if err := logger.Initialize(cfg.FlagLogLevel); err != nil {
		log.Printf("logger error: %v", err)
	}
	srv := &http.Server{Addr: cfg.RunAddr, Handler: middlewares.RequestLogger(r)}

	go func() {
		logger.Log.Info("Starting server on ", zap.String("address", cfg.RunAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server failed to start: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	logger.Log.Info("Saving metrics before shutdown...")
	if err := dump.SaveMetricToFile(context.Background()); err != nil {
		logger.Log.Info("failed to save metrics: ", zap.Any("error", err))
	} else {
		logger.Log.Info("Metrics saved.")
	}
}
