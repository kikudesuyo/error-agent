// SPDX-License-Identifier: Apache-2.0
package library

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// OpenDatabase creates a PostgreSQL connection using the configured URL.
func OpenDatabase(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL connection: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get PostgreSQL connection pool: %w", err)
	}
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(2)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}

// ValidateDatabaseTransport requires TLS for production database connections.
func ValidateDatabaseTransport(databaseURL string, applicationEnvironment string) error {
	switch applicationEnvironment {
	case "", "development", "test":
		return nil
	case "production":
	default:
		return fmt.Errorf("APP_ENV must be development, test, or production")
	}
	parsedURL, err := url.Parse(databaseURL)
	if err != nil || parsedURL.Scheme != "postgres" && parsedURL.Scheme != "postgresql" {
		return fmt.Errorf("DATABASE_URL must be a PostgreSQL URL")
	}
	sslMode := strings.ToLower(parsedURL.Query().Get("sslmode"))
	switch sslMode {
	case "require", "verify-ca", "verify-full":
		return nil
	default:
		return fmt.Errorf("production DATABASE_URL must require TLS")
	}
}

// ValidateIngestToken rejects credentials too short for a production bearer secret.
func ValidateIngestToken(token string) error {
	if len(token) < 32 {
		return fmt.Errorf("INGEST_TOKEN must contain at least 32 bytes")
	}
	return nil
}

// PingDatabase verifies database connectivity for readiness checks.
func PingDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get PostgreSQL connection pool: %w", err)
	}
	return ping(sqlDB)
}

func ping(db *sql.DB) error { return db.Ping() }
