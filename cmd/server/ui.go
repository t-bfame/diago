package server

import (
	"net/http"

	"github.com/gobuffalo/packr/v2"
	"github.com/gorilla/mux"
)

// NewUIBox Creates a packr box for diago-ui
func NewUIBox(router *mux.Router) {
	box := packr.New("diago-ui", "../../dist")

	router.PathPrefix("/static/").Handler(http.FileServer(box))
	router.PathPrefix("/asset-manifest.json").Handler(http.FileServer(box))
	router.PathPrefix("/favicon.ico").Handler(http.FileServer(box))
	router.PathPrefix("/manifest.json").Handler(http.FileServer(box))
	router.PathPrefix("/robots.txt").Handler(http.FileServer(box))

	index, _ := box.Open("index.html")
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs, _ := index.Stat()
		http.ServeContent(w, r, "index.html", fs.ModTime(), index)
	})
}
