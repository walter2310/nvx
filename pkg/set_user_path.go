package pkg

import (
	"fmt"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

func AddToUserPath(binPath string) error {
	binPathClean, err := filepath.Abs(binPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %v", err)
	}
	binPathClean = filepath.Clean(binPathClean)

	// Abrir la clave del registro de Windows donde se guarda el PATH del usuario
	key, _, err := registry.CreateKey(registry.CURRENT_USER,
		`Environment`,
		registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry: %v", err)
	}
	defer key.Close()

	existing, _, err := key.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return fmt.Errorf("failed to read PATH: %v", err)
	}

	parts := strings.Split(existing, ";")
	cleanParts := []string{}

	for _, p := range parts {
		pTrim := strings.TrimSpace(p)
		if pTrim == "" {
			continue
		}

		pAbs, _ := filepath.Abs(pTrim)
		pAbs = filepath.Clean(pAbs)
		cleanParts = append(cleanParts, pTrim)
	}

	newPath := binPathClean
	if len(cleanParts) > 0 {
		newPath += ";" + strings.Join(cleanParts, ";")
	}

	if err := key.SetStringValue("Path", newPath); err != nil {
		return fmt.Errorf("failed to update PATH: %v", err)
	}

	sendEnvChange()
	return nil
}

func sendEnvChange() {
	const HWND_BROADCAST = 0xffff
	const WM_SETTINGCHANGE = 0x001A

	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessage := user32.NewProc("SendMessageW")

	sendMessage.Call(
		uintptr(HWND_BROADCAST),
		uintptr(WM_SETTINGCHANGE),
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("Environment"))),
	)
}
