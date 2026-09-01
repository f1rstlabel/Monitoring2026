package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"sanoc/backend/internal/config"

	_ "github.com/lib/pq"
)

func configureConnectionPool(db *sql.DB) {
	if db == nil {
		return
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)
}

func InitPostgres(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.PostgresDSN()
	log.Printf("[Database] Connecting to PostgreSQL at %s:%s (db=%s)...", cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := sql.Open("postgres", dsn)
	if err == nil {
		if pingErr := db.Ping(); pingErr == nil {
			log.Printf("[Database] PostgreSQL connection established successfully to %s!", cfg.DBName)
			configureConnectionPool(db)
			return db, nil
		}
		db.Close()
	}

	// 1. Try connecting to maintenance DB 'postgres' to auto-create target database if missing
	adminDSN := fmt.Sprintf("postgres://%s@%s:%s/postgres?sslmode=%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBSSLMode)
	if cfg.DBPassword != "" {
		adminDSN = fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBSSLMode)
	}

	adminDB, adminErr := sql.Open("postgres", adminDSN)
	if adminErr == nil {
		if adminPingErr := adminDB.Ping(); adminPingErr == nil {
			log.Printf("[Database] Connected to 'postgres' root DB. Creating '%s' database...", cfg.DBName)
			_, _ = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.DBName))
			adminDB.Close()

			db, err = sql.Open("postgres", dsn)
			if err == nil && db.Ping() == nil {
				log.Printf("[Database] Successfully created and connected to '%s' database!", cfg.DBName)
				configureConnectionPool(db)
				return db, nil
			}
		} else {
			adminDB.Close()
		}
	}

	// 2. Fallback to legacy 'govmonitor' DB if present
	fallbackDSN := fmt.Sprintf("postgres://%s@%s:%s/govmonitor?sslmode=%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBSSLMode)
	if cfg.DBPassword != "" {
		fallbackDSN = fmt.Sprintf("postgres://%s:%s@%s:%s/govmonitor?sslmode=%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBSSLMode)
	}
	fallbackDB, fbErr := sql.Open("postgres", fallbackDSN)
	if fbErr == nil && fallbackDB.Ping() == nil {
		log.Printf("[Database] Connected to legacy 'govmonitor' fallback database!")
		configureConnectionPool(fallbackDB)
		return fallbackDB, nil
	}
	if fallbackDB != nil {
		fallbackDB.Close()
	}

	return nil, fmt.Errorf("failed to connect to postgres database at %s:%s", cfg.DBHost, cfg.DBPort)
}
