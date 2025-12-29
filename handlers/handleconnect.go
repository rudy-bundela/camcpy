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

	logstring := fmt.Sprintf("IP address + port received = %s:%s", IPAddress, Port)

	sse := datastar.NewSSE(w, r)
	if err := sse.ConsoleLog(logstring); err != nil {
		log.Println("Error in SSE console log: ", err)
	}

	codepenInner := make([]string, 0, 10)

	ADBoutput, err := runADBConnectWithContext(IPAddress, Port)
	ADBoutputstring := string(ADBoutput)

	if err != nil {
		log.Println("ADB output: ", ADBoutputstring)
		log.Println("ADB connect returned an error: ", err)
		codepenInner = append(codepenInner, err.Error())
	}

	codepenInner = append(codepenInner, ADBoutputstring)
	locInner := components.CodePen(codepenInner)

	if err := sse.PatchElementTempl(locInner); err != nil {
		log.Println(err)
	}

	if !strings.Contains(ADBoutputstring, "connected to") {
		return
	}

	if err := sse.Redirect("/setupcamera"); err != nil {
		log.Println("Error redirecting to setupcamera, error = ", err)
	}
}

func runADBConnectWithContext(ipaddr, port string) (out []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "adb", "connect", (ipaddr + ":" + port))

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
