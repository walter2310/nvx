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

type Usage struct {
	Finder Finder
}

type Finder interface {
	Find(dir, version string) (string, error)
}

type WindowsFinder struct{}

func (f *WindowsFinder) Find(dir, version string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read version directory: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			candidate := filepath.Join(dir, e.Name())
			if _, err := os.Stat(filepath.Join(candidate, "node.exe")); err == nil {
				return candidate, nil
			}
		} else if e.Name() == "node.exe" {
			return dir, nil
		}
	}

	return "", fmt.Errorf("could not find node.exe in version %s", version)
}

func (u *Usage) UseNodeVersion(version string) error {
	nvxDir, err := pkg.GetVersionsDir()
	if err != nil {
		return fmt.Errorf("get versions dir: %w", err)
	}

	versionDir := filepath.Join(nvxDir, version)
	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		return fmt.Errorf("version %s is not installed", version)
	}

	currentLink := filepath.Join(nvxDir, "current")
	if err := os.RemoveAll(currentLink); err != nil {
		return fmt.Errorf("cleanup current link: %w", err)
	}

	if runtime.GOOS == "windows" {
		return u.switchWindows(versionDir, nvxDir, version)
	}

	return u.switchUnix(versionDir, currentLink)
}

func (u *Usage) switchWindows(versionDir, nvxDir, version string) error {
	sourceDir, err := u.Finder.Find(versionDir, version)
	if err != nil {
		return err
	}

	files, err := collectFiles(sourceDir)
	if err != nil {
		return fmt.Errorf("collect files: %w", err)
	}

	destDir := filepath.Join(nvxDir, "current", "bin")
	if err := os.RemoveAll(destDir); err != nil {
		return fmt.Errorf("remove dest dir: %w", err)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	bar := newProgressBar(len(files))
	if err := copyFiles(files, sourceDir, destDir, bar); err != nil {
		return err
	}
	bar.Finish()

	absBinPath, _ := filepath.Abs(destDir)
	if err := pkg.AddToUserPath(absBinPath); err != nil {
		return fmt.Errorf("update PATH: %w", err)
	}

	return nil
}

func (u *Usage) switchUnix(versionDir, currentLink string) error {
	if err := os.Symlink(versionDir, currentLink); err != nil {
		return fmt.Errorf("symlink %s -> %s: %w", currentLink, versionDir, err)
	}

	return nil
}

func collectFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func copyFiles(files []string, src, dst string, bar *progressbar.ProgressBar) error {
	for _, filePath := range files {
		relPath, err := filepath.Rel(src, filePath)
		if err != nil {
			return fmt.Errorf("get relative path: %w", err)
		}
		dstPath := filepath.Join(dst, relPath)
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return fmt.Errorf("create dir: %w", err)
		}
		if err := pkg.CopyFile(filePath, dstPath); err != nil {
			return fmt.Errorf("copy %s: %w", relPath, err)
		}
		bar.Add(1)
	}
	return nil
}

func newProgressBar(total int) *progressbar.ProgressBar {
	return progressbar.NewOptions(total,
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
}

var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Use a specific version of Node.js",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version := "v" + args[0]
		u := &Usage{Finder: &WindowsFinder{}}

		if err := u.UseNodeVersion(version); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		fmt.Printf("Now using Node.js %s\n", version)
	},
}
