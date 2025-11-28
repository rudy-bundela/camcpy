// Package db contains the in-memory struct
package db

import (
	"bufio"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type ScrcpyInfo struct {
	DeviceName string   `json:"device_name"`
	Cameras    []Camera `json:"cameras"`
}

type FPSOption struct {
	Value     int  `json:"value"`
	HighSpeed bool `json:"high_speed"`
}

type SizeConfig struct {
	Resolution string      `json:"resolution"`
	Width      int         `json:"width"`
	Height     int         `json:"height"`
	FPS        []FPSOption `json:"fps"`
}

type Camera struct {
	ID       string       `json:"id"`
	Position string       `json:"position"`
	Sizes    []SizeConfig `json:"sizes"`
}

func (s *ScrcpyInfo) ParseScrcpyOutput(input string) (*ScrcpyInfo, error) {
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

	// Key = "1920x1080", Value = Map of FPS (int -> isHighSpeed bool)
	tempSizeMap := make(map[string]map[int]bool)
	var currentBaseFPS []int
	var inHighSpeedSection bool

	finalizeCamera := func() {
		if currentCamera != nil {
			// Convert map to slice
			for resStr, fpsMap := range tempSizeMap {
				// Parse width/height from key for sorting
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

				// Sort FPS options by value (ascending)
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

			// Sort sizes by Width descending (typical UI preference)
			sort.Slice(currentCamera.Sizes, func(i, j int) bool {
				return currentCamera.Sizes[i].Width > currentCamera.Sizes[j].Width
			})

			info.Cameras = append(info.Cameras, *currentCamera)
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		// 1. Check for Device Name
		if matches := reDeviceName.FindStringSubmatch(line); matches != nil {
			info.DeviceName = strings.TrimSpace(matches[1])
			continue
		}

		// 2. Check for Camera Header
		if matches := reCameraHeader.FindStringSubmatch(line); matches != nil {
			finalizeCamera() // Save previous camera if exists

			// Reset for new camera
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

		// 3. Check for High Speed Header
		if reHighSpeed.MatchString(line) {
			inHighSpeedSection = true
			continue
		}

		// 4. Check for Size Line
		if currentCamera != nil {
			if matches := reSizeLine.FindStringSubmatch(line); matches != nil {
				wStr, hStr := matches[1], matches[2]
				fpsGroup := matches[3] // Might be empty
				resolution := fmt.Sprintf("%sx%s", wStr, hStr)

				// Determine which FPS to apply
				var fpsToAdd []int
				if fpsGroup != "" {
					// This line has specific FPS (often High Speed section)
					fpsToAdd = parseFPSList(fpsGroup)
				} else {
					// This line uses the camera's base FPS (Standard section)
					fpsToAdd = currentBaseFPS
				}

				// Merge into map
				if _, exists := tempSizeMap[resolution]; !exists {
					tempSizeMap[resolution] = make(map[int]bool)
				}

				for _, f := range fpsToAdd {
					// If the FPS entry already exists as 'false' (Standard), don't overwrite it with 'true' (HighSpeed).
					// We prefer Standard mode if available for the same FPS.
					if existingIsHighSpeed, ok := tempSizeMap[resolution][f]; ok && !existingIsHighSpeed {
						continue
					}
					tempSizeMap[resolution][f] = inHighSpeedSection
				}
			}
		}
	}

	finalizeCamera() // Save the last camera

	return info, nil
}

// Helper to parse "15, 24, 30" into []int
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
