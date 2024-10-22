package cmd

import (
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

var GetTokenCmd = &cobra.Command{
	Use:   "get-token -c=[client_name] -f",
	Short: "Get a new access token from [client_name]. If [--force] is specified, will force a refresh.",
	Long: `Get a new access token from the specified client.
If the forcerefresh parameter is provided, a refresh will be forced even if the current access token is still valid.`,
	Example: `  tokendokey get-token -c=myclient
	tokendokey get-token -c=myclient -f
  tokendokey get-token --client=myclient --force`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		clientName, _ := cmd.Flags().GetString("client")
		if clientName == "" {
			fmt.Println("Error: client name is required")
			return
		}

		forceRefresh, _ := cmd.Flags().GetBool("force")

		configDir := filepath.Join(getHomeDir(), ".tokendokey", clientName)
		configFilePath := filepath.Join(configDir, "config.json")
		refreshTokenPath := filepath.Join(configDir, "refresh_token.txt")
		accessTokenPath := filepath.Join(configDir, "access_token.txt")

		configData, _ := os.ReadFile(configFilePath)
		var config Config
		json.Unmarshal(configData, &config)

		accessToken, _ := os.ReadFile(accessTokenPath)
		if !forceRefresh && len(accessToken) > 0 && isTokenValid(string(accessToken), "access") {
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

		fmt.Println(newAccessToken)
	},
}

func init() {
	GetTokenCmd.Flags().StringP("client", "c", "", "Client name for the OAuth configuration")
	GetTokenCmd.Flags().BoolP("force", "f", false, "Force refresh even if the current access token is still valid.")
	GetTokenCmd.MarkFlagRequired("client")
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
