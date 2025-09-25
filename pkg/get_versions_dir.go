package pkg

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetVersionsDir() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %v", err)
	}

	versionsDir := filepath.Join(currentDir, "versions")
	if _, err := os.Stat(versionsDir); os.IsNotExist(err) {
		return "", fmt.Errorf("versions directory does not exist")
	}

	return versionsDir, nil
}
