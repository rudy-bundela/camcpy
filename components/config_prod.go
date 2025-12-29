//go:build release

package components

const IsDev = false

// GetBaseScrcpyArgs returns arguments optimized for production
func GetBaseScrcpyArgs() []string {
	return []string{
		"-ra.mp4",
		"--no-playback",
		"--video-source=camera",
		"--no-window",
		"--no-control",
		"--audio-codec=aac",
	}
}
