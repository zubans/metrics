package main

import (
	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/controllers"
	"github.com/zubans/metrics/internal/services"
	"log"
	"time"
)

func main() {
	var cfg = config.NewAgentConfig()

	metricsService := services.NewMetricsService(cfg)

	defer log.Println("stopped")

	log.Printf("Agent send to server address %s", cfg.AddressServer)
	log.Printf("Send interval: %v, Poll interval: %v", cfg.SendInterval, cfg.PollInterval)

	metricsController := controllers.NewMetricsController(metricsService)

	stopChan := make(chan struct{})

	run(metricsController, cfg)

	<-stopChan
}

func run(metricsController *controllers.MetricsController, cfg *config.AgentConfig) {
	go func() {
		ticker := time.NewTicker(cfg.PollInterval)
		for range ticker.C {
			metricsController.UpdateMetrics()
		}
	}()

	jobs := make(chan struct{})

	for i := 0; i < cfg.RateLimit; i++ {
		go func() {
			for range jobs {
				metricsController.JSONSendMetrics()
			}
		}()
	}

	go func() {
		ticker := time.NewTicker(cfg.SendInterval)
		for range ticker.C {
			jobs <- struct{}{}
		}
	}()
}
