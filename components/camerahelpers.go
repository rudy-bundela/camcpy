package components

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"

	"github.com/starfederation/datastar-go/datastar"
)

func readSignals(r *http.Request) (*DatastarSignalsStruct, error) {
	signals := &DatastarSignalsStruct{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		return nil, fmt.Errorf("datastar error reading signals: %w", err)
	}
	return signals, nil
}

func newSSE(w http.ResponseWriter, r *http.Request) *datastar.ServerSentEventGenerator {
	return datastar.NewSSE(
		w,
		r,
		datastar.WithCompression(
			datastar.WithBrotli(
				datastar.WithBrotliLGWin(0),
			),
		),
	)
}

func (s *ScrcpyInfo) HandleGetCameraOptions(w http.ResponseWriter, r *http.Request) {
	// TODO: fix this nonsense
	sse := newSSE(w, r)

	signals, err := readSignals(r)
	if err != nil {
		fmt.Println("Datastar error reading signals in HandleGetCameraOptions:", err)
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
		log.Println("Error console logging")
	}
}

func SetCameraFPS(sse *datastar.ServerSentEventGenerator, signals *DatastarSignalsStruct, s *ScrcpyInfo) {
	fpslist := make([]int, 0)
	fpslist = append(fpslist, s.GetCameraFromID(signals.CamID).GetAvailableFPS()...)
	if err := sse.PatchElementTempl(CameraFPSComponent(fpslist)); err != nil {
		fmt.Println("Error in patching element for CameraIDComponent", err)
	}

	if !slices.Contains(fpslist, signals.Fps) {
		signals.Fps = fpslist[len(fpslist)-1]
	}
}

func SetCameraResolution(sse *datastar.ServerSentEventGenerator, signals *DatastarSignalsStruct, s *ScrcpyInfo) {
	resolutions := make([]ResolutionOption, 0)
	resolutions = append(resolutions, s.GetCameraFromID(signals.CamID).GetResolutionsForFPS(signals.Fps)...)
	if err := sse.PatchElementTempl(CameraResolutionComponent(resolutions)); err != nil {
		fmt.Println("Error in patching element for CameraIDComponent", err)
	}

	if slices.ContainsFunc(resolutions, func(r ResolutionOption) bool {
		return strings.Contains(r.Label, "1920x1080 (high-speed)")
	}) {
		signals.Resolution = "1920x1080 (high-speed)"
	} else {
		signals.Resolution = resolutions[0].Label
	}
}

func SetCameraID(sse *datastar.ServerSentEventGenerator, signals *DatastarSignalsStruct, s *ScrcpyInfo) {
	newCamera := make([]Camera, 0)
	newCamera = append(newCamera, s.GetCameraFromPosition(signals.Position)...)
	if err := sse.PatchElementTempl(CameraIDComponent(newCamera)); err != nil {
		fmt.Println("Error in patching element for CameraIDComponent", err)
	}

	signals.CamID = newCamera[0].ID
}

func (s *ScrcpyInfo) HandleCameraIDUpdate(w http.ResponseWriter, r *http.Request) {
	signals, err := readSignals(r)
	if err != nil {
		fmt.Println("Datastar error reading signals in HandleCameraUpdate:", err)
	}

	sse := newSSE(w, r)

	SetCameraID(sse, signals, s)
	SetCameraFPS(sse, signals, s)
	SetCameraResolution(sse, signals, s)

	if err := sse.MarshalAndPatchSignals(signals); err != nil {
		fmt.Println("Error marshalling and patching signals in HandleCameraUpdate", err)
	}
}

func (s *ScrcpyInfo) HandleCameraFPSUpdate(w http.ResponseWriter, r *http.Request) {
	signals, err := readSignals(r)
	if err != nil {
		fmt.Println("Datastar error reading signals in HandleCameraUpdate:", err)
	}

	sse := newSSE(w, r)

	SetCameraFPS(sse, signals, s)
	SetCameraResolution(sse, signals, s)

	if err := sse.MarshalAndPatchSignals(signals); err != nil {
		fmt.Println("Error marshalling and patching signals in HandleCameraUpdate", err)
	}
}

func (s *ScrcpyInfo) HandleCameraResolutionUpdate(w http.ResponseWriter, r *http.Request) {
	signals, err := readSignals(r)
	if err != nil {
		fmt.Println("Datastar error reading signals in HandleCameraUpdate:", err)
	}

	sse := newSSE(w, r)

	SetCameraResolution(sse, signals, s)

	if err := sse.MarshalAndPatchSignals(signals); err != nil {
		fmt.Println("Error marshalling and patching signals in HandleCameraUpdate", err)
	}
}

func (s *ScrcpyInfo) PrintStruct(w http.ResponseWriter, r *http.Request) {
	jsonoutput, _ := json.Marshal(s)
	fmt.Println(string(jsonoutput))
	if _, err := w.Write(jsonoutput); err != nil {
		fmt.Println("Error writing jsonoutput")
	}
}

func (s *ScrcpyInfo) GetCameraFromPosition(position string) []Camera {
	cameraList := make([]Camera, 0)
	for _, camera := range s.Cameras {
		if camera.Position == position {
			cameraList = append(cameraList, camera)
		}
	}
	return cameraList
}

func (s *ScrcpyInfo) GetCameraFromID(cameraID string) *Camera {
	camera := &Camera{}
	for _, cameras := range s.Cameras {
		if cameras.ID == cameraID {
			camera = &cameras
		}
	}
	return camera
}

// GetResolutionsForFPS returns all resolutions that support the specific frame rate.
// It also returns the specific FPSOption configuration (e.g. to check HighSpeed requirements).
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

// GetAvailableFPS returns a sorted list of all unique FPS values supported by this camera.
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

func (s *ScrcpyInfo) HandleStartStream(w http.ResponseWriter, r *http.Request) {
	signals, err := readSignals(r)
	if err != nil {
		fmt.Println("Datastar error reading signals in HandleStartStream:", err)
	}

	log.Println(signals)
}
