package middlewares

import (
	"bytes"
	"github.com/zubans/metrics/internal/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

type RequestInfo struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	BODY   string
	Time   string `json:"time"`
}

type ResponseInfo struct {
	Status int `json:"status"`
	Size   int `json:"size"`
}

func (r *loggingResponseWriter) Write(data []byte) (int, error) {
	if r.responseData.status == 0 {
		r.responseData.status = http.StatusOK
	}

	size, err := r.ResponseWriter.Write(data)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(status int) {
	r.ResponseWriter.WriteHeader(status)
	r.responseData.status = status
}

func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		var bodyString string
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				logger.Log.Info("failed to read request body", zap.Error(err))
			} else {
				bodyString = string(bodyBytes)
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		responseData := &responseData{
			status: 0,
			size:   0,
		}

		lw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(lw, r)

		duration := time.Since(start)
		logger.Log.Info("got incoming HTTP request",
			zap.Any("response", ResponseInfo{
				Status: responseData.status,
				Size:   responseData.size,
			}),
			zap.Any("request", RequestInfo{
				Method: r.Method,
				URL:    r.URL.String(),
				BODY:   bodyString,
				Time:   duration.String(),
			}),
		)
	})
}
