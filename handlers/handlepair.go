// Package handlers contains information about the handlers
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"

	"github.com/rudy-bundela/camcpy/components"

	"github.com/starfederation/datastar-go/datastar"
)

func HandlePairing(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}
	formvalues := r.Form
	outputSlice := make([]string, 0, 10)

	sse := datastar.NewSSE(w, r)
	resultingOutput, err := runCommand(formvalues)
	log.Println(string(resultingOutput), "\bError =", err)
	if err == nil {
		if err := sse.PatchElementTempl(components.ConnectForm(), datastar.WithSelectorID("pairform")); err != nil {
			log.Println("Error in patching ConnectForm component: ", err)
		}
	}

	outputSlice = append(outputSlice, string(resultingOutput))
	outputSlice = append(outputSlice, err.Error())

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
	log.Println("Running command...")

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error starting adb pair command: %v\n", err)
	}

	return output, err
}
