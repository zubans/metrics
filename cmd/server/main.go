package main

import (
	"fmt"
	"github.com/zubans/metrics/internal/config"
	"github.com/zubans/metrics/internal/handler"
	"github.com/zubans/metrics/internal/router"
	"github.com/zubans/metrics/internal/services"
	"github.com/zubans/metrics/internal/storage"
	"log"
	"net/http"
)

func main() {
	cfg := config.NewServerConfig()

	var memStorage = storage.NewMemStorage()
	var serv = services.NewMetricService(memStorage)
	memHandler := handler.NewHandler(serv)

	r := router.GetRouter(memHandler)

	log.Println(fmt.Printf("Starting server on %s \n", cfg.RunAddr))
	err := http.ListenAndServe(cfg.RunAddr, r)
	if err != nil {
		log.Fatal(err)
	}
}
