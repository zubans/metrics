package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/controllers"
	"github.com/zubans/metrics/internal/services"
	"github.com/zubans/metrics/internal/version"
)

func main() {
	version.PrintBuildInfo()

	var cfg = config.NewAgentConfig()

	metricsService := services.NewMetricsService(cfg)

	defer log.Println("stopped")

	log.Printf("Agent send to server address %s", cfg.AddressServer)
	log.Printf("Send interval: %v, Poll interval: %v", cfg.SendInterval, cfg.PollInterval)

	metricsController := controllers.NewMetricsController(metricsService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	run(ctx, metricsController, cfg)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sigCh

	cancel()
	time.Sleep(100 * time.Millisecond)
	metricsController.OldJSONSendMetrics()
	metricsController.JSONSendMetrics()
}

func run(ctx context.Context, metricsController *controllers.MetricsController, cfg *config.AgentConfig) {
	go func() {
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
		ticker := time.NewTicker(cfg.SendInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				metricsController.OldJSONSendMetrics()
				metricsController.JSONSendMetrics()
			}
		}
	}()
}
