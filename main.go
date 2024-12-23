package main

import (
	"tokendokey/cmd"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{Use: "tokendokey"}

	rootCmd.AddCommand(cmd.InitCmd)
	rootCmd.AddCommand(cmd.GetTokenCmd)
	rootCmd.AddCommand(cmd.LoginCmd)
	rootCmd.AddCommand(cmd.LogoutCmd)
	rootCmd.AddCommand(cmd.ExportCmd)
	rootCmd.AddCommand(cmd.ImportCmd)
	rootCmd.AddCommand(cmd.ListCmd)
	rootCmd.AddCommand(cmd.DeleteCmd)
	rootCmd.AddCommand(cmd.MTLSTokenCmd)

	rootCmd.Execute()
}
