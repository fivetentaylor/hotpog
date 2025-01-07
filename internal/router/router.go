package router

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/fivetentaylor/hotpog/internal/config"
	"github.com/fivetentaylor/hotpog/internal/db"
	"github.com/fivetentaylor/hotpog/internal/handlers"
)

//go:embed static
var static embed.FS

func NewRouter(config *config.Config) *http.ServeMux {
	mux := http.NewServeMux()
	db, err := db.New(config.DBUrl)
	if err != nil {
		panic(err)
	}

	handler := handlers.NewHandler(db)

	// routes
	mux.HandleFunc("GET /login", handler.LoginPage)
	mux.HandleFunc("GET /register", handler.RegisterPage)
	mux.HandleFunc("POST /register", handler.Register)
	mux.HandleFunc("POST /login", handler.Login)
	mux.HandleFunc("POST /login/magic", handler.SendMagicLink)
	mux.HandleFunc("GET /logout", handler.Logout)
	mux.HandleFunc("GET /verify", handler.VerifyUserEmail)
	mux.Handle("POST /verify", handler.RequireAuth(http.HandlerFunc(handler.VerifyUserEmail)))

	// Serve static files
	staticFS, err := fs.Sub(static, "static")
	if err != nil {
		panic(err)
	}

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Home route
	mux.Handle("GET /", handler.RequireAuth(http.HandlerFunc(handler.Home)))

	return mux
}
