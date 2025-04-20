package router

import (
	"compress/gzip"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/zubans/metrics/internal/handler"
	"net/http"
	"strings"
)

func GetRouter(h *handler.Handler) http.Handler {
	r := chi.NewRouter()

	r.With(middleware.Compress(5, "text/html")).Get("/", h.ShowMetrics)
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetric)
	r.Route("/value/{type}", func(r chi.Router) {
		r.Route("/{name}", func(r chi.Router) {
			r.Post("/{value}", h.UpdateMetric)
			r.Get("/", h.GetMetric)
		})
	})
	r.With(GzipMiddleware).Post("/update/", h.UpdateMetricJSON)
	r.Post("/value/", h.GetMetricJSON)

	return r
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "Invalid gzip body", http.StatusBadRequest)
				return
			}

			defer func(gzReader *gzip.Reader) {
				err := gzReader.Close()
				if err != nil {
					http.Error(w, "Error close gzReader", http.StatusBadRequest)
				}
			}(gzReader)
			r.Body = gzReader

			w.Header().Set("Content-Encoding", "gzip")
			gw := gzip.NewWriter(w)
			defer func(gw *gzip.Writer) {
				err := gw.Close()
				if err != nil {
					http.Error(w, "Error close gzWriter", http.StatusBadRequest)
				}
			}(gw)

			gzipWriter := gzipResponseWriter{Writer: gw, ResponseWriter: w}
			next.ServeHTTP(gzipWriter, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
