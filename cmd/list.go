package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all entities from Backstage",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		kind, _ := cmd.Flags().GetString("kind")
		annotationKey, _ := cmd.Flags().GetString("annotation")
		filter := ""

		if kind != "" {
			if len(kind) > 0 && kind[len(kind)-1] == 's' {
				kind = kind[:len(kind)-1] // Remove the last character
			}
			filter = fmt.Sprintf("?filter=kind=%s", kind)
		}

		if annotationKey != "" {
			if filter != "" {
				filter += fmt.Sprintf("&filter=metadata.annotations.%s", annotationKey)
			} else {
				filter = fmt.Sprintf("?filter=metadata.annotations.%s", annotationKey)
			}
		}

		url := fmt.Sprintf("%s/api/catalog/entities%s", baseUrl, filter)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return
		}

		addAuthHeader(req) // Add authentication header

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
				fmt.Printf("kind: %s\tname: %s\n",
					entity.Kind,
					entity.Metadata.Name,
				)
			}
		}
	},
}

func init() {
	listCmd.Flags().StringP("kind", "k", "", "Filter entities by kind [Resource|Component|System|Domain|User|Group|Location]")
	listCmd.Flags().StringP("annotation", "a", "", "Filter entities by annotation key")
	rootCmd.AddCommand(listCmd)
}
