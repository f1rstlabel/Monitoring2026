package repository

import (
	"database/sql"
	"fmt"
	"log"

	"govmonitor-it/backend/internal/config"

	_ "github.com/lib/pq"
)

func InitPostgres(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.PostgresDSN()
	log.Printf("[Database] Connecting to PostgreSQL at %s:%s (db=%s)...", cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize postgres driver: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres database at %s:%s: %w", cfg.DBHost, cfg.DBPort, err)
	}

	log.Printf("[Database] PostgreSQL connection established successfully!")
	return db, nil
}
