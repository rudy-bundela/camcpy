// Package handlers contains information about the handlers
package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/rudy-bundela/camcpy/components"

	"github.com/starfederation/datastar-go/datastar"
)

func HandlePairing(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println(err)
	}

	IPAddress := r.Form.Get("ipaddr")
	Port := r.Form.Get("port")
	PairingCode := r.Form.Get("code")

	sse := datastar.NewSSE(w, r)

	codepenInner := make([]string, 0, 10)

	ADBPairOutput, err := runADBPairWithContext(IPAddress, Port, PairingCode)
	ADBPairOutputstring := string(ADBPairOutput)

	if err != nil {
		log.Println("ADB pair output: ", ADBPairOutputstring)
		log.Println("ADB pair returned an error: ", err)
		codepenInner = append(codepenInner, ADBPairOutputstring)
		codepenInner = append(codepenInner, err.Error())
		locInner := components.CodePen(codepenInner)
		if err := sse.PatchElementTempl(locInner); err != nil {
			log.Println(err)
		}
		return
	}

	codepenInner = append(codepenInner, ADBPairOutputstring)
	locInner := components.CodePen(codepenInner)

	if err := sse.PatchElementTempl(locInner); err != nil {
		log.Println(err)
	}

	if err := sse.Redirect("/connect"); err != nil {
		log.Println("Error redirecting to connect, error = ", err)
	}
}

func runADBPairWithContext(ipaddr, port, code string) (out []byte, err error) {
	// 5-second timeout for adb pair (might be needed for some input cases)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "adb", "pair", (ipaddr + ":" + port), code)
	logstring := fmt.Sprintf("Running command: adb pair %s:%s %s", ipaddr, port, code)
	log.Println(logstring)

	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error starting adb pair command: %v\n", err)
	}

	if ctx.Err() == context.DeadlineExceeded {
		message := "Command timed out - check the IP address, port number and pairing code"
		out = []byte(message)
	}

	log.Println("ADBoutput = ", string(out))
	return out, err
}
