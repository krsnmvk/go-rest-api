package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Service interface {
	Close() error
	Health() map[string]string
	Pool() *postgresService
}

type postgresService struct {
	pool *pgxpool.Pool
}

var postgresInstance *postgresService

func NewPostgres() *postgresService {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("could not load .env file: %v", err)
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	if postgresInstance != nil {
		return postgresInstance
	}

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Failed parse PostgreSQL config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	return &postgresService{
		pool: pool,
	}
}

func (ps *postgresService) Pool() *pgxpool.Pool {
	return ps.pool
}

func (ps *postgresService) Close() error {
	ps.pool.Close()
	log.Println("Postgres connection pool closed successfully")
	return nil
}
