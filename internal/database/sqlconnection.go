package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
)

type Service interface {
	Health() map[string]string
	Close() error
	Pool() *pgxpool.Pool
}

type postgresService struct {
	pool *pgxpool.Pool
}

var (
	dsn              = os.Getenv("DATABASE_URL")
	postgresInstance *postgresService
)

func NewPostgres() Service {
	if postgresInstance != nil {
		return NewPostgres()
	}

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Failed to parse config: %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}

	return &postgresService{pool: pool}
}

func (ps *postgresService) Pool() *pgxpool.Pool {
	return ps.pool
}

func (ps *postgresService) Close() error {
	log.Println("Disconnected from the database.")
	ps.pool.Close()
	return nil
}
