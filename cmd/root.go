package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type EntitiesResponse struct {
	Items    []Entity `json:"items"`
	PageInfo struct {
		NextCursor string `json:"nextCursor"`
	} `json:"pageInfo"`
	TotalItems int `json:"totalItems"`
}

type Entity struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name        string                 `json:"name"`
		Namespace   string                 `json:"namespace"`
		Description string                 `json:"description"`
		Annotations map[string]interface{} `json:"annotations"`
		Links       []interface{}          `json:"links"`
		Tags        []string               `json:"tags"`
	} `json:"metadata"`
	Relations []interface{}          `json:"relations"`
	Spec      map[string]interface{} `json:"spec"`
}

type Entities []Entity

var rootCmd = &cobra.Command{
	Use:   "backstage-cli",
	Short: "A CLI tool to interact with Backstage API",
	Long: `backstage-cli is a command line interface tool that allows you to 
interact with Backstage API in read-only mode. You can fetch information 
about entities, APIs, and other entities.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Placeholder functions for fetching entities, users, and groups
func fetchEntities(queryParameters string) []Entity {
	var entities []Entity
	var nextCursor string
	for {
		url := fmt.Sprintf("%s/api/catalog/entities/by-query%s", baseUrl, queryParameters)
		if nextCursor != "" {
			url += fmt.Sprintf("&cursor=%s", nextCursor) // Append nextCursor to the URL
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			return nil
		}

		addAuthHeader(req) // Add authentication header

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			log.Fatalf("%s", body)
		} else {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return nil
			}

			var entitiesResponse EntitiesResponse
			err = json.Unmarshal(body, &entitiesResponse)
			if err != nil {
				fmt.Println("Error unmarshalling JSON:", err)
				return nil
			}

			// Process the items
			entities = append(entities, entitiesResponse.Items...)

			// Check for nextCursor to continue fetching
			nextCursor = entitiesResponse.PageInfo.NextCursor
			if nextCursor == "" {
				break // Exit the loop if there are no more cursors
			}
		}
	}
	// Implement your logic to fetch all entities
	return entities
}

func displayEntities(entities Entities) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer w.Flush()

	// Print each key-value pair
	if len(entities) != 1 {
		// fmt.Fprintln(w, "KIND\tNAME\tOWNER\tURL")
		fmt.Fprintln(w, "KIND\tNAME")
		for _, entity := range entities {
			// fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			fmt.Fprintf(w, "%s\t%s\n",
				entity.Kind,
				entity.Metadata.Name,
				// entity.Spec["owner"],
				// entity.Metadata.Annotations["backstage.io/view-url"],
			) // Use Fprintf to write to the tabwriter
		}
		// Print the whole yaml for the single entity
	} else {
		// Print the entire YAML without filtering
		marshaledYAML, err := yaml.Marshal(entities[0])
		if err != nil {
			fmt.Println("Error marshalling YAML:", err)
			return
		}
		// Print the resulting YAML
		fmt.Println(string(marshaledYAML))
	}
}

func getEntityRef(entity Entity) string {
	return fmt.Sprintf("%s:%s/%s", strings.ToLower(entity.Kind), entity.Metadata.Namespace, entity.Metadata.Name)
}
