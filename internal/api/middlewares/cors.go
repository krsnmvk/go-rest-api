package middlewares

import (
	"log"
	"net/http"
	"slices"
)

func isOriginAllowed(origin string) bool {
	allowedOrigins := []string{
		"http://localhost:3000",
		"https://localhost:8080",
		"http://localhost:5173",
	}

	return slices.Contains(allowedOrigins, origin)
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("CORS START")

		origin := r.Header.Get("Origin")

		if !isOriginAllowed(origin) {
			http.Error(w, "CORS error: Origin not allowed.", http.StatusForbidden)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST PATCH, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.Header().Set("Access-Control-Expose-header", "Ahthorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
