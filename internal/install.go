package internal

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/walter2310/nvx/pkg"
)

// Installer orchestrates downloading and installing Node.js versions.
type Installer struct {
	InstallDir string
	Downloader Downloader
	Extractor  Extractor
}

// Downloader defines how binaries are downloaded.
type Downloader interface {
	Download(version, platform, extension string) (string, error)
}

// Extractor defines how archives are extracted.
type Extractor interface {
	Extract(src, dest string) error
}

// HTTPDownloader fetches Node.js binaries from the official distribution site.
type HTTPDownloader struct{}

func (d *HTTPDownloader) Download(version, platform, extension string) (string, error) {
	url := fmt.Sprintf("https://nodejs.org/dist/%s/node-%s-%s-x64.%s",
		version, version, platform, extension)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download Node.js: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download Node.js: status %d", resp.StatusCode)
	}

	targetDir := filepath.Join("versions", version)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("error creating directory: %v", err)
	}

	filename := fmt.Sprintf("node-%s-%s-x64.%s", version, platform, extension)
	targetFile := filepath.Join(targetDir, filename)

	out, err := os.Create(targetFile)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer out.Close()

	bar := progressbar.DefaultBytes(resp.ContentLength, "Downloading Node.js")
	if _, err := io.Copy(io.MultiWriter(out, bar), resp.Body); err != nil {
		return "", fmt.Errorf("failed to save Node.js: %v", err)
	}

	return targetFile, nil
}

// ZipExtractor handles .zip archives.
type ZipExtractor struct{}

func (e *ZipExtractor) Extract(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	var files []*zip.File
	for _, f := range r.File {
		if !f.FileInfo().IsDir() {
			files = append(files, f)
		}
	}

	bar := progressbar.Default(int64(len(files)), "Extracting Node.js")

	numWorkers := runtime.NumCPU()
	jobs := make(chan *zip.File, len(files))
	errs := make(chan error, len(files))
	var wg sync.WaitGroup

	for range numWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 32*1024)
			for f := range jobs {
				fPath := filepath.Join(dest, f.Name)

				if err := os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
					errs <- err
					return
				}

				rc, err := f.Open()
				if err != nil {
					errs <- err
					return
				}

				outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err != nil {
					rc.Close()
					errs <- err
					return
				}

				_, err = io.CopyBuffer(outFile, rc, buf)

				outFile.Close()
				rc.Close()

				if err != nil {
					errs <- err
					return
				}
				bar.Add(1)
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	wg.Wait()
	close(errs)

	if len(errs) > 0 {
		return <-errs
	}
	return nil
}

// Install orchestrates the installation of a specific Node.js version.
func (i *Installer) Install(version string) error {
	platform, extension := pkg.IdentifyOS()

	archivePath, err := i.Downloader.Download(version, platform, extension)
	if err != nil {
		return err
	}

	if extension == "zip" {
		if err := i.Extractor.Extract(archivePath, filepath.Dir(archivePath)); err != nil {
			return fmt.Errorf("failed to extract Node.js: %v", err)
		}
	}

	fmt.Println("Node installed successfully")
	return nil
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a specific version of Node.js",
	Long:  `Install a specific version of Node.js and set it as the active version.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeVersion := "v" + args[0]

		if err := pkg.ValidateArgSyntax(args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		i := &Installer{
			InstallDir: "versions",
			Downloader: &HTTPDownloader{},
			Extractor:  &ZipExtractor{},
		}

		if err := i.Install(nodeVersion); err != nil {
			fmt.Printf("Error installing Node.js: %v\n", err)
		}
	},
}
