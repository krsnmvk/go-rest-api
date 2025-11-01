package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/krsnmvk/gorestapi/internal/database"
	"github.com/krsnmvk/gorestapi/internal/handler"
	m "github.com/krsnmvk/gorestapi/internal/middleware"
	"github.com/krsnmvk/gorestapi/pkg/utils"
)

func main() {
	port := ":8080"
	certFile := "cert.pem"
	keyFile := "key.pem"

	mux := http.NewServeMux()

	db := database.NewPostgres()
	queries := database.NewQueries(db.Pool())

	for k, v := range db.Health() {
		log.Printf("%s: %s", k, v)
	}

	teachers := handler.NewTeachersHandler(queries)

	mux.HandleFunc("/", handler.RootHandler)
	mux.HandleFunc("/teachers/", teachers.ServeHTTP)
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
