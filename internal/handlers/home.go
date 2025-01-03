package handlers

import (
	"fmt"
	"net/http"

	"github.com/fivetentaylor/hotpog/internal/templ/components"
)

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Handling request for %s\n", r.URL.Path)
	component := components.HomePage()
	component.Render(r.Context(), w)
}
