package main

import (
	"flag"
	"fmt"
	"github.com/zubans/metrics/cmd/server/internal/handler"
	"github.com/zubans/metrics/cmd/server/internal/storage"
	"log"
	"net/http"
)

var flagRunAddr string

func main() {
	var memStorage = storage.NewMemStorage()
	memHandler := handler.NewHandler(memStorage)

	r := getRouter(memHandler)

	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	log.Println(fmt.Printf("Starting server on %s", flagRunAddr))
	err := http.ListenAndServe(flagRunAddr, r)
	if err != nil {
		log.Fatal(err)
	}
}
