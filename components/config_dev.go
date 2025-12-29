//go:build !release

package components

func GetBaseScrcpyArgs() []string {
	return []string{
		"--video-source=camera",
	}
}
