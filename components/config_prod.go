//go:build release

package main

// GetBaseScrcpyArgs returns arguments optimized for production
func GetBaseScrcpyArgs() []string {
	return []string{
		"-ra.mp4"
		"--no-playback"
		"--no-window",
		"--no-control",
		"--no-window",
		"--audio-codec=aac",
	}
}
