package main

import (
	"fmt"
	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/controllers"
	"github.com/zubans/metrics/internal/services"
	"log"
	"net/http"
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

	go func() {
		for {
			time.Sleep(cfg.SendInterval)
			log.Printf("sended %s", cfg.SendInterval)

			err := pingServer(fmt.Sprintf("http://%s", cfg.AddressServer))
			if err != nil {
				log.Printf("Ping failed: %v", err)
			} else {
				log.Printf("Ping successful: Server is up and running.")
			}
			time.Sleep(cfg.PollInterval)
			metricsController.JSONSendMetrics()
			//metricsController.SendMetrics()
		}
	}()
}

func pingServer(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error pinging server: %v", err)
		return fmt.Errorf("error pinging server: %v", err)
	}
	defer resp.Body.Close()

	log.Printf("Server responded with status code: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("Server returned non-OK status: %s", resp.Status)
		return fmt.Errorf("server returned non-OK status: %s", resp.Status)
	}

	return nil
}
