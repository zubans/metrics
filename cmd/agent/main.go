package main

import (
	"flag"
	"github.com/zubans/metrics/cmd/agent/internal/controllers"
	"github.com/zubans/metrics/cmd/agent/internal/services"
	"log"
	"time"
)

var a string

func main() {
	metricsService := services.NewMetricsService()

	metricsController := controllers.NewMetricsController(metricsService)

	var repInt int
	var pollInt int

	flag.StringVar(&a, "a", "localhost:8080", "address and port to run server")

	flag.IntVar(&repInt, "r", 10, "report send interval")
	flag.IntVar(&pollInt, "p", 2, "poll interval")
	flag.Parse()

	defer log.Println("stopped")

	go func() {
		for {
			metricsController.UpdateMetrics()

			time.Sleep(time.Duration(pollInt) * time.Second)
		}
	}()

	go func() {
		for {
			metricsController.SendMetrics()

			time.Sleep(time.Duration(repInt) * time.Second)
		}
	}()

	select {}
}
