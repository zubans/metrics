package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/controllers"
	"github.com/zubans/metrics/internal/logger"
	"github.com/zubans/metrics/internal/services"
	"github.com/zubans/metrics/internal/version"
	"go.uber.org/zap"
)

func main() {
	version.PrintBuildInfo()

	var cfg = config.NewAgentConfig()

	// Initialize logger
	if err := logger.Initialize("info"); err != nil {
		log.Printf("logger initialization failed: %v", err)
	}

	metricsService := services.NewMetricsService(cfg)

	defer logger.Log.Info("Agent stopped")

	logger.Log.Info("Agent starting",
		zap.String("server_address", cfg.AddressServer),
		zap.Duration("send_interval", cfg.SendInterval),
		zap.Duration("poll_interval", cfg.PollInterval))

	metricsController := controllers.NewMetricsController(metricsService)
	defer metricsController.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	run(ctx, &wg, metricsController, cfg)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigCh

	cancel()
	wg.Wait()

	metricsController.SendMetrics()
}

func run(ctx context.Context, wg *sync.WaitGroup, metricsController *controllers.MetricsController, cfg *config.AgentConfig) {
	wg.Add(2)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(cfg.PollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metricsController.UpdateMetrics()
			}
		}
	}()

	go func() {
		defer wg.Done()
		ticker := time.NewTicker(cfg.SendInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metricsController.SendMetrics()
			}
		}
	}()
}
