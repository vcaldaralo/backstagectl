package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

type Entity struct {
	Metadata struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"metadata"`
	Kind string `json:"kind"`
}

// type EntitiesResponse struct {
// 	Items []Entity `json:"items"`
// }

type Entities []Entity

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all entities from Backstage",
	Run: func(cmd *cobra.Command, args []string) {
		var client *http.Client

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

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/catalog/entities?filter=kind=component", baseURL), nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}

		// Only add token auth if cert/key not provided
		if tlsCertPath == "" && tlsKeyPath == "" {
			if token == "" {
				fmt.Println("Error: either token or TLS certificate/key pair must be provided")
				return
			}
			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Fatalf("%s", body)
		} else {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			var entities Entities
			if err := json.Unmarshal(body, &entities); err != nil {
				fmt.Printf("Error parsing response: %v\n", err)
				return
			}

			for _, entity := range entities {
				// if entity.Kind == "Component" {
				fmt.Printf("Name: %s\nKind: %s\nDescription: %s\n\n",
					entity.Metadata.Name,
					entity.Kind,
					entity.Metadata.Description)
				// }
			}
			// fmt.Printf("%s", body)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
