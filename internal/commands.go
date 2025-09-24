package internal

func RegisterCommands() {
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(useCmd)
}
