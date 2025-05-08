package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var getCmd = &cobra.Command{
	Use:   "get [kind|entityRef] [name]",
	Short: "Display one or many Backstage entities",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication
		annotation, _ := cmd.Flags().GetString("annotation")

		if len(args) == 0 {
			log.Fatalf("Error: no kind ([component|system|domain|group|user|location]) or entityRef ({kind}:{namespace}/{entity}) provided. Please specify one to check them")
		}

		filter := parseArgs(args)

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
			fmt.Print(string(marshaledYAML))
		} else {
			var data [][]string
			for _, entity := range entities {
				entityRef := getEntityRef(entity)
				_, namespace, name := getKindNamespaceName(entityRef)
				newRow := []string{
					namespace,
					name,
					getEntityUrl(entity),
				}
				data = append(data, newRow)
			}

			outputFormat, _ := cmd.Flags().GetString("output")
			header := []string{"NAMESPACE", "NAME", "URL"}
			formatOutput(header, data, outputFormat)
		}
	},
}

func init() {
	getCmd.Flags().StringP("annotation", "a", "", "Filter entities by annotation key")
	getCmd.Flags().StringP("output", "o", "table", "Output format [table|json]")
	rootCmd.AddCommand(getCmd)
}
