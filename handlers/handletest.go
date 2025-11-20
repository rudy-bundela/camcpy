// Package handlers contains information about the handlers
package handlers

import (
	"bufio"
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

	// TODO: Form validation

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
	deviceAddress := fmt.Sprintf("%s:%s", formvalues.Get("ipaddr"), formvalues.Get("port"))
	deviceCode := formvalues.Get("code")
	fmt.Printf("Device address = %s; code = %s\n", deviceAddress, deviceCode)

	cmd := exec.Command("adb", "pair", deviceAddress, deviceCode)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Panicf("Error creating StdinPipe: %v\n", err)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Panicf("Error creating StdoutPipe: %v\n", err)
	}
	defer stdout.Close()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Panicf("Error creating StderrPipe: %v\n", err)
	}
	defer stderr.Close()

	if err := cmd.Start(); err != nil {
		log.Panicf("Error starting adb pair command: %v\n", err)
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println("Scanner stdout output = ", scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Println("Scanner stderr output = ", scanner.Text())
		}
	}()

	errScanner := bufio.NewScanner(stderr)

	for errScanner.Scan() {
		fmt.Println("errScanner output = ", errScanner.Text())
	}

	if err := cmd.Wait(); err != nil {
		fmt.Printf("adb pair command finished with error: %v\n", err)
	}

	return "hello"
}
