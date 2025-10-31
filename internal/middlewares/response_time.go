package middlewares

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	n, err := sr.ResponseWriter.Write(b)
	sr.size += n
	return n, err
}

func ResponseTimeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)
		responseTime := duration.Milliseconds()

		w.Header().Set("X-Response-Time", time.Duration(responseTime*int64(time.Millisecond)).String())
		w.Header().Set("X-Response-Size", fmt.Sprint(recorder.size))

		log.Printf("%s %s %d %v %dB From %s",
			r.Method,
			r.URL.Path,
			recorder.statusCode,
			duration,
			recorder.size,
			r.RemoteAddr,
		)
	})
}
