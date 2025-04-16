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

	go func() {
		for {
			time.Sleep(cfg.SendInterval)
			metricsController.JSONSendMetrics()
			//metricsController.SendMetrics()

		}
	}()

	go func() {
		for {
			metricsController.UpdateMetrics()
			time.Sleep(cfg.PollInterval)
		}
	}()

	select {}
}
