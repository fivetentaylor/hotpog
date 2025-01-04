package db

import (
	"database/sql"
	_ "github.com/lib/pq" // Add this import

	sqlc "github.com/fivetentaylor/hotpog/internal/db/generated"
)

type DB struct {
	Raw     *sql.DB
	Queries *sqlc.Queries
}

func New(dbURL string) (*DB, error) {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	return &DB{
		Raw:     db,
		Queries: sqlc.New(db),
	}, nil
}
