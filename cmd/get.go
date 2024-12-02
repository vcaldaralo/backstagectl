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

		_, _, _, filter := parseArgs(args)

		if annotation != "" {
			if filter != "" {
				filter += fmt.Sprintf(",metadata.annotations.%s", annotation)
			} else {
				filter = fmt.Sprintf("filter=metadata.annotations.%s", annotation)
			}
		}

		entities := fetchEntitiesByQuery(filter)

		if len(entities) == 1 {
			entity := entities[0]
			entities[0].Metadata.Annotations["backstage.io/web-url"] = getEntityUrl(entity)
			entities[0].Metadata.Annotations["backstage.io/entity-ref"] = getEntityRef(entity)
			marshaledYAML, err := yaml.Marshal(entities[0])
			if err != nil {
				fmt.Println("error marshalling YAML:", err)
				return
			}
			// Print the resulting YAML
			fmt.Print(string(marshaledYAML))
		} else {
			var data [][]string
			for _, entity := range entities {
				newRow := []string{entity.Metadata.Namespace, entity.Kind, entity.Metadata.Name, getEntityRef(entity), getEntityUrl(entity)}
				// if annotation != "" {
				// 	newRow = []string{entity.Kind, entity.Metadata.Name, entity.Metadata.Annotations[annotation].(string), viewUrl}
				// }
				data = append(data, newRow)
			}
			header := []string{"NAMESPACE", "KIND", "NAME", "ENTITYREF", "URL"}
			displayEntities(header, data)
		}
	},
}

func init() {
	getCmd.Flags().StringP("annotation", "a", "", "Filter entities by annotation key")
	rootCmd.AddCommand(getCmd)
}
