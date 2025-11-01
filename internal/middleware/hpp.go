package middleware

import (
	"log"
	"net/http"
	"net/url"
)

type Mode int

const (
	Reject Mode = iota
	KeepFirst
	KeepLast
	JoinComma
)

func HppMiddleware(mode Mode, allowDuplicates map[string]bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if err := r.ParseForm(); err != nil {
				log.Printf("HPP blocked: failed to parse form: %v\n", err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			cleaned := make(url.Values)

			for key, vals := range r.Form {

				if len(vals) <= 1 {
					cleaned[key] = vals
					continue
				}

				if allowDuplicates[key] {
					cleaned[key] = vals
					continue
				}

				switch mode {
				case Reject:
					log.Printf("HPP blocked: duplicate parameter '%s' detected with values %v from %s\n",
						key, vals, r.RemoteAddr)
					http.Error(w, "Duplicate parameter detected", http.StatusBadRequest)
					return

				case KeepFirst:
					cleaned[key] = []string{vals[0]}
					log.Printf("HPP adjusted: kept first value '%s' for parameter '%s' from %s\n",
						vals[0], key, r.RemoteAddr)

				case KeepLast:
					cleaned[key] = []string{vals[len(vals)-1]}
					log.Printf("HPP adjusted: kept last value '%s' for parameter '%s' from %s\n",
						vals[len(vals)-1], key, r.RemoteAddr)

				case JoinComma:
					joined := join(vals)
					cleaned[key] = []string{joined}
					log.Printf("HPP adjusted: joined values %v into '%s' for parameter '%s' from %s\n",
						vals, joined, key, r.RemoteAddr)
				}
			}

			r.Form = cleaned
			r.PostForm = cleaned

			next.ServeHTTP(w, r)
		})
	}
}

func join(vals []string) string {
	out := ""
	for i, v := range vals {
		if i > 0 {
			out += ","
		}
		out += v
	}
	return out
}
