// main.go
package main

import (
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

func handleConnectedEndpoint(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)
	if err := sse.PatchElementTempl(components.ConnectForm(), datastar.WithSelectorID("pairform")); err != nil {
		log.Println("Error patching form templ component: ", err)
	}
	if err := sse.PatchElementTempl(components.CodePen([]string{"Waiting for input..."})); err != nil {
		log.Println("Error patching codepen templ component: ", err)
	}
}

func main() {
	mux := http.NewServeMux()

	// Serve static assets from filesystem
	mux.Handle("/static/",
		disableCacheInDevMode(
			http.StripPrefix("/static",
				http.FileServer(http.Dir("static")))))

	scrcpyStruct := components.ScrcpyInfo{}
	// Main routes
	// TODO: create proper components for this section
	mux.Handle("/", templ.Handler(components.Welcome()))
	mux.Handle("/pair", templ.Handler(components.PairformComponent()))
	mux.Handle("/connect", templ.Handler(components.ConnectformComponent()))

	mux.Handle("/cameraoptions", http.HandlerFunc(handlers.HandlePairing))
	mux.Handle("/mediamtxoptions", http.HandlerFunc(handlers.HandlePairing))
	mux.Handle("/monitor", http.HandlerFunc(handlers.HandlePairing))

	// Datastar handlers
	mux.Handle("/pairingendpoint", http.HandlerFunc(handlers.HandlePairing))
	mux.Handle("/adbconnect", http.HandlerFunc(handlers.HandleADBConnect))
	mux.Handle("/setupcamerasse", http.HandlerFunc(scrcpyStruct.HandleGetCameraOptions))
	mux.Handle("/setupcamera", templ.Handler(components.Layout(components.SetupCamera())))
	mux.Handle("/connectendpoint", http.HandlerFunc(handleConnectedEndpoint))

	// Camera specific datastar handlers
	mux.Handle("/camera/idupdate", http.HandlerFunc(scrcpyStruct.HandleCameraIDUpdate))
	mux.Handle("/camera/fpsupdate", http.HandlerFunc(scrcpyStruct.HandleCameraFPSUpdate))
	mux.Handle("/camera/resolutionupdate", http.HandlerFunc(scrcpyStruct.HandleCameraResolutionUpdate))
	mux.Handle("POST /camera/startstream", http.HandlerFunc(scrcpyStruct.HandleStartStream))
	mux.Handle("/printstruct", http.HandlerFunc(scrcpyStruct.PrintStruct))

	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
