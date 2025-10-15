package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	m "github.com/krsnmvk/gorestapi/internal/api/middlewares"
	"github.com/krsnmvk/gorestapi/internal/api/routes"
	"github.com/krsnmvk/gorestapi/internal/database"
	"github.com/krsnmvk/gorestapi/internal/utils"
)

func main() {
	port := 8080

	cert := "cert.pem"
	key := "key.pem"

	db := database.NewPostgres()
	defer db.Close()

	for key, value := range db.Health() {
		log.Printf("%s: %s\n", key, value)
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	secureMux := utils.Chain(
		routes.RegisterRoutes(),
		m.ResponseTimeMiddleware,
		m.CorsMiddleware,
		m.SecurityHeadersMiddleware,
	)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      secureMux,
		TLSConfig:    tlsConfig,
		IdleTimeout:  time.Minute,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
	}

	log.Printf("Server running on https://localhost:%d", port)
	if err := server.ListenAndServeTLS(cert, key); err != nil {
		panic(fmt.Sprintf("Error starting the server: %v\n", err))
	}
}
