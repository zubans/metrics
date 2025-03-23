package main

import (
	"github.com/zubans/metrics/internal/handler"
	"github.com/zubans/metrics/internal/storage"
	"log"
	"net/http"
)

func main() {
	memStorage := storage.NewMemStorage()
	memHandler := handler.NewHandler(memStorage)

	log.Println("Starting server")

	err := http.ListenAndServe(":8080", memHandler.Router())
	if err != nil {
		log.Fatal(err)
	}
}
