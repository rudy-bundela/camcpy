package components

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/starfederation/datastar-go/datastar"
)

// --- Helper Functions ---

func readSignals(r *http.Request) (*DatastarSignalsStruct, error) {
	signals := &DatastarSignalsStruct{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		return nil, fmt.Errorf("datastar error reading signals: %w", err)
	}
	return signals, nil
}

func newSSE(w http.ResponseWriter, r *http.Request) *datastar.ServerSentEventGenerator {
	return datastar.NewSSE(w, r, datastar.WithCompression(datastar.WithBrotli(datastar.WithBrotliLGWin(0))))
}

// --- ScrcpyInfo Methods ---

func (s *ScrcpyInfo) HandleGetCameraOptions(w http.ResponseWriter, r *http.Request) {
	sse := newSSE(w, r)
	signals, err := readSignals(r)
	if err != nil {
		log.Println("Datastar error reading signals in HandleGetCameraOptions:", err)
	}

	if s.DeviceName != "" {
		return
	}

	scrcpyOutput, err := RunGetScrcpyDetails()
	if err != nil {
		runOnScrcpyError(sse, err)
		return
	}

	if err := s.ParseScrcpyOutput(string(scrcpyOutput)); err != nil {
		log.Println("Error parsing scrcpy output:", err)
	}

	if err := sse.PatchElementTempl(Layout(CameraComponent(s, signals))); err != nil {
		log.Println("Error patching Layout")
	}
}

func (s *ScrcpyInfo) HandleCameraIDUpdate(w http.ResponseWriter, r *http.Request) {
	signals, err := readSignals(r)
	if err != nil {
		log.Println("Error reading signals:", err)
	}
	sse := newSSE(w, r)

	SetCameraID(sse, signals, s)
	SetCameraFPS(sse, signals, s)
	SetCameraResolution(sse, signals, s)

	if err := sse.MarshalAndPatchSignals(signals); err != nil {
		log.Println("Error patching signals:", err)
	}
}

func (s *ScrcpyInfo) HandleCameraFPSUpdate(w http.ResponseWriter, r *http.Request) {
	signals, err := readSignals(r)
	if err != nil {
		log.Println("Error reading signals:", err)
	}
	sse := newSSE(w, r)

	SetCameraFPS(sse, signals, s)
	SetCameraResolution(sse, signals, s)

	if err := sse.MarshalAndPatchSignals(signals); err != nil {
		log.Println("Error patching signals:", err)
	}
}

func (s *ScrcpyInfo) HandleCameraResolutionUpdate(w http.ResponseWriter, r *http.Request) {
	signals, err := readSignals(r)
	if err != nil {
		log.Println("Error reading signals:", err)
	}
	sse := newSSE(w, r)

	SetCameraResolution(sse, signals, s)

	if err := sse.MarshalAndPatchSignals(signals); err != nil {
		log.Println("Error patching signals:", err)
	}
}

func (s *ScrcpyInfo) HandleStartStream(w http.ResponseWriter, r *http.Request) {
	signals, err := readSignals(r)
	if err != nil {
		log.Printf("Datastar error: %v", err)
		return
	}

	sse := newSSE(w, r)

	// 1. If a stream is already running, cancel it
	if s.cancelStream != nil {
		s.cancelStream()
	}

	// 2. Create a new context for this specific stream
	ctx, cancel := context.WithCancel(context.Background())
	s.cancelStream = cancel

	// 3. Build the argument slice
	args := GetBaseScrcpyArgs()
	args = append(args, fmt.Sprintf("--camera-fps=%d", signals.Fps))
	args = append(args, fmt.Sprintf("--camera-id=%s", signals.CamID))

	resLabel := signals.Resolution
	if strings.Contains(resLabel, " (high-speed)") {
		resLabel, _ = strings.CutSuffix(resLabel, " (high-speed)")
		args = append(args, "--camera-high-speed")
	}
	args = append(args, fmt.Sprintf("--camera-size=%s", resLabel))

	// 4. Create Command with Context
	cmd := exec.CommandContext(ctx, "scrcpy", args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = os.Stdout // Pipe to server logs for visibility

	log.Printf("Executing: scrcpy %s", strings.Join(args, " "))

	if err := cmd.Start(); err != nil {
		s.pushStatus(sse, "Failed to start: "+err.Error(), "text-red-600")
		log.Printf("Startup error: %v", err)
		return
	}

	message := fmt.Sprintf("Scrcpy streaming from camera ID: %s, at a resolution of %s, with a target of %d fps", signals.CamID, resLabel, signals.Fps)

	s.pushStatus(sse, message, "text-green-600")

	// 5. Background monitor
	err = cmd.Wait()
	if err != nil {
		if ctx.Err() == context.Canceled {
			s.pushStatus(sse, "Scrcpy stopped", "text-gray-500")
			log.Println("Scrcpy stream stopped by user/new request.")
		} else {
			s.pushStatus(sse, "ERROR: Scrcpy disconnected", "text-red-600")
			log.Printf("Scrcpy exited with error: %v\nStderr: %s", err, stderr.String())
		}
	}
}

func (s *ScrcpyInfo) pushStatus(sse *datastar.ServerSentEventGenerator, message string, colourClass string) {
	if err := sse.PatchElements(
		fmt.Sprintf(`<div id="stream-status" class="font-bold %s">%s</div>`, colourClass, message),
	); err != nil {
		log.Println("Error pushing status: ", err.Error())
	}
}

func (s *ScrcpyInfo) HandleStopStream(w http.ResponseWriter, r *http.Request) {
	if s.cancelStream != nil {
		log.Println("Stopping scrcpy stream...")
		s.cancelStream()
		s.cancelStream = nil
		return
	}

	sse := newSSE(w, r)
	s.pushStatus(sse, "No stream running", "text-gray-500")
	time.Sleep(2 * time.Second)
	s.pushStatus(sse, "But ready to stream :) ", "text-green-600")
}

func (s *ScrcpyInfo) PrintStruct(w http.ResponseWriter, r *http.Request) {
	jsonoutput, _ := json.Marshal(s)
	fmt.Println(string(jsonoutput))
	if _, err := w.Write(jsonoutput); err != nil {
		log.Println("Error writing json output in PrintStruct test method: ", err)
	}
}

// --- Data Logic Helpers ---

func SetCameraFPS(sse *datastar.ServerSentEventGenerator, signals *DatastarSignalsStruct, s *ScrcpyInfo) {
	fpslist := s.GetCameraFromID(signals.CamID).GetAvailableFPS()
	if err := sse.PatchElementTempl(CameraFPSComponent(fpslist)); err != nil {
		fmt.Println("Error patching FPS component", err)
	}

	if !slices.Contains(fpslist, signals.Fps) && len(fpslist) > 0 {
		signals.Fps = fpslist[len(fpslist)-1]
	}
}

func SetCameraResolution(sse *datastar.ServerSentEventGenerator, signals *DatastarSignalsStruct, s *ScrcpyInfo) {
	resolutions := s.GetCameraFromID(signals.CamID).GetResolutionsForFPS(signals.Fps)
	if err := sse.PatchElementTempl(CameraResolutionComponent(resolutions)); err != nil {
		fmt.Println("Error patching Resolution component", err)
	}

	hasHighRes := slices.ContainsFunc(resolutions, func(r ResolutionOption) bool {
		return strings.Contains(r.Label, "1920x1080 (high-speed)")
	})

	if hasHighRes {
		signals.Resolution = "1920x1080 (high-speed)"
	} else if len(resolutions) > 0 {
		signals.Resolution = resolutions[0].Label
	}
}

func SetCameraID(sse *datastar.ServerSentEventGenerator, signals *DatastarSignalsStruct, s *ScrcpyInfo) {
	newCameras := s.GetCameraFromPosition(signals.Position)
	if err := sse.PatchElementTempl(CameraIDComponent(newCameras)); err != nil {
		fmt.Println("Error patching ID component", err)
	}

	if len(newCameras) > 0 {
		signals.CamID = newCameras[0].ID
	}
}

func (s *ScrcpyInfo) GetCameraFromPosition(position string) []Camera {
	var list []Camera
	for _, c := range s.Cameras {
		if c.Position == position {
			list = append(list, c)
		}
	}
	return list
}

func (s *ScrcpyInfo) GetCameraFromID(cameraID string) *Camera {
	for i := range s.Cameras {
		if s.Cameras[i].ID == cameraID {
			return &s.Cameras[i]
		}
	}
	return &Camera{}
}

func (c *Camera) GetResolutionsForFPS(targetFPS int) []ResolutionOption {
	var options []ResolutionOption
	for _, size := range c.Sizes {
		for _, fpsOpt := range size.FPS {
			if fpsOpt.Value == targetFPS {
				label := size.Resolution
				if fpsOpt.HighSpeed {
					label += " (high-speed)"
				}
				options = append(options, ResolutionOption{
					Value:     size.Resolution,
					Label:     label,
					HighSpeed: fpsOpt.HighSpeed,
				})
				break
			}
		}
	}
	return options
}

func (c *Camera) GetAvailableFPS() []int {
	fpsSet := make(map[int]bool)
	for _, size := range c.Sizes {
		for _, fpsOpt := range size.FPS {
			fpsSet[fpsOpt.Value] = true
		}
	}
	fpsList := make([]int, 0, len(fpsSet))
	for fps := range fpsSet {
		fpsList = append(fpsList, fps)
	}
	slices.Sort(fpsList)
	return fpsList
}
