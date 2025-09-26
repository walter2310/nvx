package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

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

		var files []string
		err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to scan files: %v", err)
		}

		bar := progressbar.NewOptions(len(files),
			progressbar.OptionSetDescription("📦 Setting Node.js files..."),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowCount(),
			progressbar.OptionShowIts(),
			progressbar.OptionSetWidth(10),
			progressbar.OptionThrottle(65*time.Millisecond),
			progressbar.OptionOnCompletion(func() {
				fmt.Fprint(os.Stderr, "\n")
			}),
			progressbar.OptionSpinnerType(14),
			progressbar.OptionFullWidth(),
		)

		destDir := filepath.Join(nvxDir, "current", "bin")
		os.RemoveAll(destDir)
		os.MkdirAll(destDir, 0755)

		for _, filePath := range files {
			relPath, err := filepath.Rel(sourceDir, filePath)
			if err != nil {
				return fmt.Errorf("failed to get relative path: %v", err)
			}

			dstPath := filepath.Join(destDir, relPath)

			dstDir := filepath.Dir(dstPath)
			if err := os.MkdirAll(dstDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", dstDir, err)
			}

			if err := pkg.CopyFile(filePath, dstPath); err != nil {
				return fmt.Errorf("failed to copy %s: %v", relPath, err)
			}

			bar.Add(1)
		}

		bar.Finish()

		binPath := filepath.Join(nvxDir, "current", "bin")
		absBinPath, _ := filepath.Abs(binPath)

		if err := pkg.AddToUserPath(absBinPath); err != nil {
			return fmt.Errorf("failed to update PATH: %v", err)
		}

	} else {
		if err := os.Symlink(versionDir, currentLink); err != nil {
			return fmt.Errorf("failed to switch version: %v", err)
		}
	}

	return nil
}
