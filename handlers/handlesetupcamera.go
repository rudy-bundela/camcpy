package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os/exec"

	"camcpy/main/components"

	"github.com/starfederation/datastar-go/datastar"
)

func HandleSetupCamera(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r)

	layout := components.Layout(components.CodePen([]string{"Fetching information..."}))
	if err := sse.PatchElementTempl(layout); err != nil {
		log.Println("Error patching component: ", err)
	}

	outputStringSlice := make([]string, 10)

	output, err := runGetScrcpyDetails()
	if err == nil {
		scrcpyOutput, err := ParseScrcpyOutput(string(output))
		if err != nil {
			log.Printf("Error parsing output from scrcpy: %v\n", err)
		}
		jsonData, _ := json.MarshalIndent(scrcpyOutput, "", "	")
		outputStringSlice = append(outputStringSlice, string(jsonData))
		if err := sse.ConsoleLogf("Output from scrcpy: %v", outputStringSlice); err != nil {
			log.Println("Error sending log to console: ", err)
		}
		if err := sse.PatchElementTempl(components.CodePen(outputStringSlice)); err != nil {
			log.Println("Error patching codepen component with scrcpy details: ", err)
		}
	} else {
		if err := sse.PatchElementTempl(components.CodePen([]string{"Scrcpy returned an error"})); err != nil {
			log.Println("Error patching codepen component with scrcpy details: ", err)
		}
		if err := sse.ConsoleLogf("Scrcpy returned an error: %v", err); err != nil {
			log.Println("Error sending log message to console: ", err)
		}
	}
}

func runGetScrcpyDetails() (output []byte, err error) {
	cmd := exec.Command("scrcpy", "--list-camera-sizes")
	output, err = cmd.Output()
	log.Printf("Output from scrcpy: %v", string(output))
	return output, err
}
