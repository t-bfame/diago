package server

import (
	"net/http"

	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
)

// NewUIBox Creates a packr box for diago-ui
func NewUIBox(router *mux.Router) {
	box := packr.New("diago-ui", "../../dist")

	router.Handle("/", http.FileServer(box))
	router.PathPrefix("/static/").Handler(http.FileServer(box))
}
