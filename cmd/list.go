package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all clients or display settings for a specific client",
	Long:  `List all clients for the current user. Or display settings for a specific client.`,
	Example: `  tokendokey list
  tokendokey list -c=myclient`,
	Run: func(cmd *cobra.Command, args []string) {
		clientName, _ := cmd.Flags().GetString("client")
		if clientName == "" {
			listAllClients()
		} else {
			displayClientSettings(clientName)
		}
	},
}

func init() {
	ListCmd.Flags().StringP("client", "c", "", "Specify the client name")
}

func listAllClients() {
	configDir := filepath.Join(getHomeDir(), ".tokendokey")
	// dir := ".tokendokey"
	files, err := os.ReadDir(configDir)
	if err != nil {
		fmt.Println("Error reading .tokendokey directory:", err)
		return
	}

	fmt.Println("Current user have the following client settings in .tokendokey directory:")
	for _, file := range files {
		if file.IsDir() {
			fmt.Println(file.Name())
		}
	}
}

func displayClientSettings(clientName string) {
	configPath := filepath.Join(getHomeDir(), ".tokendokey", clientName, "config.json")
	// configPath := filepath.Join(".tokendokey", clientName, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("Error reading config.json for client", clientName, ":", err)
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Println("Error parsing config.json for client", clientName, ":", err)
		return
	}

	if clientSecret, ok := config["client_secret"].(string); ok && clientSecret != "" {
		config["client_secret"] = maskString(clientSecret)
	}

	configJSON, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("Error formatting config.json for client", clientName, ":", err)
		return
	}

	fmt.Println("Current user has the following settings for client:", clientName)
	fmt.Println(string(configJSON))
}

func maskString(s string) string {
	if len(s) <= 4 {
		return strings.Repeat("*", len(s))
	}
	return s[:1] + strings.Repeat("*", len(s)-2) + s[len(s)-1:]
}
