package handlers

import (
	"database/sql"
	_ "github.com/lib/pq" // Add this import

	sqlc "github.com/fivetentaylor/hotpog/internal/db/generated"
)

type Handler struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewHandler(dbURL string) *Handler {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}

	return &Handler{
		db:      db,
		queries: sqlc.New(db),
	}
}
