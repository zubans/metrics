package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/zubans/metrics/internal/handler"
	"github.com/zubans/metrics/internal/middlewares"
	"net/http"
)

func GetRouter(h *handler.Handler) http.Handler {
	r := chi.NewRouter()

	r.With(middlewares.GzipMiddleware).Get("/", h.ShowMetrics)
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetric)
	r.Route("/value/{type}", func(r chi.Router) {
		r.Route("/{name}", func(r chi.Router) {
			r.Post("/{value}", h.UpdateMetric)
			r.Get("/", h.GetMetric)
		})
	})
	r.With(middlewares.GzipMiddleware).Post("/update/", h.UpdateMetricJSON)
	r.Post("/value/", h.GetMetricJSON)
	r.Get("/ping", h.PingServer)

	return r
}
