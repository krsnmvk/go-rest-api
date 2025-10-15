package middlewares

import (
	"log"
	"net/http"
)

func SecurityHeadersMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("SECURITY HEADERS START")

		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		w.Header().Set("Permissions-Policy", "geolocation=(), camera=(), microphone=()")

		next.ServeHTTP(w, r)
	})
}
