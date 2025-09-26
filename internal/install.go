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

		if err := installNodeVersion(nodeVersion); err != nil {
			fmt.Printf("Error installing Node.js: %v\n", err)
		}
	},
}

func installNodeVersion(nodeVersion string) error {
	platform, extension := pkg.IdentifyOS()

	nodeInstallerUrl := "https://nodejs.org/dist/" + nodeVersion +
		"/node-" + nodeVersion + "-" + platform + "-x64." + extension

	resp, err := http.Get(nodeInstallerUrl)
	if err != nil {
		return fmt.Errorf("failed to download Node.js: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download Node.js: status %d", resp.StatusCode)
	}

	targetFile, err := DownloadNodeVersionBinary(nodeVersion, nodeVersion, resp)

	if err != nil {
		return err
	}

	if extension == "zip" {
		if err := unzip(targetFile, filepath.Dir(targetFile)); err != nil {
			return fmt.Errorf("failed to unzip Node.js: %v", err)
		}
	}

	fmt.Printf("Node installed successfully ")

	return nil
}

func DownloadNodeVersionBinary(dirname string, nodeVersion string, resp *http.Response) (string, error) {
	targetDir := filepath.Join("versions", dirname)
	platform, extension := pkg.IdentifyOS()

	err := os.MkdirAll(targetDir, 0755) // 0755 grants read/write/execute for owner, read/execute for group/others
	if err != nil {
		return "", fmt.Errorf("error creating directory: %v", err)
	}

	filename := "node-" + nodeVersion + "-" + platform + "-x64." + extension
	targetFile := filepath.Join(targetDir, filename)

	out, err := os.Create(targetFile)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}

	defer out.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"Downloading Node.js",
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save Node.js: %v", err)
	}

	return targetFile, nil
}

func unzip(src string, dest string) error {
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

	// Pool workers
	numWorkers := runtime.NumCPU()
	jobs := make(chan *zip.File, len(files))
	errs := make(chan error, len(files))

	var wg sync.WaitGroup

	// Lanzamos los workers
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := make([]byte, 32*1024) // buffer reutilizable por worker
			for f := range jobs {
				fPath := filepath.Join(dest, f.Name)

				// Crear directorios si hacen falta
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
