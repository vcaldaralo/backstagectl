package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var getEntityCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a specific entity and print its YAML",
	Args:  cobra.NoArgs, // No positional arguments required
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		ref, _ := cmd.Flags().GetString("ref")                 // Get the ref flag
		kind, _ := cmd.Flags().GetString("kind")               // Get the kind flag
		name, _ := cmd.Flags().GetString("name")               // Get the name flag
		namespace, _ := cmd.Flags().GetString("namespace")     // Get the namespace flag
		ancestry, _ := cmd.Flags().GetBool("ancestry")         // Get the filter flag
		filterRelations, _ := cmd.Flags().GetBool("relations") // Get the filter flag

		if len(kind) > 0 && kind[len(kind)-1] == 's' {
			kind = kind[:len(kind)-1] // Remove the last character
		}

		// Split ref into kind, namespace, and name
		if ref != "" {
			tokens := strings.Split(ref, ":") // Split by ':'
			if len(tokens) == 2 {
				kind = tokens[0]                   // First token is kind
				t := strings.Split(tokens[1], "/") // Split by '/'
				if len(t) == 2 {
					namespace = t[0]
					name = t[1]
				} else if len(t) == 1 {
					name = t[0] // Third token is name
				}
			}
		}

		anc := map[bool]string{false: "", true: "/ancestry"}[ancestry]

		url := fmt.Sprintf("%s/api/catalog/entities/by-name/%s/%s/%s%s", baseUrl, kind, namespace, name, anc) // Use ref for URL

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
			log.Fatalf("Status code: %d, %s\n%s", resp.StatusCode, url, body)
		} else {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response: %v\n", err)
				return
			}

			var data map[string]interface{}
			err = yaml.Unmarshal(body, &data)
			if err != nil {
				fmt.Println("Error unmarshalling YAML:", err)
				return
			}

			if !filterRelations {
				// Keys to filter (exclude these keys)
				excludeKeys := map[string]bool{"relations": true}

				// Filtered map
				filtered := map[string]interface{}{}
				for key, value := range data {
					if !excludeKeys[key] {
						filtered[key] = value
					}
				}

				// Marshal the map back to YAML
				marshaledYAML, err := yaml.Marshal(filtered)
				if err != nil {
					fmt.Println("Error marshalling YAML:", err)
					return
				}
				// Print the resulting YAML
				fmt.Println(string(marshaledYAML))
			} else {
				// Print the entire YAML without filtering
				marshaledYAML, err := yaml.Marshal(data)
				if err != nil {
					fmt.Println("Error marshalling YAML:", err)
					return
				}
				// Print the resulting YAML
				fmt.Println(string(marshaledYAML))
			}
		}
	},
}

func init() {
	getEntityCmd.Flags().StringP("ref", "r", "", "The reference of the entity {kind}:[default/]{name} (ie. component:default/zookeeper)")   // Add kind flag
	getEntityCmd.Flags().StringP("kind", "k", "component", "The kind of the entity [Resource|Component|System|Domain|User|Group|Location]") // Add kind flag
	getEntityCmd.Flags().StringP("name", "n", "", "The name of the entity")                                                                 // Add name flag
	getEntityCmd.Flags().StringP("namespace", "N", "default", "The namespace of the entity")                                                // Add namespace flag
	getEntityCmd.Flags().BoolP("ancestry", "a", false, "Wheter or not tho retrieve the ancestry of the entity")                             // Add name flag
	getEntityCmd.Flags().BoolP("relations", "l", false, "Print all relations")                                                              // Add filter flag
	// Mark name as required
	rootCmd.AddCommand(getEntityCmd) // Add the new command to the list command
}
