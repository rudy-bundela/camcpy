package components

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
)

func (s *ScrcpyInfo) HandleGetCameraOptions(w http.ResponseWriter, r *http.Request) {
	// TODO: fix this nonsense
	sse := datastar.NewSSE(w, r, datastar.WithCompression(datastar.WithBrotli(datastar.WithBrotliLGWin(0))))

	signals := &DatastarSignalsStruct{}
	if err := datastar.ReadSignals(r, signals); err != nil {
		fmt.Println("Datastar error reading signals in HandleGetCameraOptions: ", err)
	}

	if s.DeviceName != "" {
		return
	}

	scrcpyOutput, err := RunGetScrcpyDetails()
	if err != nil {
		runOnScrcpyError(sse, err)
	}

	if err := s.ParseScrcpyOutput(string(scrcpyOutput)); err != nil {
		log.Println("Error parsing scrcpy output: ", err)
	}

	if err := sse.PatchElementTempl(Layout(CameraComponent(s, signals))); err != nil {
		log.Println("Error console logging")
	}
}

func (s *ScrcpyInfo) HandleCameraUpdate(w http.ResponseWriter, r *http.Request) {
	sse := datastar.NewSSE(w, r, datastar.WithCompression(datastar.WithBrotli(datastar.WithBrotliLGWin(0))))

	signals := &DatastarSignalsStruct{}

	if err := datastar.ReadSignals(r, signals); err != nil {
		fmt.Println("Datastar error reading signals in HandleCameraUpdate: ", err)
	}

	fmt.Println("Signals in HandleCameraUpdate = ", signals)

	if err := sse.PatchElementTempl(Layout(CameraComponent(s, signals))); err != nil {
		log.Println("Error console logging")
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

				// Add the lightning bolt if this specific FPS/Res combo needs high speed
				if fpsOpt.HighSpeed {
					label += " (high-speed)"
				}

				options = append(options, ResolutionOption{
					Value:     size.Resolution,
					Label:     label,
					HighSpeed: fpsOpt.HighSpeed,
				})
				// We found the match for this resolution, stop checking other FPSs for this specific size
				break
			}
		}
	}
	return options
}

// GetAvailableFPS returns a sorted list of all unique FPS values supported by this camera.
func (c *Camera) GetAvailableFPS() []int {
	uniqueFPS := make(map[int]bool)
	for _, size := range c.Sizes {
		for _, fpsOpt := range size.FPS {
			uniqueFPS[fpsOpt.Value] = true
		}
	}

	var sortedFPS []int
	for fps := range uniqueFPS {
		sortedFPS = append(sortedFPS, fps)
	}
	// Sort ascending
	for i := 0; i < len(sortedFPS); i++ {
		for j := i + 1; j < len(sortedFPS); j++ {
			if sortedFPS[i] > sortedFPS[j] {
				sortedFPS[i], sortedFPS[j] = sortedFPS[j], sortedFPS[i]
			}
		}
	}
	return sortedFPS
}
