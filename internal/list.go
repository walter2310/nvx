package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/walter2310/nvx/pkg"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all version of Node.js",
	Run: func(cmd *cobra.Command, args []string) {
		listNodeVersions()
	},
}

func listNodeVersions() error {
	versionsDir, err := pkg.GetVersionsDir()
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	files, err := os.ReadDir(versionsDir)
	if err != nil {
		return fmt.Errorf("failed to read versions directory: %v", err)
	}

	fmt.Println("Installed Node.js versions:")
	for _, file := range files {
		if file.IsDir() {
			if file.Name() == "current" {
				continue
			}

			fmt.Println("- " + file.Name())
		}
	}

	return nil
}
