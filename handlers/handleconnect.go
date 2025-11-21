package handlers

import (
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
)

func HandleADBConnect(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	IPAddress := r.Form.Get("ipaddr")
	Port := r.Form.Get("port")
	sse := datastar.NewSSE(w, r)
	sse.ConsoleLog("This is working")
	sse.ConsoleLog(string(IPAddress + ":" + Port))
}
