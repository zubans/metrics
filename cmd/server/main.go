package main

import (
	"github.com/zubans/metrics/cmd/server/internal/handler"
	"github.com/zubans/metrics/cmd/server/internal/storage"
	"log"
	"net/http"
)

func main() {
	var memStorage = storage.NewMemStorage()
	memHandler := handler.NewHandler(memStorage)

	log.Println("Starting server")

	r := getRouter(memHandler)
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
