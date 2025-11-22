package handlers

import (
	"fmt"
	"net/http"
	"os/exec"

	"camcpy/main/components"

	"github.com/starfederation/datastar-go/datastar"
)

func HandleADBConnect(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println("Error in parsing form: ", err)
	}

	IPAddress := r.Form.Get("ipaddr")
	Port := r.Form.Get("port")

	logstring := fmt.Sprintf("IP address + port received = %s:%s", IPAddress, Port)

	sse := datastar.NewSSE(w, r)
	if err := sse.ConsoleLog(logstring); err != nil {
		fmt.Println("Error in SSE console log: ", err)
	}

	ADBoutput, err := runADBConnect(IPAddress, Port)
	if err != nil {
		fmt.Println("ADB output: ", string(ADBoutput))
		fmt.Println("ADB connect returned an error: ", err)
		return
	}

	codepenInner := make([]string, 0, 10)
	codepenInner = append(codepenInner, string(ADBoutput))
	locInner := components.CodePen(codepenInner)

	if err := sse.PatchElementTempl(locInner); err != nil {
		fmt.Println(err)
	}
}

func runADBConnect(ipaddr, port string) (out []byte, err error) {
	cmd := exec.Command("adb", "connect", (ipaddr + ":" + port))
	out, err = cmd.Output()
	if err != nil {
		fmt.Println("Error running command: ", err)
	}
	fmt.Println("Command output: ", string(out))
	fmt.Println("Error running command: ", err)
	return out, err
}
