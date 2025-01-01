// cmd/server/main.go
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/fivetentaylor/hotpog/internal/config"
	"github.com/fivetentaylor/hotpog/internal/router"
)

//go:embed static
var static embed.FS

func main() {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("config: %+v\n", c)

	r := router.NewRouter()

	// Strip "static" prefix for serving
	staticFS, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatal(err)
	}

	// Serve static files
	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	fmt.Printf("Server starting on %s\n", c.Port)
	/*
		if err := http.ListenAndServe(":3333", nil); err != nil {
			log.Fatal(err)
		}*/

	if err := http.ListenAndServeTLS(c.Port, c.CertPath, c.KeyPath, r); err != nil {
		log.Fatal(err)
	}
}
