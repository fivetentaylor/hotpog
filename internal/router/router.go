package router

import (
	"net/http"

	"github.com/fivetentaylor/hotpog/internal/config"
	"github.com/fivetentaylor/hotpog/internal/handlers"
)

func NewRouter(config *config.Config) *http.ServeMux {
	mux := http.NewServeMux()
	handler := handlers.NewHandler(config.DBUrl)

	// routes
	mux.HandleFunc("GET /register", handler.RegisterPage)
	mux.HandleFunc("POST /register", handler.Register)
	mux.HandleFunc("GET /login", handler.LoginPage)
	mux.HandleFunc("POST /login", handler.Login)

	mux.Handle("GET /", handler.RequireAuth(http.HandlerFunc(handler.Home)))

	return mux
}
