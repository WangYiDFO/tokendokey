package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var LogoutCmd = &cobra.Command{
	Use:   "logout -c=[client_name]",
	Short: "Logout from the specified client by simply remove access and refresh tokens.",
	Run: func(cmd *cobra.Command, args []string) {
		clientName, _ := cmd.Flags().GetString("client")

		configDir := filepath.Join(getHomeDir(), ".tokendokey", clientName)
		refreshTokenPath := filepath.Join(configDir, "refresh_token.txt")
		accessTokenPath := filepath.Join(configDir, "access_token.txt")

		err := os.Remove(refreshTokenPath)
		if err != nil {
			fmt.Println("Error removing refresh token:", err)
		}

		err = os.Remove(accessTokenPath)
		if err != nil {
			fmt.Println("Error removing access token:", err)
		}

		fmt.Println("Logged out successfully.")
	},
}

func init() {
	LogoutCmd.Flags().StringP("client", "c", "", "Client name for the OAuth configuration")
	LogoutCmd.MarkFlagRequired("client")
}
