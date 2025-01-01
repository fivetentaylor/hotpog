// cmd/server/main.go
package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/fivetentaylor/hotpog/templates/components"
)

//go:embed static
var static embed.FS

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Handling request for %s\n", r.URL.Path)
		component := components.HelloWorld()
		component.Render(r.Context(), w)
	})

	// Strip "static" prefix for serving
	staticFS, err := fs.Sub(static, "static")
	if err != nil {
		log.Fatal(err)
	}

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	log.Println("Server starting on :3333")
	/*
		if err := http.ListenAndServe(":3333", nil); err != nil {
			log.Fatal(err)
		}*/

	if err := http.ListenAndServeTLS(":3333",
		"certs/localhost+1.pem",
		"certs/localhost+1-key.pem",
		nil); err != nil {
		log.Fatal(err)
	}
}
