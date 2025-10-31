package middlewares

import (
	"log"
	"net/http"
	"slices"
)

var allowedOrigins = []string{
	"http://localhost:3000",
	"http://localhost:5713",
}

func isAllowedOrigin(origin string) bool {
	return slices.Contains(allowedOrigins, origin)
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		if origin == "" && !isAllowedOrigin(origin) {
			log.Printf("CORS blocked: origin %s not allowed\n", origin)
			http.Error(w, "Forbidden - CORS", http.StatusForbidden)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
