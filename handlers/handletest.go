// Package handlers contains information about the handlers
package handlers

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"strings"

	"github.com/starfederation/datastar-go/datastar"
)

func HandleTest(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}
	fmt.Println(r.Form)
	formvalues := r.Form

	sse := datastar.NewSSE(w, r)
	resultingOutput := runCommand(formvalues)
	alertOutput := fmt.Sprintf("console.log('%s')", resultingOutput)
	alertOutput = strings.ReplaceAll(alertOutput, "\n", "")
	fmt.Println(alertOutput)

	if err := sse.ExecuteScript(alertOutput); err != nil {
		log.Fatal(err)
	}
}

func runCommand(formvalues url.Values) (resultingoutput string) {
	for k, v := range formvalues {
		fmt.Println(k, v)
	}
	cmd := exec.Command("echo", formvalues.Get("port"), formvalues.Get("code"))

	output, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Command output: %s", output)
	return string(output)
}
