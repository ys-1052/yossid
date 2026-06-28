package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// 1. Get database URL from env, default to local docker postgres
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5433/yossid?sslmode=disable"
	}

	fmt.Printf("Connecting to database: %s\n", maskURL(dbURL))
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Connected to database successfully.")

	// 2. Create schema_migrations table if not exists
	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}

	// 3. Find migrations directory
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	var upMigrations []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			upMigrations = append(upMigrations, file.Name())
		}
	}
	sort.Strings(upMigrations)

	fmt.Printf("Found %d migrations in %s\n", len(upMigrations), migrationsDir)

	// 4. Apply migrations
	for _, filename := range upMigrations {
		var exists bool
		err = db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)", filename).Scan(&exists)
		if err != nil {
			log.Fatalf("Failed to check if migration %s was applied: %v", filename, err)
		}

		if exists {
			fmt.Printf("Migration %s already applied. Skipping.\n", filename)
			continue
		}

		fmt.Printf("Applying migration: %s...\n", filename)
		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", filename, err)
		}

		// Execute in transaction
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			log.Fatalf("Failed to start transaction: %v", err)
		}

		_, err = tx.ExecContext(ctx, string(content))
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to execute migration %s: %v", filename, err)
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations (filename) VALUES ($1)", filename)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to record migration execution %s: %v", filename, err)
		}

		if err := tx.Commit(); err != nil {
			log.Fatalf("Failed to commit transaction for %s: %v", filename, err)
		}
		fmt.Printf("Migration %s applied successfully.\n", filename)
	}

	fmt.Println("Migrations completed successfully.")
}

func maskURL(url string) string {
	parts := strings.Split(url, "@")
	if len(parts) < 2 {
		return url
	}
	subparts := strings.Split(parts[0], "://")
	if len(subparts) < 2 {
		return "postgres://***:***@" + parts[1]
	}
	return subparts[0] + "://***:***@" + parts[1]
}
