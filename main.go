// main.go
package main

import (
	"fmt"
	"net/http"

	"camcpy/components"

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

	// Your other routes...

	mux.Handle("/", templ.Handler(components.Index()))

	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", mux)
}
