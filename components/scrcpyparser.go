// Package components contains the in-memory struct
package components

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/starfederation/datastar-go/datastar"
)

type ScrcpyInfo struct {
	DeviceName     string             `json:"device_name"`
	Cameras        []Camera           `json:"cameras"`
	cancelStream   context.CancelFunc `json:"-"`
	ActiveSettings DatastarSignalsStruct
}

type Camera struct {
	ID       string       `json:"id"`
	Position string       `json:"position"`
	Sizes    []SizeConfig `json:"sizes"`
}

type SizeConfig struct {
	Resolution string      `json:"resolution"`
	Width      int         `json:"width"`
	Height     int         `json:"height"`
	FPS        []FPSOption `json:"fps"`
}

type FPSOption struct {
	Value     int  `json:"value"`
	HighSpeed bool `json:"high_speed"`
}

type ResolutionOption struct {
	Value     string
	Label     string
	HighSpeed bool
}

type DatastarSignalsStruct struct {
	Position   string `json:"position"`
	Fps        int    `json:"fps,string"`
	Resolution string `json:"resolution"`
	CamID      string `json:"camid"`
}

func RunGetScrcpyDetails() (output []byte, err error) {
	cmd := exec.Command("scrcpy", "--list-camera-sizes")
	output, err = cmd.Output()
	return output, err
}

func (s *ScrcpyInfo) ParseScrcpyOutput(input string) error {
	scanner := bufio.NewScanner(strings.NewReader(input))

	// Regex patterns

	reDeviceName := regexp.MustCompile(`\[server\] INFO: Device:\s+(.*)`)

	// Matches: --camera-id=0    (back, 4080x3060, fps=[15, 24, 30, 60])
	reCameraHeader := regexp.MustCompile(`--camera-id=(\d+)\s+\(([^,]+),.*fps=\[(.*?)\]`)

	// Matches: - 1920x1080
	// OR:      - 1920x1080 (fps=[120, 240])
	reSizeLine := regexp.MustCompile(`^\s*-\s+(\d+)x(\d+)(?:\s+\(fps=\[(.*?)\]\))?`)

	reHighSpeed := regexp.MustCompile(`High speed capture`)

	info := s

	var currentCamera *Camera

	tempSizeMap := make(map[string]map[int]bool)
	var currentBaseFPS []int
	var inHighSpeedSection bool

	finalizeCamera := func() {
		if currentCamera != nil {
			for resStr, fpsMap := range tempSizeMap {
				parts := strings.Split(resStr, "x")
				w, _ := strconv.Atoi(parts[0])
				h, _ := strconv.Atoi(parts[1])

				fpsOptions := make([]FPSOption, 0, len(fpsMap))
				for fpsVal, isHighSpeed := range fpsMap {
					fpsOptions = append(fpsOptions, FPSOption{
						Value:     fpsVal,
						HighSpeed: isHighSpeed,
					})
				}

				sort.Slice(fpsOptions, func(i, j int) bool {
					return fpsOptions[i].Value < fpsOptions[j].Value
				})

				currentCamera.Sizes = append(currentCamera.Sizes, SizeConfig{
					Resolution: resStr,
					Width:      w,
					Height:     h,
					FPS:        fpsOptions,
				})
			}

			sort.Slice(currentCamera.Sizes, func(i, j int) bool {
				return currentCamera.Sizes[i].Width > currentCamera.Sizes[j].Width
			})

			info.Cameras = append(info.Cameras, *currentCamera)
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		if matches := reDeviceName.FindStringSubmatch(line); matches != nil {
			info.DeviceName = strings.TrimSpace(matches[1])
			continue
		}

		if matches := reCameraHeader.FindStringSubmatch(line); matches != nil {
			finalizeCamera()

			tempSizeMap = make(map[string]map[int]bool)
			currentBaseFPS = parseFPSList(matches[3])
			inHighSpeedSection = false

			currentCamera = &Camera{
				ID:       matches[1],
				Position: matches[2],
				Sizes:    []SizeConfig{},
			}
			continue
		}

		if reHighSpeed.MatchString(line) {
			inHighSpeedSection = true
			continue
		}

		if currentCamera != nil {
			if matches := reSizeLine.FindStringSubmatch(line); matches != nil {
				wStr, hStr := matches[1], matches[2]
				fpsGroup := matches[3] // Might be empty
				resolution := fmt.Sprintf("%sx%s", wStr, hStr)

				var fpsToAdd []int
				if fpsGroup != "" {
					fpsToAdd = parseFPSList(fpsGroup)
				} else {
					fpsToAdd = currentBaseFPS
				}

				if _, exists := tempSizeMap[resolution]; !exists {
					tempSizeMap[resolution] = make(map[int]bool)
				}

				for _, f := range fpsToAdd {
					if existingIsHighSpeed, ok := tempSizeMap[resolution][f]; ok && !existingIsHighSpeed {
						continue
					}
					tempSizeMap[resolution][f] = inHighSpeedSection
				}
			}
		}
	}

	finalizeCamera()

	return nil
}

func parseFPSList(input string) []int {
	parts := strings.Split(input, ",")
	var result []int
	for _, p := range parts {
		clean := strings.TrimSpace(p)
		if val, err := strconv.Atoi(clean); err == nil {
			result = append(result, val)
		}
	}
	return result
}

func runOnScrcpyError(sse *datastar.ServerSentEventGenerator, err error) {
	log.Println("Error getting information from --list-camera-sizes: ", err)
	for i := range 3 {
		if err := sse.PatchElementTempl(Layout(CodePen(
			[]string{fmt.Sprintf("Error getting information from scrcpy, redirecting to pairing page in %d...", 3-i)}))); err != nil {
			log.Println("Error in patching codepen with timer: ", err)
		}
		time.Sleep(1 * time.Second)
	}
	if err := sse.Redirect("/pair"); err != nil {
		log.Println("Error in sse redirect when redirecting from setupcamera")
	}
}
