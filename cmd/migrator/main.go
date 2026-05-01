package main

import (
	"errors"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	dbURL := getEnv("DATABASE_URL", "postgres://uptime_user:supersecretpassword@localhost:5434/uptime_monitor?sslmode=disable")
	migrationsPath := getEnv("MIGRATIONS_PATH", "file://migrations")

	log.Printf("Connecting to DB: %s", dbURL)
	log.Printf("Using migrations from %s", migrationsPath)

	m, err := migrate.New(migrationsPath, dbURL)

	if err != nil {
		log.Fatalf("Failed to initialize migrator: %v", err)
	}

	defer m.Close()

	log.Println("Applying database migrations...")

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Database is already up to date! No migrations applied. ")
		} else {
			log.Fatalf("Failed to apply migrations: %v", err)
		}
	} else {
		log.Println("Migrations applied successufully")
	}
}

// universal function getEnv
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
