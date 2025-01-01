package handlers

import (
	"fmt"
	"net/http"

	"github.com/fivetentaylor/hotpog/internal/templ/components"
)

func Home(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Handling request for %s\n", r.URL.Path)
	component := components.HelloWorld()
	component.Render(r.Context(), w)
}
