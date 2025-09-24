package cmd

func RegisterCommands() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(useCmd)
}
