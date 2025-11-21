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

func HandlePairing(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	formvalues := r.Form

	sse := datastar.NewSSE(w, r)
	resultingOutput, err := runCommand(formvalues)
	fmt.Println(string(resultingOutput), "\bError =", err)
	if err == nil {
		sse.PatchElementTempl(components.ConnectForm(), datastar.WithSelectorID("pairform"))
	}

	outputSlice := make([]string, 0, 10)
	outputSlice = append(outputSlice, string(resultingOutput))

	locInner := components.CodePen(outputSlice)

	if err := sse.PatchElementTempl(locInner); err != nil {
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
