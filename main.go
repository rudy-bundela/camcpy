// main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"camcpy/main/components"
	"camcpy/main/handlers"

	"github.com/a-h/templ"
	"github.com/starfederation/datastar-go/datastar"
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
	mux.Handle("/pairingendpoint", http.HandlerFunc(handlers.HandlePairing))
	mux.Handle("/adbconnect", http.HandlerFunc(handlers.HandleADBConnect))
	mux.Handle("/connectendpoint", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sse := datastar.NewSSE(w, r)
		sse.PatchElementTempl(components.ConnectForm(), datastar.WithSelectorID("pairform"))
		sse.PatchElementTempl(components.CodePen([]string{"Waiting for input..."}))
	}))

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
