package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	n, err := rr.ResponseWriter.Write(b)
	rr.size += n
	return n, err
}

func ResponseTimeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		w.Header().Set("X-Response-Time", duration.String())
		w.Header().Set("X-Response-Size", fmt.Sprintf("%dB", rec.size))

		log.Printf("%s %s %d %v %dB From %s",
			r.Method, r.URL.Path, rec.statusCode, duration, rec.size, r.RemoteAddr)
	})
}
