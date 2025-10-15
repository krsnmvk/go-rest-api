package middlewares

import (
	"log"
	"net/http"
	"time"
)

type ResponseTime struct {
	http.ResponseWriter
	statusCode int
}

func (rt *ResponseTime) WeiteHeader(statusCode int) {
	rt.ResponseWriter.WriteHeader(statusCode)
	rt.statusCode = statusCode

}

func ResponseTimeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("RESPONSE TIME START")

		start := time.Now()

		rt := &ResponseTime{
			statusCode:     http.StatusOK,
			ResponseWriter: w,
		}

		next.ServeHTTP(rt, r)

		duration := time.Since(start)

		log.Printf("%s %s %d %s", r.Method, r.RequestURI, rt.statusCode, duration)
	})
}
