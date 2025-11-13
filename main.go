// main.go
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

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
	// sse.ExecuteScript(`console.log("Hello from server!")`)
	// sse.ExecuteScript(`alert("SSE is cool!")`)
	updateCount := rand.Intn(10) + 3
	sse.PatchElements(fmt.Sprintf("<pre><code>Sending down %d numbers</code></pre>", updateCount),
		datastar.WithSelectorID("appendhere"), datastar.WithModeInner(), datastar.WithModeAppend())
	for range updateCount {
		sse.PatchElements(fmt.Sprintf("<pre><code>%d</code></pre>", rand.Intn(100)),
			datastar.WithSelectorID("appendhere"), datastar.WithModeInner(), datastar.WithModeAppend())
		time.Sleep(1000 * time.Millisecond)
	}
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
