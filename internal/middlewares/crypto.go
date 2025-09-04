package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/zubans/metrics/internal/cryptoutil"
)

func DecryptRequestMiddleware(decrypt func(*cryptoutil.Envelope) ([]byte, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Encrypted") != "1" {
				next.ServeHTTP(w, r)
				return
			}

			var env cryptoutil.Envelope
			if err := json.NewDecoder(r.Body).Decode(&env); err != nil {
				http.Error(w, "invalid encrypted payload", http.StatusBadRequest)
				return
			}
			plaintext, err := decrypt(&env)
			if err != nil {
				http.Error(w, "decryption failed", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(plaintext))
			r.ContentLength = int64(len(plaintext))
			r.Header.Set("Content-Encoding", "gzip")
			r.Header.Del("X-Encrypted")
			next.ServeHTTP(w, r)
		})
	}
}
