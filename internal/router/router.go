package router

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/fivetentaylor/hotpog/internal/config"
	"github.com/fivetentaylor/hotpog/internal/handlers"
)

//go:embed static
var static embed.FS

func NewRouter(config *config.Config) *http.ServeMux {
	mux := http.NewServeMux()
	handler := handlers.NewHandler(config.DBUrl)

	// routes
	mux.HandleFunc("GET /register", handler.RegisterPage)
	mux.HandleFunc("POST /register", handler.Register)
	mux.HandleFunc("GET /login", handler.LoginPage)
	mux.HandleFunc("POST /login", handler.Login)

	// Strip "static" prefix for serving
	staticFS, err := fs.Sub(static, "static")
	if err != nil {
		panic(err)
	}

	// Serve static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	mux.Handle("GET /", handler.RequireAuth(http.HandlerFunc(handler.Home)))

	return mux
}
