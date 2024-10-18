package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/cobra"
)

type Config struct {
	ClientID      string `json:"client_id"`
	ClientSecret  string `json:"client_secret"`
	TokenIssueURL string `json:"token_issue_url"`
}

func main() {
	var rootCmd = &cobra.Command{Use: "tokendokey"}

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(getTokenCmd)

	rootCmd.Execute()
}

var initCmd = &cobra.Command{
	Use:   "init [client_name]",
	Short: "Initialize a new OAuth client configuration",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clientName := args[0]
		configDir := filepath.Join(".", clientName)
		os.MkdirAll(configDir, os.ModePerm)

		configFilePath := filepath.Join(configDir, "config.json")
		refreshTokenPath := filepath.Join(configDir, "refresh-token.txt")
		accessTokenPath := filepath.Join(configDir, "access-token.txt")

		reader := bufio.NewReader(os.Stdin)

		fmt.Print("Enter Client ID: ")
		clientID, _ := reader.ReadString('\n')
		clientID = strings.TrimSpace(clientID)

		fmt.Print("Enter Client Secret (leave blank if not applicable): ")
		clientSecret, _ := reader.ReadString('\n')
		clientSecret = strings.TrimSpace(clientSecret)

		fmt.Print("Enter Token Issue URL: ")
		tokenIssueURL, _ := reader.ReadString('\n')
		tokenIssueURL = strings.TrimSpace(tokenIssueURL)

		config := Config{
			ClientID:      clientID,
			ClientSecret:  clientSecret,
			TokenIssueURL: tokenIssueURL,
		}

		configData, _ := json.MarshalIndent(config, "", "  ")
		os.WriteFile(configFilePath, configData, 0644)
		os.WriteFile(refreshTokenPath, []byte{}, 0644)
		os.WriteFile(accessTokenPath, []byte{}, 0644)

		fmt.Println("Configuration initialized successfully.")
	},
}

var getTokenCmd = &cobra.Command{
	Use:   "get-token [client_name]",
	Short: "Get a new access token",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clientName := args[0]
		configDir := filepath.Join(".", clientName)
		configFilePath := filepath.Join(configDir, "config.json")
		refreshTokenPath := filepath.Join(configDir, "refresh-token.txt")
		accessTokenPath := filepath.Join(configDir, "access-token.txt")

		configData, _ := os.ReadFile(configFilePath)
		var config Config
		json.Unmarshal(configData, &config)

		accessToken, _ := os.ReadFile(accessTokenPath)
		if len(accessToken) > 0 && isTokenValid(string(accessToken), "access") {
			// fmt.Println("Access token is still valid. Access Token:")
			fmt.Println(string(accessToken))
			return
		}

		refreshToken, _ := os.ReadFile(refreshTokenPath)
		if len(refreshToken) == 0 || !isTokenValid(string(refreshToken), "refresh") {
			fmt.Println("Refresh token is invalid. Please get new Refresh token.")
			return
		}

		form := url.Values{
			"client_id":     {config.ClientID},
			"grant_type":    {"refresh_token"},
			"refresh_token": {string(refreshToken)},
		}

		if config.ClientSecret != "" {
			form.Add("client_secret", config.ClientSecret)
		}

		req, _ := http.NewRequest("POST", config.TokenIssueURL, strings.NewReader(form.Encode()))
		// if config.ClientSecret != "" {
		// 	req.SetBasicAuth(config.ClientID, config.ClientSecret)
		// }
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Println("Error getting new access token:", err)
			return
		}
		if resp.StatusCode != http.StatusOK {
			fmt.Println("Error getting new access token:", resp.StatusCode)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		var tokenResponse map[string]string
		json.Unmarshal(body, &tokenResponse)

		newAccessToken := tokenResponse["access_token"]
		newrefreshToken := tokenResponse["refresh_token"]

		os.WriteFile(accessTokenPath, []byte(newAccessToken), 0644)
		os.WriteFile(refreshTokenPath, []byte(newrefreshToken), 0644)

		// fmt.Println("New access token obtained successfully. Access Token:")
		fmt.Println(newAccessToken)
	},
}

func isTokenValid(tokenString string, tokenType string) bool {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			expirationTime := time.Unix(int64(exp), 0)

			var validityDuration time.Duration
			switch tokenType {
			case "access":
				validityDuration = 30 * time.Second
			case "refresh":
				validityDuration = 1 * time.Minute
			default:
				validityDuration = 30 * time.Second
			}

			return time.Now().Before(expirationTime.Add(-validityDuration))
		}
	}
	return false
}
