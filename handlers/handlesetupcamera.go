package handlers

import (
	"log"
	"net/http"

	"camcpy/main/components"

	"github.com/starfederation/datastar-go/datastar"
)

func HandleSetupCamera(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	layout := components.Layout(components.CodePen([]string{"Hello"}))
	if err := sse.PatchElementTempl(layout); err != nil {
		log.Println("Error patching component: ", err)
	}

	if err := sse.ConsoleLog("Hello today"); err != nil {
		log.Println("Error printing to console: ", err)
	}
}
