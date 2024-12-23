package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Check if the current access token is available and valid
func isAccessTokenValid(accessTokenPath string) (string, error) {
	accessToken, err := os.ReadFile(accessTokenPath)
	if err != nil {
		return err.Error(), err
	}

	if isTokenValid(string(accessToken), "access") {
		return string(accessToken), nil
	}

	return "Access token is invalid", fmt.Errorf("access token is invalid")
}

// Check if the refresh token is available and valid
func isRefreshTokenValid(refreshTokenPath string) (string, error) {
	refreshToken, err := os.ReadFile(refreshTokenPath)
	if err != nil {
		return err.Error(), err
	}

	if isTokenValid(string(refreshToken), "refresh") {
		return string(refreshToken), nil
	}
	return "Refresh token is invalid", fmt.Errorf("refresh token is invalid")
}

// Refresh the access token using the refresh token
func refreshAccessToken(refreshTokenPath, configFilePath, accessTokenPath string) (string, error) {
	var config Config
	configData, err := os.ReadFile(configFilePath)
	if err != nil {
		return err.Error(), err
	}
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return err.Error(), err
	}

	refreshToken, err := os.ReadFile(refreshTokenPath)
	if err != nil {
		return err.Error(), err
	}

	refreshTokenReq := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {string(refreshToken)},
		"client_id":     {config.ClientID},
	}
	if config.ClientSecret != "" {
		refreshTokenReq.Set("client_secret", config.ClientSecret)
	}

	refreshTokenResp, err := http.Post(config.TokenIssueURL, "application/x-www-form-urlencoded", strings.NewReader(refreshTokenReq.Encode()))
	if err != nil {
		return err.Error(), err
	}
	defer refreshTokenResp.Body.Close()

	body, _ := io.ReadAll(refreshTokenResp.Body)
	var refreshTokenResponse map[string]string
	json.Unmarshal(body, &refreshTokenResponse)

	if refreshTokenResponse["access_token"] != "" {
		newAccessToken := refreshTokenResponse["access_token"]
		newRefreshToken := refreshTokenResponse["refresh_token"]
		os.WriteFile(accessTokenPath, []byte(newAccessToken), 0644)
		os.WriteFile(refreshTokenPath, []byte(newRefreshToken), 0644)

		return newAccessToken, nil
	}

	return "Failed to refresh access token", fmt.Errorf("failed to refresh access token")
}

// Get a new access token using mTLS Direct Grant flow
func getNewAccessToken(clientCertPath, clientKeyPath, caCertPath, configFilePath, accessTokenPath, refreshTokenPath string) (string, error) {
	var config Config
	configData, err := os.ReadFile(configFilePath)
	if err != nil {
		return err.Error(), err
	}
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return err.Error(), err
	}

	cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return err.Error(), err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	if caCertPath != "" {
		caCert, err := os.ReadFile(caCertPath)
		if err != nil {
			return err.Error(), err
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig.RootCAs = caCertPool
	} else {
		tlsConfig.InsecureSkipVerify = true
	}

	tr := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: tr}

	directGrantReq := url.Values{
		"grant_type": {"password"},
		"client_id":  {config.ClientID},
	}
	if config.ClientSecret != "" {
		directGrantReq.Set("client_secret", config.ClientSecret)
	}

	directGrantResp, err := client.Post(config.TokenIssueURL, "application/x-www-form-urlencoded", strings.NewReader(directGrantReq.Encode()))
	if err != nil {
		return err.Error(), err
	}
	defer directGrantResp.Body.Close()

	body, _ := io.ReadAll(directGrantResp.Body)
	var directGrantResponse map[string]string
	json.Unmarshal(body, &directGrantResponse)

	if directGrantResponse["access_token"] != "" {
		newAccessToken := directGrantResponse["access_token"]
		newRefreshToken := directGrantResponse["refresh_token"]
		os.WriteFile(accessTokenPath, []byte(newAccessToken), 0644)
		os.WriteFile(refreshTokenPath, []byte(newRefreshToken), 0644)

		return newAccessToken, nil
	}

	return "Failed to get new access token", fmt.Errorf("failed to get new access token")
}

var MTLSTokenCmd = &cobra.Command{
	Use:   "mtls-token -c=[client_name] -cert=[client_cert_path] -key=[client_key_path]  -caCert=[ca_cert_path]",
	Short: "Get Access Token from [client_name] through OAuth service using mTLS Direct Grant flow.",
	Long: `Get Access Token from the specified client through the OAuth service using mTLS Direct Grant flow.
You need to provide the client certificate and key paths.`,
	Example: `  tokendokey mtls-token -c=myclient -cert=path/to/client.crt -key=path/to/client.key  -caCert=path/to/ca.crt
  tokendokey mtls-token --client=myclient --cert=path/to/client.crt --key=path/to/client.key --caCert=path/to/ca.crt`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		clientName, _ := cmd.Flags().GetString("client")
		if clientName == "" {
			fmt.Println("Error: client name is required")
			return
		}

		clientCertPath, _ := cmd.Flags().GetString("cert")
		if clientCertPath == "" {
			fmt.Println("Error: client certificate path is required")
			return
		}

		clientKeyPath, _ := cmd.Flags().GetString("key")
		if clientKeyPath == "" {
			fmt.Println("Error: client key path is required")
			return
		}

		caCertPath, _ := cmd.Flags().GetString("caCert")

		configDir := filepath.Join(getHomeDir(), ".tokendokey", clientName)
		configFilePath := filepath.Join(configDir, "config.json")
		refreshTokenPath := filepath.Join(configDir, "refresh_token.txt")
		accessTokenPath := filepath.Join(configDir, "access_token.txt")

		// Check if access token is available and valid
		validAccessToken, err := isAccessTokenValid(accessTokenPath)
		if err != nil {
			// cached access token is invalid, next check refresh token
			_, err := isRefreshTokenValid(refreshTokenPath)
			if err != nil {
				// no refresh token, get new access token
				validAccessToken, err = getNewAccessToken(clientCertPath, clientKeyPath, caCertPath, configFilePath, accessTokenPath, refreshTokenPath)
				if err != nil {
					fmt.Println("Error getting new access token:", err)
					return
				}
				fmt.Println(validAccessToken)
				return
			} else {
				// refresh token is valid, refresh access token
				validAccessToken, err := refreshAccessToken(refreshTokenPath, configFilePath, accessTokenPath)
				if err != nil {
					fmt.Println("Error refreshing access token:", err)
					return
				}
				fmt.Println(validAccessToken)
				return

			}
		} else {
			fmt.Println(validAccessToken)
			return
		}

	},
}

func init() {
	MTLSTokenCmd.Flags().StringP("client", "c", "", "Client name for the OAuth configuration")
	MTLSTokenCmd.Flags().StringP("cert", "t", "", "Path to the client certificate file")
	MTLSTokenCmd.Flags().StringP("key", "k", "", "Path to the client key file")
	MTLSTokenCmd.Flags().StringP("caCert", "r", "", "Path to the server certificate file (optional)")
	MTLSTokenCmd.MarkFlagRequired("client")
	MTLSTokenCmd.MarkFlagRequired("cert")
	MTLSTokenCmd.MarkFlagRequired("key")
}
