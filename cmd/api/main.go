package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	m "github.com/krsnmvk/gorestapi/internal/api/middlewares"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/teachers", teachersHandler)
	mux.HandleFunc("/students", studentsHandler)
	mux.HandleFunc("/execs", execsHandler)

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	port := 8080

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      m.ResponseTimeMiddleware(m.CorsMiddleware(m.SecurityHeadersMiddleware(mux))),
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
	}

	cert := "cert.pem"
	key := "key.pem"

	log.Printf("Server running on https://localhost:%d", port)
	if err := server.ListenAndServeTLS(cert, key); err != nil {
		panic(fmt.Sprintf("Error starting the server: %v\n", err))
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Root Route"))
	fmt.Println("Hello Root Route")
}

func teachersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("Hello GET Method on Teachers Route"))
		fmt.Println("Hello GET Method on Teachers Route")

	case http.MethodPost:
		w.Write([]byte("Hello POST Method on Teachers Route"))
		fmt.Println("Hello POST Method on Teachers Route")

	case http.MethodPatch:
		w.Write([]byte("Hello PATCH Method on Teachers Route"))
		fmt.Println("Hello PATCH Method on Teachers Route")

	case http.MethodPut:
		w.Write([]byte("Hello PUT Method on Teachers Route"))
		fmt.Println("Hello PUT Method on Teachers Route")

	case http.MethodDelete:
		w.Write([]byte("Hello DELETE Method on Teachers Route"))
		fmt.Println("Hello DELETE Method on Teachers Route")

	default:
		w.Write([]byte("Method Not Allowed"))
		fmt.Println("Method Not Allowed")
	}
}

func studentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Students Route"))
	fmt.Println("Hello Students Route")
}

func execsHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Execs Route"))
	fmt.Println("Hello Execs Route")
}
