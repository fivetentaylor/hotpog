package router

import (
	"net/http"

	"github.com/fivetentaylor/hotpog/internal/handlers"
)

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// routes
	mux.HandleFunc("/", handlers.Home)

	return mux
}
