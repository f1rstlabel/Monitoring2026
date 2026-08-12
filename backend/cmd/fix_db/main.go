package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASSWORD", "")
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")

	// Connect to default root DB 'postgres'
	var dsn string
	if pass != "" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/postgres?sslmode=disable", user, pass, host, port)
	} else {
		dsn = fmt.Sprintf("postgres://%s@%s:%s/postgres?sslmode=disable", user, host, port)
	}

	log.Printf("[DB Fixer] Connecting to PostgreSQL root database at %s:%s...", host, port)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to open postgres driver: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping PostgreSQL root DB: %v", err)
	}
	log.Println("[DB Fixer] Connected to PostgreSQL root successfully!")

	// Check existing databases
	var hasGovmonitor, hasSanoc bool
	rows, err := db.Query("SELECT datname FROM pg_database WHERE datname IN ('govmonitor', 'sanoc')")
	if err != nil {
		log.Fatalf("Failed to query pg_database: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			if name == "govmonitor" {
				hasGovmonitor = true
			}
			if name == "sanoc" {
				hasSanoc = true
			}
		}
	}

	log.Printf("[DB Fixer] Existing databases check -> govmonitor: %v, sanoc: %v", hasGovmonitor, hasSanoc)

	if hasSanoc {
		log.Println("[DB Fixer] ✅ Database 'sanoc' already exists! Adminer and app can connect directly to 'sanoc'.")
		return
	}

	if hasGovmonitor && !hasSanoc {
		log.Println("[DB Fixer] Renaming database 'govmonitor' -> 'sanoc'...")

		// Terminate existing connections to 'govmonitor' so ALTER DATABASE can succeed
		_, _ = db.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'govmonitor' AND pid <> pg_backend_pid();")

		// Rename database
		_, err := db.Exec("ALTER DATABASE govmonitor RENAME TO sanoc;")
		if err != nil {
			log.Fatalf("Failed to rename database govmonitor to sanoc: %v", err)
		}

		log.Println("[DB Fixer] 🎉 SUCCESS! Database 'govmonitor' has been renamed to 'sanoc' in PostgreSQL!")
		return
	}

	if !hasSanoc && !hasGovmonitor {
		log.Println("[DB Fixer] Creating database 'sanoc'...")
		_, err := db.Exec("CREATE DATABASE sanoc;")
		if err != nil {
			log.Fatalf("Failed to create database sanoc: %v", err)
		}
		log.Println("[DB Fixer] 🎉 SUCCESS! Database 'sanoc' has been created in PostgreSQL!")
	}
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}
