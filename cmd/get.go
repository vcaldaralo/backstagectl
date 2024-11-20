package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var getCmd = &cobra.Command{
	Use:   "get [kind|ref] [name]",
	Short: "Display one or many Backstage entities",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication
		annotation, _ := cmd.Flags().GetString("annotation")

		_, filter := parseArgs(args)

		if annotation != "" {
			if filter != "" {
				filter += fmt.Sprintf(",metadata.annotations.%s", annotation)
			} else {
				filter = fmt.Sprintf("filter=metadata.annotations.%s", annotation)
			}
		}

		entities := fetchEntities(filter)

		if len(entities) == 1 {
			marshaledYAML, err := yaml.Marshal(entities)
			if err != nil {
				fmt.Println("Error marshalling YAML:", err)
				return
			}
			// Print the resulting YAML
			fmt.Println(string(marshaledYAML))
		} else {
			output := [][]string{
				{"KIND", "NAME", "URL"},
			}
			// if annotation != "" {
			// 	output = [][]string{
			// 		{"KIND", "NAME", strings.ToUpper(annotation), "URL"},
			// 	}
			// }

			for _, entity := range entities {
				viewUrl := entity.Metadata.Annotations["backstage.io/view-url"].(string)
				newRow := []string{entity.Kind, entity.Metadata.Name, viewUrl}
				// if annotation != "" {
				// 	newRow = []string{entity.Kind, entity.Metadata.Name, entity.Metadata.Annotations[annotation].(string), viewUrl}
				// }
				output = append(output, newRow)
			}

			displayEntities(output)

		}
	},
}

func init() {
	getCmd.Flags().StringP("annotation", "a", "", "Filter entities by annotation key")
	rootCmd.AddCommand(getCmd)
}
