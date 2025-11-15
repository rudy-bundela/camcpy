// main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"camcpy/main/components"
	"camcpy/main/handlers"

	"github.com/a-h/templ"
)

var dev = true

func disableCacheInDevMode(next http.Handler) http.Handler {
	if !dev {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()

	// Serve static assets from filesystem
	mux.Handle("/static/",
		disableCacheInDevMode(
			http.StripPrefix("/static",
				http.FileServer(http.Dir("static")))))

	mux.Handle("/", templ.Handler(components.Index()))
	mux.Handle("/formendpoint", http.HandlerFunc(handlers.HandleTest))

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
