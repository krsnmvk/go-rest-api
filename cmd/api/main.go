package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/krsnmvk/gorestapi/internal/db"
	"github.com/krsnmvk/gorestapi/internal/handler"
	m "github.com/krsnmvk/gorestapi/internal/middleware"
	"github.com/krsnmvk/gorestapi/pkg/utils"
)

func main() {
	port := ":8080"
	certFile := "cert.pem"
	keyFile := "key.pem"

	mux := http.NewServeMux()

	db := db.NewPostgres()
	for k, v := range db.Health() {
		log.Printf("%s: %s", k, v)
	}

	mux.HandleFunc("/", handler.RootHandler)
	mux.HandleFunc("/teachers/", handler.TeachersHandler)
	mux.HandleFunc("/students/", handler.StudentsHandler)
	mux.HandleFunc("/execs/", handler.ExecsHandler)

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	secureMux := utils.Chain(mux,
		m.CorsMiddleware,
		m.SecurityHeadersMiddleware,
		m.RateLimitMiddleware,
		m.HppMiddleware(m.Reject, map[string]bool{"tags": true, "ids": true}),
		m.GzipMiddleware,
		m.ResponseTimeMiddleware,
	)

	server := &http.Server{
		Addr:         port,
		Handler:      secureMux,
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	log.Println("Server running on https://localhost:8080")
	if err := server.ListenAndServeTLS(certFile, keyFile); err != nil {
		log.Fatalf("Error starting the server: %v\n", err)
	}
}
