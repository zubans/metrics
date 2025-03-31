package main

import (
	"flag"
	"github.com/zubans/metrics/cmd/agent/internal/config"
	"github.com/zubans/metrics/cmd/agent/internal/controllers"
	"github.com/zubans/metrics/cmd/agent/internal/services"
	"log"
	"time"
)

func main() {
	metricsService := services.NewMetricsService()

	metricsController := controllers.NewMetricsController(metricsService)

	var repInt int
	var pollInt int
	var addr string

	flag.StringVar(&addr, "a", "localhost:8080", "address and port to run server")

	flag.IntVar(&repInt, "r", 10, "report send interval")
	flag.IntVar(&pollInt, "p", 2, "poll interval")
	flag.Parse()

	defer log.Println("stopped")

	cfg := config.Config{
		AddressServer: addr,
		SendInterval:  time.Duration(repInt) * time.Second,
		PollInterval:  time.Duration(pollInt) * time.Second,
	}

	go func() {
		for {
			metricsController.UpdateMetrics()

			time.Sleep(cfg.PollInterval)
		}
	}()

	go func() {
		for {
			metricsController.SendMetrics()

			time.Sleep(cfg.SendInterval)
		}
	}()

	select {}
}
