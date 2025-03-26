package main

import (
	"github.com/zubans/metrics/cmd/agent/internal/controllers"
	"github.com/zubans/metrics/cmd/agent/internal/services"
	"time"
)

func main() {
	metricsService := services.NewMetricsService()

	metricsController := controllers.NewMetricsController(metricsService)

	go func() {
		for {
			metricsController.UpdateMetrics()

			time.Sleep(2 * time.Second)
		}
	}()

	go func() {
		for {
			metricsController.SendMetrics()

			time.Sleep(10 * time.Second)
		}
	}()

	select {}
}
