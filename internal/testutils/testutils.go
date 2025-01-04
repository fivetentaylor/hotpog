package testutils

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"testing"

	_ "github.com/lib/pq"

	"github.com/fivetentaylor/hotpog/internal/config"
	"github.com/fivetentaylor/hotpog/internal/db"
)

var (
	testDB *db.DB
	once   sync.Once
)

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

// RunWithTestDB runs a test with a test database connection
func RunWithTestDB(t *testing.T, testFunc func(db *db.DB)) {
	db := GetTestDB(t)
	testFunc(db)
}
