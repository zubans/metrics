package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/zubans/metrics/internal/handler"
	"net/http"
)

func GetRouter(h *handler.Handler) http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.ShowMetrics)
	r.Post("/value/", h.GetJSONMetric)
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetric)
	r.Post("/update/", h.UpdateMetricJSON)
	r.Route("/value/{type}", func(r chi.Router) {
		r.Route("/{name}", func(r chi.Router) {
			r.Post("/{value}", h.UpdateMetric)
			r.Get("/", h.GetMetric)
		})
	})

	return r
}
