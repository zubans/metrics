package main

import (
	"flag"
	"github.com/zubans/metrics/cmd/agent/internal/config"
	"github.com/zubans/metrics/cmd/agent/internal/controllers"
	"github.com/zubans/metrics/cmd/agent/internal/services"
	"log"
	"os"
	"strconv"
	"time"
)

func main() {
	metricsService := services.NewMetricsService()

	//metricsController := controllers.NewMetricsController(metricsService)

	var repInt int
	var pollInt int
	var addr string

	flag.StringVar(&addr, "a", "localhost:8080", "address and port to run server")

	flag.IntVar(&repInt, "r", 10, "report send interval")
	flag.IntVar(&pollInt, "p", 2, "poll interval")
	flag.Parse()

	defer log.Println("stopped")

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		addr = envRunAddr
	}

	if r := os.Getenv("REPORT_INTERVAL"); r != "" {
		repInt, _ = strconv.Atoi(r)
	}

	if p := os.Getenv("POLL_INTERVAL"); p != "" {
		pollInt, _ = strconv.Atoi(p)
	}

	cfg := config.Config{
		AddressServer: addr,
		SendInterval:  time.Duration(repInt) * time.Second,
		PollInterval:  time.Duration(pollInt) * time.Second,
	}

	log.Printf("Server started at %s", cfg.AddressServer)
	log.Printf("Send interval: %v, Poll interval: %v", cfg.SendInterval, cfg.PollInterval)

	metricsController := controllers.NewMetricsController(metricsService, cfg)

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
