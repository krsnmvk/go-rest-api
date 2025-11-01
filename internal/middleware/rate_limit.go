package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type client struct {
	tokens      float64
	lastRequest time.Time
}

var (
	clients      = make(map[string]*client)
	clientsMutex sync.Mutex
	RateLimit    = 10
	rateWindow   = time.Minute
)

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
		if ip == "::1" {
			ip = "127.0.0.1"
		}

		clientsMutex.Lock()
		defer clientsMutex.Unlock()

		c, exists := clients[ip]
		if !exists {
			c = &client{tokens: float64(RateLimit), lastRequest: time.Now()}
			clients[ip] = c
		}

		now := time.Now()
		elapsed := now.Sub(c.lastRequest).Seconds()
		c.tokens += elapsed / rateWindow.Seconds() * float64(RateLimit)
		if c.tokens > float64(RateLimit) {
			c.tokens = float64(RateLimit)
		}
		c.lastRequest = now

		if c.tokens < 1 {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		c.tokens--
		next.ServeHTTP(w, r)
	})
}
