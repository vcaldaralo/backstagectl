package cmd

import (
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

type EntitiesResponse struct {
	Items []Entity `json:"items"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all entities from Backstage",
	Run: func(cmd *cobra.Command, args []string) {
		client := &http.Client{}

		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/catalog/entities", baseURL), nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}

		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

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
			var entities EntitiesResponse
			if err := json.Unmarshal(body, &entities); err != nil {
				fmt.Printf("Error parsing response: %v\n", err)
				return
			}

			for _, entity := range entities.Items {
				// if entity.Kind == "Component" {
				fmt.Printf("Name: %s\nDescription: %s\n\n",
					entity.Metadata.Name,
					entity.Metadata.Description)
				// }
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
