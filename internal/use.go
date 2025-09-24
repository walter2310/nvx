package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/schollz/progressbar/v3"
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
	homeDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get user home: %v", err)
	}

	nvxDir := filepath.Join(homeDir, "versions")
	versionDir := filepath.Join(nvxDir, nodeVersion)

	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed", nodeVersion)
	}

	currentLink := filepath.Join(nvxDir, "current")
	os.RemoveAll(currentLink)

	if runtime.GOOS == "windows" {
		// Buscar automáticamente dónde está node.exe
		entries, err := os.ReadDir(versionDir)
		if err != nil {
			return fmt.Errorf("failed to read version directory: %v", err)
		}

		var sourceDir string
		for _, e := range entries {
			if e.IsDir() {
				candidate := filepath.Join(versionDir, e.Name())
				if _, err := os.Stat(filepath.Join(candidate, "node.exe")); err == nil {
					sourceDir = candidate
					break
				}
			} else if e.Name() == "node.exe" {
				// node.exe está directamente en versionDir
				sourceDir = versionDir
				break
			}
		}

		if sourceDir == "" {
			return fmt.Errorf("could not find node.exe in version %s", nodeVersion)
		}

		// Contar archivos para el progress bar
		var totalFiles int64
		filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				totalFiles++
			}
			return nil
		})

		// Crear progress bar para la copia de archivos
		bar := progressbar.DefaultBytes(
			totalFiles,
			"📦 Setting Node.js files...",
		)

		// Crear destino en current/bin
		destDir := filepath.Join(nvxDir, "current", "bin")
		os.RemoveAll(destDir)
		os.MkdirAll(destDir, 0755)

		// Copiar todos los archivos de sourceDir a destDir con progress bar
		err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			relPath, _ := filepath.Rel(sourceDir, path)
			dstPath := filepath.Join(destDir, relPath)

			dstDir := filepath.Dir(dstPath)
			os.MkdirAll(dstDir, 0755)

			if err := pkg.CopyFile(path, dstPath); err != nil {
				return fmt.Errorf("failed to copy %s: %v", relPath, err)
			}

			// Actualizar progress bar después de cada archivo copiado
			bar.Add64(1)
			return nil
		})
		if err != nil {
			return err
		}

		binPath := filepath.Join(nvxDir, "current", "bin")
		absBinPath, _ := filepath.Abs(binPath)

		if err := pkg.AddToUserPath(absBinPath); err != nil {
			return fmt.Errorf("failed to update PATH: %v", err)
		}

	} else {
		// En Unix/macOS mantenemos symlink
		if err := os.Symlink(versionDir, currentLink); err != nil {
			return fmt.Errorf("failed to switch version: %v", err)
		}
	}

	return nil
}
