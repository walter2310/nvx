package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/walter2310/nvx/pkg"
)

var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Use a specific version of Node.js",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeVersion := "v" + args[0]

		if err := useNodeVersion(nodeVersion); err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Now using Node.js %s\n", nodeVersion)
		}
	},
}

func useNodeVersion(nodeVersion string) error {
	versionDir := filepath.Join("versions", nodeVersion)
	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed", nodeVersion)
	}

	currentLink := filepath.Join("versions", "current")

	os.RemoveAll(currentLink)

	if runtime.GOOS == "windows" {
		// Copiar archivos en Windows
		if err := pkg.CopyDir(versionDir, currentLink); err != nil {
			return fmt.Errorf("failed to copy version: %v", err)
		}
	} else {
		// Crear symlink en Unix/macOS
		if err := os.Symlink(versionDir, currentLink); err != nil {
			return fmt.Errorf("failed to switch version: %v", err)
		}
	}

	return nil
}
