package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/zubans/metrics/cmd/server/internal/handler"
	"net/http"
)

func GetRouter(h *handler.Handler) http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.ShowMetrics)
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetric)
	r.Route("/value/{type}", func(r chi.Router) {
		r.Route("/{name}", func(r chi.Router) {
			r.Post("/{value}", h.UpdateMetric)
			r.Get("/", h.GetMetric)
		})
	})

	return r
}
