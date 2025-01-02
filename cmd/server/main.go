// cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/fivetentaylor/hotpog/internal/config"
	"github.com/fivetentaylor/hotpog/internal/router"
)

func main() {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	r := router.NewRouter(c)

	fmt.Printf("Server starting on %s\n", c.Port)
	/*
		if err := http.ListenAndServe(":3333", nil); err != nil {
			log.Fatal(err)
		}*/

	if err := http.ListenAndServeTLS(c.Port, c.CertPath, c.KeyPath, r); err != nil {
		log.Fatal(err)
	}
}
