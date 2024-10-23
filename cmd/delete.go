package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var DeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the config folder for a client",
	Run: func(cmd *cobra.Command, args []string) {
		clientName, _ := cmd.Flags().GetString("clientname")
		if clientName == "" {
			fmt.Println("Error: --clientname parameter is required")
			return
		}

		configPath := filepath.Join(getHomeDir(), ".tokendokey", clientName)

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Println("Error: Config folder does not exist:", configPath)
			return
		}

		err := os.RemoveAll(configPath)
		if err != nil {
			fmt.Println("Error: Unable to delete config folder:", err)
			return
		}

		fmt.Println("Config folder deleted successfully for client:", clientName)
	},
}

func init() {
	DeleteCmd.Flags().StringP("clientname", "c", "", "Client name")
	DeleteCmd.MarkFlagRequired("clientname")
}
