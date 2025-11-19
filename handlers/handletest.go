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
	fmt.Printf("Device address = %s\n", deviceAddress)

	cmd := exec.Command("echo", "pair", deviceAddress)

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

	fmt.Printf("ADB starting...")

	scanner := bufio.NewScanner(stdout)
	go func() {
		fmt.Println("printing from scanner")
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("ADB output: %s\n", line)
		}
	}()

	errScanner := bufio.NewScanner(stderr)
	go func() {
		for errScanner.Scan() {
			fmt.Printf("ADB Error: %s\n", errScanner.Text())
		}
	}()

	if err := cmd.Wait(); err != nil {
		fmt.Printf("adb pair command finished with error: %v\n", err)
	} else {
		fmt.Println("adb pair command executed successfully.")
	}

	return "hello"
}
