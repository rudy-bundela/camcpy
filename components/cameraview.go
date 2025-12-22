package components

import (
	"fmt"
	"slices"
)

type CameraViewModel struct {
	DeviceName       string
	Cameras          []CameraOption
	SelectedCameraID string

	FPSOptions  []int
	SelectedFPS int

	Resolutions        []ResolutionOption
	SelectedResolution string
}

type CameraOption struct {
	ID    string
	Label string
}

// NewCameraViewModel constructs the view model using helper methods on the Camera struct.
func NewCameraViewModel(info *ScrcpyInfo, selectedCamID string, selectedFPS int, selectedRes string) CameraViewModel {
	vm := CameraViewModel{
		DeviceName:         info.DeviceName,
		SelectedCameraID:   selectedCamID,
		SelectedFPS:        selectedFPS,
		SelectedResolution: selectedRes,
	}

	// 1. Populate Camera Options and find the active Camera object
	var selectedCamera *Camera

	// If no cameras found, return empty VM
	if len(info.Cameras) == 0 {
		return vm
	}

	for _, cam := range info.Cameras {
		label := fmt.Sprintf("ID %s (%s)", cam.ID, cam.Position)
		vm.Cameras = append(vm.Cameras, CameraOption{ID: cam.ID, Label: label})

		// Check if this is the selected camera
		if cam.ID == selectedCamID {
			// Create a copy or pointer to the loop variable carefully
			c := cam
			selectedCamera = &c
		}
	}

	// Fallback: If selected ID not found (or empty), default to the first camera
	if selectedCamera == nil {
		selectedCamera = &info.Cameras[0]
		vm.SelectedCameraID = selectedCamera.ID
	}

	// 2. Populate FPS Options using the Helper (defined in scrcpy_helpers.go)
	vm.FPSOptions = selectedCamera.GetAvailableFPS()

	// Validate/Default the SelectedFPS
	fpsIsValid := slices.Contains(vm.FPSOptions, vm.SelectedFPS)

	if !fpsIsValid && len(vm.FPSOptions) > 0 {
		// Heuristic: Prefer 60, then 30, then the highest available
		has60 := false
		has30 := false
		maxFPS := 0

		for _, f := range vm.FPSOptions {
			if f == 60 {
				has60 = true
			}
			if f == 30 {
				has30 = true
			}
			if f > maxFPS {
				maxFPS = f
			}
		}

		if has60 {
			vm.SelectedFPS = 60
		} else if has30 {
			vm.SelectedFPS = 30
		} else {
			vm.SelectedFPS = maxFPS
		}
	}

	// 3. Populate Resolution Options using the Helper (defined in scrcpy_helpers.go)
	// This now uses the simplified logic that checks for specific FPS support
	vm.Resolutions = selectedCamera.GetResolutionsForFPS(vm.SelectedFPS)

	// Validate/Default the SelectedResolution
	resIsValid := false
	for _, r := range vm.Resolutions {
		if r.Value == vm.SelectedResolution {
			resIsValid = true
			break
		}
	}

	// If the previously selected resolution isn't available for this new FPS/Camera,
	// default to the first one in the list (usually the highest res due to sort order).
	if !resIsValid && len(vm.Resolutions) > 0 {
		vm.SelectedResolution = vm.Resolutions[0].Value
	}

	return vm
}
