package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func displayEntities(entities Entities) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer w.Flush()

	// Print each key-value pair
	if len(entities) != 1 {
		fmt.Fprintln(w, "KIND\tNAME")
		for _, entity := range entities {
			fmt.Fprintf(w, "%s\t%s\n", entity.Kind, entity.Metadata.Name) // Use Fprintf to write to the tabwriter
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

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Display one or many Backstage entities",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		var kind, name string

		if len(args) > 0 {
			kind = args[0] // Assign first argument to kind
			allowedKinds := map[string]bool{
				"resources": true,
				"component": true,
				"system":    true,
				"domain":    true,
				"user":      true,
				"group":     true,
				"location":  true,
			} // Define allowed kinds
			if !allowedKinds[kind] {
				log.Fatalf("error: backstage doesn't have a resource kind '%s'\nAllowed kinds are: %v", kind, allowedKinds)
			}
		}
		if len(args) > 1 {
			name = args[1] // Assign second argument to name
		}

		annotationKey, _ := cmd.Flags().GetString("annotation")

		filter := ""
		if kind != "" {
			if len(kind) > 0 && kind[len(kind)-1] == 's' {
				kind = kind[:len(kind)-1] // Remove the last character
			}
			filter = fmt.Sprintf("?filter=kind=%s", kind)
		}
		if name != "" {
			if filter != "" {
				filter += fmt.Sprintf(",metadata.name=%s", name)
			} else {
				filter = fmt.Sprintf("?filter=metadata.name=%s", name)
			}
		}
		if annotationKey != "" {
			if filter != "" {
				filter += fmt.Sprintf(",metadata.annotations.%s", annotationKey)
			} else {
				filter = fmt.Sprintf("?filter=metadata.annotations.%s", annotationKey)
			}
		}

		var entities []Entity
		var nextCursor string
		for {
			url := fmt.Sprintf("%s/api/catalog/entities/by-query%s", baseUrl, filter)
			if nextCursor != "" {
				if filter != "" {
					url += fmt.Sprintf("&cursor=%s", nextCursor) // Append nextCursor to the URL
				} else {
					url += fmt.Sprintf("?cursor=%s", nextCursor)
				}
			}

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

				var entitiesResponse EntitiesResponse
				err = json.Unmarshal(body, &entitiesResponse)
				if err != nil {
					fmt.Println("Error unmarshalling JSON:", err)
					return
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
		displayEntities(entities)
	},
}

func init() {
	// getCmd.Flags().StringP("kind", "k", "", "Filter entities by kind [resource|component|system|domain|user|group|location]")
	getCmd.Flags().StringP("annotation", "a", "", "Filter entities by annotation key")
	rootCmd.AddCommand(getCmd)
}
