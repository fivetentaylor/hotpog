package testutils

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"

	"github.com/fivetentaylor/hotpog/internal/config"
	"github.com/fivetentaylor/hotpog/internal/db"
)

var (
	testDB *db.DB
	once   sync.Once
)

func getFileDirectory() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}

// GetTestDB returns a database connection for testing
func GetTestDB(t *testing.T) *db.DB {
	once.Do(func() {
		var err error
		c := config.Get()

		testDB, err = db.New(c.DBUrl)
		if err != nil {
			log.Fatal(err)
		}

		// Test connection
		err = testDB.Raw.Ping()
		if err != nil {
			log.Fatal("Could not connect to test database:", err)
		}

		fd := getFileDirectory()
		migrationsPath := fmt.Sprintf("file://%s/db/migrations", filepath.Dir(fd))
		fmt.Printf("migrationsPath: %s\n", migrationsPath)

		// drop all tables
		if err := dropAllTables(testDB.Raw); err != nil {
			t.Fatal(err)
		}

		// Run migrations
		m, err := migrate.New(
			migrationsPath,
			c.DBUrl,
		)
		if err != nil {
			log.Fatal("Failed to create migrate instance:", err)
		}

		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Failed to run migrations:", err)
		}
	})

	// Clear all tables before each test
	if err := truncateAllTables(testDB.Raw); err != nil {
		t.Fatal(err)
	}

	return testDB
}

func truncateAllTables(db *sql.DB) error {
	// Get all table names
	rows, err := db.Query(`
		SELECT tablename 
		FROM pg_catalog.pg_tables 
		WHERE schemaname = 'public'
	`)
	if err != nil {
		return fmt.Errorf("failed to get table names: %v", err)
	}
	defer rows.Close()

	// Truncate each table
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %v", err)
		}

		_, err = db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", tableName))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %v", tableName, err)
		}
	}

	return rows.Err()
}

func dropAllTables(db *sql.DB) error {
	// Get all table names in the public schema
	rows, err := db.Query(`
        SELECT tablename 
        FROM pg_tables 
        WHERE schemaname = 'public'
    `)
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}
	defer rows.Close()

	// Build and execute DROP TABLE statements
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return fmt.Errorf("failed to scan table name: %w", err)
		}

		// Drop the table with CASCADE to handle dependencies
		_, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", tableName))
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", tableName, err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating tables: %w", err)
	}

	return nil
}

// RunWithTestDB runs a test with a test database connection
func RunWithTestDB(t *testing.T, testFunc func(db *db.DB)) {
	db := GetTestDB(t)
	testFunc(db)
}
