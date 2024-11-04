package cmd

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// Generate a cryptographically secure random string as the code verifier
func generateCodeVerifier() (string, error) {
	verifier := make([]byte, 43)
	_, err := rand.Read(verifier)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(verifier), nil
}

// Hash the code verifier using SHA-256 and base64 URL encode it to create the code challenge
func generateCodeChallenge(verifier string) (string, error) {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:]), nil
}

var LoginCmd = &cobra.Command{
	Use:   "login -c=[client_name] [-o|--offline-token]",
	Short: "Login to [client_name] through OAuth service using Device Code flow. When [offline-token] is provided, will get offline token instead of a regular refresh token.",
	Long: `Login to the specified client through the OAuth service using the Device Code flow.
If the -ot|--offline-token flag is provided, an offline token will be obtained instead of a regular refresh token.`,
	Example: `  tokendokey login -c=myclient
	tokendokey login -c=myclient -o
  tokendokey login --client=myclient
  tokendokey login --client=myclient --offline-token`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		clientName, _ := cmd.Flags().GetString("client")
		if clientName == "" {
			fmt.Println("Error: client name is required")
			return
		}

		offlineToken, _ := cmd.Flags().GetBool("offline-token")

		configDir := filepath.Join(getHomeDir(), ".tokendokey", clientName)
		configFilePath := filepath.Join(configDir, "config.json")
		refreshTokenPath := filepath.Join(configDir, "refresh_token.txt")
		accessTokenPath := filepath.Join(configDir, "access_token.txt")

		// Load configuration
		var config Config
		configData, err := os.ReadFile(configFilePath)
		if err != nil {
			fmt.Println("Error loading configuration:", err)
			return
		}
		err = json.Unmarshal(configData, &config)
		if err != nil {
			fmt.Println("Error unmarshaling configuration:", err)
			return
		}

		// Generate code verifier and challenge
		verifier, err := generateCodeVerifier()
		if err != nil {
			fmt.Println("Error generating code verifier:", err)
			return
		}
		challenge, err := generateCodeChallenge(verifier)
		if err != nil {
			fmt.Println("Error generating code challenge:", err)
			return
		}

		// Request device code
		deviceCodeURL := config.DeviceCodeURL
		deviceCodeReq := url.Values{
			"client_id":             {config.ClientID},
			"code_challenge":        {challenge},
			"code_challenge_method": {"S256"},
		}

		if offlineToken {
			deviceCodeReq.Add("scope", "offline_access")
		}

		deviceCodeResp, err := http.Post(deviceCodeURL, "application/x-www-form-urlencoded", strings.NewReader(deviceCodeReq.Encode()))
		if err != nil {
			fmt.Println("Error requesting device code:", err)
			return
		}
		defer deviceCodeResp.Body.Close()

		body, _ := io.ReadAll(deviceCodeResp.Body)
		var deviceCodeResponse map[string]string
		json.Unmarshal(body, &deviceCodeResponse)

		if deviceCodeResponse["error"] != "" {
			fmt.Println("Error requesting device code:", deviceCodeResponse["error"])
			return
		}

		// Prompt user to visit URL and enter code
		verificationURI := deviceCodeResponse["verification_uri_complete"]
		if verificationURI == "" {
			verificationURI = deviceCodeResponse["verification_uri"]
			fmt.Println("Please visit the following URL and enter the code:")
			fmt.Println(verificationURI)
			fmt.Println("Enter the user code:", deviceCodeResponse["user_code"])
		} else {
			fmt.Println("Please visit the following URL and enter the code:")
			fmt.Println(verificationURI)
		}
		// Prompt user to visit URL and enter code, once done from browser, press any key to continue.
		fmt.Println("Once finshed on browser, Press any key to continue...")
		fmt.Scanln()

		// Poll for authorization
		pollURL := config.TokenIssueURL
		pollReq := url.Values{
			"grant_type":    {"urn:ietf:params:oauth:grant-type:device_code"},
			"device_code":   {deviceCodeResponse["device_code"]},
			"client_id":     {config.ClientID},
			"code_verifier": {verifier},
		}
		for {
			pollResp, err := http.Post(pollURL, "application/x-www-form-urlencoded", strings.NewReader(pollReq.Encode()))
			if err != nil {
				fmt.Println("Error polling for authorization:", err)
				return
			}

			defer pollResp.Body.Close()

			if pollResp.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(pollResp.Body)
				var pollResponse map[string]string
				json.Unmarshal(body, &pollResponse)

				if pollResponse["access_token"] != "" {
					// Authorization successful, obtain access token
					newAccessToken := pollResponse["access_token"]
					newrefreshToken := pollResponse["refresh_token"]
					fmt.Println("Access token obtained:")
					fmt.Println(newAccessToken)

					// Write access token to file
					os.WriteFile(accessTokenPath, []byte(newAccessToken), 0644)
					os.WriteFile(refreshTokenPath, []byte(newrefreshToken), 0644)
					return
				}
			}

			// Authorization not yet granted, wait and try again
			fmt.Println("Error polling for authorization: Response code", pollResp.Status)
			fmt.Println("Keep trying in 5 seconds. Waiting for you to open the URL above. Or exit")
			time.Sleep(5 * time.Second)

		}
	},
}

func init() {
	LoginCmd.Flags().StringP("client", "c", "", "Client name for the OAuth configuration")
	LoginCmd.Flags().BoolP("offline-token", "o", false, "Get offline token instead of a regular refresh token")
	LoginCmd.MarkFlagRequired("client")
}
