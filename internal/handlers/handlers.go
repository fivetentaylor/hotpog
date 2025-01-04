package handlers

import (
	"github.com/fivetentaylor/hotpog/internal/db"
)

type Handler struct {
	DB *db.DB
}

func NewHandler(db *db.DB) *Handler {
	return &Handler{DB: db}
}
