package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nvx",
	Short: "NVX is a CLI tool for managing Node.js versions",
	Long:  `NVX is a CLI tool for managing your Node.js versions with ease and efficiency.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing nvx %v\n", err)
		os.Exit(1)
	}
}
