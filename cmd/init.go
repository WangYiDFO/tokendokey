package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type Config struct {
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	TokenIssueURL string `json:"token_issue_url"`
	DeviceCodeURL string `json:"device_authorization_endpoint"`
}

var InitCmd = &cobra.Command{
	Use:   "init -c [client_name]",
	Short: "Initialize a new OAuth client configuration",
	Long: `Initialize a new OAuth client configuration with the specified client name.
This command sets up the necessary configuration files for OAuth/OIDC authentication.`,
	Example: `  tokendokey init -c=myclient
  tokendokey init --client=myclient`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		clientName, _ := cmd.Flags().GetString("client")
		if clientName == "" {
			fmt.Println("Error: client name is required")
			return
		}

		configDir := filepath.Join(getHomeDir(), ".tokendokey", clientName)
		os.MkdirAll(configDir, os.ModePerm)

		configFilePath := filepath.Join(configDir, "config.json")
		refreshTokenPath := filepath.Join(configDir, "refresh_token.txt")
		accessTokenPath := filepath.Join(configDir, "access_token.txt")

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter Client ID: ")
		clientID, _ := reader.ReadString('\n')
		clientID = strings.TrimSpace(clientID)

		fmt.Print("Enter Client Secret (leave blank if not applicable): ")
		clientSecret, _ := reader.ReadString('\n')
		clientSecret = strings.TrimSpace(clientSecret)

		fmt.Print("Enter OAuth/OIDC Discovery URL/Well-Known URL: ")
		discoveryURL, _ := reader.ReadString('\n')
		discoveryURL = strings.TrimSpace(discoveryURL)

		// Fetch OAuth/OIDC discovery document
		resp, err := http.Get(discoveryURL)
		if err != nil {
			fmt.Println("Error fetching discovery document:", err)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var discoveryDoc map[string]interface{}
		json.Unmarshal(body, &discoveryDoc)

		tokenIssueURL, ok := discoveryDoc["token_endpoint"].(string)
		if !ok {
			fmt.Print("token_endpoint not found in discovery document. Please enter the token endpoint manually: ")
			tokenIssueURL, _ = reader.ReadString('\n')
			tokenIssueURL = strings.TrimSpace(tokenIssueURL)
		}

		deviceAuthURL, ok := discoveryDoc["device_authorization_endpoint"].(string)
		if !ok {
			fmt.Print("device_authorization_endpoint not found in discovery document. Please enter the device authorization endpoint manually: ")
			deviceAuthURL, _ = reader.ReadString('\n')
			deviceAuthURL = strings.TrimSpace(deviceAuthURL)
		}

		config := Config{
			ClientID:      clientID,
			ClientSecret:  clientSecret,
			TokenIssueURL: tokenIssueURL,
			DeviceCodeURL: deviceAuthURL,
		}

		configData, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile(configFilePath, configData, 0644)
		os.WriteFile(refreshTokenPath, []byte{}, 0644)
		os.WriteFile(accessTokenPath, []byte{}, 0644)

		fmt.Println("Configuration initialized successfully.")
	},
}

func init() {
	InitCmd.Flags().StringP("client", "c", "", "Client name for the OAuth configuration")
	InitCmd.MarkFlagRequired("client")
}

func getHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		os.Exit(1)
	}
	return home
}
