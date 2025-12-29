//go:build !release

package components

const IsDev = true

func GetBaseScrcpyArgs() []string {
	return []string{
		"--video-source=camera",
	}
}
