package pkg

import (
	"runtime"
)

func IdentifyOS() (string, string) {
	os := runtime.GOOS
	var platform, extension string

	switch os {
	case "windows":
		platform = "win"
		extension = "zip"
	case "darwin":
		platform = "darwin"
		extension = "tar.xz"
	case "linux":
		platform = "linux"
		extension = "tar.xz"
	default:
		return "", ""
	}

	return platform, extension
}
