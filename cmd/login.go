package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var client *http.Client
var (
	baseUrl     string
	token       string
	tlsCertPath string
	tlsKeyPath  string
)

type AuthConfig struct {
	BaseUrl     string `json:"baseUrl"`
	Token       string `json:"token"`
	TLSCertPath string `json:"tls_cert_path"`
	TLSKeyPath  string `json:"tls_key_path"`
}

func initAuth() {
	// Load authentication details from file
	authConfig := loadAuthConfig("./.config.json")
	if authConfig != nil {
		baseUrl = authConfig.BaseUrl
		token = authConfig.Token
		tlsCertPath = authConfig.TLSCertPath
		tlsKeyPath = authConfig.TLSKeyPath
	}

	// Create TLS client if cert/key provided
	if tlsCertPath != "" && tlsKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
		if err != nil {
			fmt.Printf("Error loading TLS certificate: %v\n", err)
			return
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
		}
	} else {
		client = &http.Client{}
	}
}

func addAuthHeader(req *http.Request) {
	// Only add token auth if cert/key not provided
	if tlsCertPath == "" && tlsKeyPath == "" {
		if token == "" {
			fmt.Println("Error: either token or TLS certificate/key pair must be provided")
			return
		}
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	}
}

func saveAuthConfig(filename string) {
	authConfig := AuthConfig{
		BaseUrl:     baseUrl,
		Token:       token,
		TLSCertPath: tlsCertPath,
		TLSKeyPath:  tlsKeyPath,
	}

	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating auth config file: %v\n", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(authConfig); err != nil {
		fmt.Printf("Error saving auth config: %v\n", err)
	}
}

func loadAuthConfig(filename string) *AuthConfig {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// File does not exist, return nil
			return nil
		}
		fmt.Printf("Error opening auth config file: %v\n", err)
		return nil
	}
	defer file.Close()

	var authConfig AuthConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&authConfig); err != nil {
		fmt.Printf("Error loading auth config: %v\n", err)
		return nil
	}

	if authConfig.BaseUrl == "" {
		fmt.Println("Error: No baseUrl for the backstage instance provided")
		return nil
	}

	// Validate loaded values
	if authConfig.Token == "" && (authConfig.TLSCertPath == "" || authConfig.TLSKeyPath == "") {
		fmt.Println("Error: No valid authentication details found in ./.config.json")
		return nil
	}

	return &authConfig
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Save authentication details to a file",
	Run: func(cmd *cobra.Command, args []string) {
		baseUrl, _ = cmd.Flags().GetString("baseUrl")
		token, _ = cmd.Flags().GetString("token")
		tlsCertPath, _ = cmd.Flags().GetString("tls-cert")
		tlsKeyPath, _ = cmd.Flags().GetString("tls-key")

		if baseUrl == "" {
			fmt.Println("Error: No baseUrl for the backstage instance provided")
			return
		}
		// Validate input
		if token == "" && (tlsCertPath == "" || tlsKeyPath == "") {
			fmt.Println("Error: You must provide either a token or both TLS certificate and key paths")
			return
		}

		// Save the authentication details to ./.config.json
		saveAuthConfig("./.config.json")
		fmt.Println("Authentication details saved to ./.config.json")
	},
}

func init() {
	loginCmd.Flags().StringP("baseUrl", "u", "", "Backstage API base URL (required)")
	loginCmd.Flags().StringP("token", "t", "", "Authentication token")
	loginCmd.Flags().StringP("tls-cert", "c", "", "Path to TLS certificate")
	loginCmd.Flags().StringP("tls-key", "k", "", "Path to TLS key")
	rootCmd.MarkPersistentFlagRequired("url")
	rootCmd.AddCommand(loginCmd) // Add the new command to the root command
}
