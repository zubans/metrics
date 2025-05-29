package middlewares

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
)

func HashCheck(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			hashHeader := r.Header.Get("HashSHA256")
			if hashHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "can't read body", http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			h := hmac.New(sha256.New, []byte(key))
			h.Write(body)
			expectedHash := hex.EncodeToString(h.Sum(nil))

			if expectedHash != hashHeader {
				http.Error(w, "incorrect header hash", http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
