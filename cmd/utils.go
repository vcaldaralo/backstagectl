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

	"gopkg.in/yaml.v3"
)

// getEntityRef is a placeholder for your actual implementation.
// Replace this with your actual function logic.
func getEntityRef(entity Entity) string {
	return fmt.Sprintf("%s:%s/%s", strings.ToLower(entity.Kind), entity.Metadata.Namespace, entity.Metadata.Name)

}

// parseArgs parses command line arguments and returns relevant values.
func parseArgs(args []string) ([]string, string) {
	var kind, name, filter string

	if len(args) > 0 {
		kind = args[0] // Assign first argument to kind

		allowedKinds := map[string]bool{
			"resource":  true,
			"component": true,
			"system":    true,
			"domain":    true,
			"user":      true,
			"group":     true,
			"location":  true,
		} // Define allowed kinds

		if len(kind) > 0 && kind[len(kind)-1] == 's' {
			kind = kind[:len(kind)-1] // Remove the last character
		}
		if !allowedKinds[kind] {
			log.Fatalf("error: backstage doesn't have a resource kind '%s'\nAllowed kinds are: %v", kind, allowedKinds)
		}
	}

	if len(args) > 1 && name == "" {
		name = args[1] // Assign second argument to name
	}

	if kind != "" {
		if len(kind) > 0 && kind[len(kind)-1] == 's' {
			kind = kind[:len(kind)-1] // Remove the last character
		}
		filter = fmt.Sprintf("filter=kind=%s", kind)
	}
	if name != "" {
		if filter != "" {
			filter += fmt.Sprintf(",metadata.name=%s", name)
		} else {
			filter = fmt.Sprintf("filter=metadata.name=%s", name)
		}
	}

	return []string{kind, name}, filter
}

func fetchEntities(queryParameters string) []Entity {
	var entities []Entity
	var nextCursor string
	for {
		url := fmt.Sprintf("%s/api/catalog/entities/by-query?%s", baseUrl, queryParameters)
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

func displayEntities(output [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer w.Flush()

	// Print each key-value pair
	if len(output) != 2 {
		header := output[0]
		fmt.Fprintln(w, strings.Join(header, "\t"))

		for _, row := range output[1:] { // Skip the header
			fmt.Fprintln(w, strings.Join(row, "\t"))
		}
	} else {
		// Print the entire YAML without filtering
		marshaledYAML, err := yaml.Marshal(output[1:])
		if err != nil {
			fmt.Println("Error marshalling YAML:", err)
			return
		}
		// Print the resulting YAML
		fmt.Println(string(marshaledYAML))
	}
}
