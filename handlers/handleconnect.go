package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/rudy-bundela/camcpy/components"

	"github.com/starfederation/datastar-go/datastar"
)

func HandleADBConnect(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Println("Error in parsing form: ", err)
	}

	IPAddress := r.Form.Get("ipaddr")
	Port := r.Form.Get("port")

	sse := datastar.NewSSE(w, r)

	codepenInner := make([]string, 0, 10)

	ADBConnectOutput, err := runADBConnectWithContext(IPAddress, Port)
	ADBConnectOutputString := string(ADBConnectOutput)

	if err != nil {
		log.Println("ADB output: ", ADBConnectOutputString)
		log.Println("ADB connect returned an error: ", err)
		codepenInner = append(codepenInner, err.Error())
	}

	codepenInner = append(codepenInner, ADBConnectOutputString)
	locInner := components.CodePen(codepenInner)

	if err := sse.PatchElementTempl(locInner); err != nil {
		log.Println(err)
	}

	if !strings.Contains(ADBConnectOutputString, "connected to") {
		return
	}

	if err := sse.Redirect("/setupcamera"); err != nil {
		log.Println("Error redirecting to setupcamera, error = ", err)
	}
}

func runADBConnectWithContext(ipaddr, port string) (out []byte, err error) {
	// 5-second timeout for adb connect (definitely needed for some input cases)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "adb", "connect", (ipaddr + ":" + port))

	logstring := fmt.Sprintf("Running command: adb connect %s:%s", ipaddr, port)
	log.Println(logstring)

	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Println("Error running command: ", err)
	}

	if ctx.Err() == context.DeadlineExceeded {
		message := "Command timed out - check the IP address and Port number"
		out = []byte(message)
	}
	log.Println("ADBoutput = ", string(out))
	return out, err
}
