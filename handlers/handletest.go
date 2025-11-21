// Package handlers contains information about the handlers
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"camcpy/main/components"

	"github.com/starfederation/datastar-go/datastar"
)

func HandleTest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	formvalues := r.Form

	// TODO: Form validation

	sse := datastar.NewSSE(w, r)
	resultingOutput, _ := runCommand(formvalues)
	if err := sse.ConsoleLog("Returned from runCommand"); err != nil {
		log.Println(err)
	}
	newComponent := components.CodePen(string(resultingOutput))

	if err := sse.PatchElementTempl(newComponent); err != nil {
		log.Println(err)
	}
}

func runCommand(formvalues url.Values) (resultingOutput []byte, err error) {
	deviceAddress := fmt.Sprintf("%s:%s", formvalues.Get("ipaddr"), formvalues.Get("port"))
	deviceCode := formvalues.Get("code")
	fmt.Printf("Device address = %s; code = %s\n", deviceAddress, deviceCode)

	cmd := exec.Command("adb", "pair", deviceAddress, deviceCode)
	fmt.Println("Running command...")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error starting adb pair command: %v\n", err)
	}

	return output, err
}
