// main.go
package main

import (
	"fmt"
	"log"
	"net/http"

	"camcpy/components"

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

func handleTest(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r)
	sse := datastar.NewSSE(w, r)
	sse.ExecuteScript(`console.log("Hello from server!")`)
	sse.PatchElementTempl(components.Index())
	sse.ExecuteScript(`alert("SSE is cool!")`)
	log.Println(w.Write([]byte(r.Host)))
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
	mux.Handle("/formendpoint", http.HandlerFunc(handleTest))

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
