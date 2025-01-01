// cmd/server/main.go
package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	"github.com/fivetentaylor/hotpog/internal/router"
)

//go:embed static
var static embed.FS

func main() {
	r := router.NewRouter()

	// Strip "static" prefix for serving
	staticFS, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatal(err)
	}

	// Serve static files
	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	log.Println("Server starting on :3333")
	/*
		if err := http.ListenAndServe(":3333", nil); err != nil {
			log.Fatal(err)
		}*/

	if err := http.ListenAndServeTLS(":3333",
		"certs/localhost+1.pem",
		"certs/localhost+1-key.pem",
		r); err != nil {
		log.Fatal(err)
	}
}
