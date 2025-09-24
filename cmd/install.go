package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

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
		fmt.Println("Node.js unzipped successfully")
	}

	fmt.Printf("Downloading Node.js from %s\n", nodeInstallerUrl)

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

	_, err = io.Copy(out, resp.Body)
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

	for _, f := range r.File {
		fPath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fPath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
