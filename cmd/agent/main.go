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
		for {
			metricsController.UpdateMetrics()
			time.Sleep(cfg.PollInterval)
			log.Printf("updated %s", cfg.PollInterval)

		}
	}()

	time.Sleep(cfg.SendInterval)

	go func() {
		for {
			log.Printf("sended %s", cfg.SendInterval)
			metricsController.JSONSendMetrics()
			time.Sleep(cfg.SendInterval)
			//metricsController.SendMetrics()
		}
	}()
}
