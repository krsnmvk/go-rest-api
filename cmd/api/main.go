package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	m "github.com/krsnmvk/gorestapi/internal/api/middlewares"
	"github.com/krsnmvk/gorestapi/internal/api/routes"
	"github.com/krsnmvk/gorestapi/internal/database"
	"github.com/krsnmvk/gorestapi/internal/utils"
)

var (
	done = make(chan bool, 1)
	db   = database.NewPostgres()
	pool = db.Pool()
)

func main() {
	port := 8080

	queries := database.NewQueries(pool)

	cert := "cert.pem"
	key := "key.pem"

	for key, value := range db.Health() {
		log.Printf("%s: %s\n", key, value)
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	secureMux := utils.Chain(
		routes.RegisterRoutes(queries, pool),
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

	go gracefulShutdown(server)

	log.Printf("Server running on https://localhost:%d", port)
	if err := server.ListenAndServeTLS(cert, key); err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("Error starting the server: %v\n", err))
	}

	<-done
	log.Println("Graceful shutdown completed.")
}

func gracefulShutdown(apiServer *http.Server) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("Shutting down gracefully.")
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := apiServer.Shutdown(ctx); err != nil {
		panic(fmt.Sprintf("Server force to shutdown with error: %v\n", err))
	}

	if err := db.Close(); err != nil {
		log.Printf("Failed to close the database connection: %v\n", err)
	}

	log.Println("Server exiting.")
	done <- true
}
