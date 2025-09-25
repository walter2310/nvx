package internal

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/walter2310/nvx/pkg"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall a specific version of Node.js",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		nodeVersion := "v" + args[0]

		if err := pkg.ValidateArgSyntax(args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if err := uninstallNodeVersion(nodeVersion); err != nil {
			fmt.Printf("Error uninstalling Node.js: %v\n", err)
		}
	},
}

func uninstallNodeVersion(nodeVersion string) error {
	versionsDir, err := pkg.GetVersionsDir()
	if err != nil {
		return fmt.Errorf("error: %v", err)
	}

	targetDir := fmt.Sprintf("%s/%s", versionsDir, nodeVersion)

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return fmt.Errorf("Node.js version %s is not installed", nodeVersion)
	}

	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("failed to uninstall Node.js version %s: %v", nodeVersion, err)
	}

	fmt.Printf("Node.js version %s uninstalled successfully\n", nodeVersion)

	return nil
}
