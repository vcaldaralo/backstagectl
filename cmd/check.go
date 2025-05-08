package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check issues in Backstage catalog",
}

var orphanCmd = &cobra.Command{
	Use:   "orphan",
	Short: "Orphan entities",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		entities := fetchEntitiesByQuery("filter=metadata.annotations.backstage.io/orphan=true")

		var data [][]string
		for _, entity := range entities {
			entityRef := getEntityRef(entity)
			_, namespace, name := getKindNamespaceName(entityRef)
			row := []string{
				namespace,
				name,
				getEntityUrl(entity),
			}
			data = append(data, row)
		}

		outputFormat, _ := cmd.Flags().GetString("output")
		header := []string{"NAMESPACE", "NAME", "URL"}
		formatOutput(header, data, outputFormat)
	},
}

var missingAnnotationCmd = &cobra.Command{
	Use:   "missingannotation [kind|entityRef] [annotation]",
	Short: "Annotation that is missing for a group of entities",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth()

		var annotation string

		if len(args) < 2 {
			log.Fatalf("Error: insufficient arguments. First argument must be 'kind' or 'entityRef', and second argument must be the annotation key.")
		} else if len(args) > 2 {
			log.Fatalf("Error: too many arguments provided. Please specify exactly two arguments: 'kind' or 'entityRef' and the annotation key.")
		}

		filter := parseArgs(args[:len(args)-1])

		annotation = args[1]

		params := fmt.Sprintf("fields=kind,metadata.namespace,metadata.name,metadata.annotation&%s", filter)

		entities := fetchEntitiesByQuery(params)

		var data [][]string
		for _, entity := range entities {
			_, ok := entity.Metadata.Annotations[annotation].(string)
			if !ok {
				entityRef := getEntityRef(entity)
				_, namespace, name := getKindNamespaceName(entityRef)
				row := []string{
					namespace,
					name,
					annotation,
					getEntityUrl(entity),
				}
				data = append(data, row)
			}
		}

		outputFormat, _ := cmd.Flags().GetString("output")
		header := []string{"NAMESPACE", "NAME", "MISSINGANNOTATION", "URL"}
		formatOutput(header, data, outputFormat)
	},
}

var entityNotFoundCmd = &cobra.Command{
	Use:   "notfound [kind|entityRef] [name]",
	Short: "Relations that don't exist for an entity",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		if len(args) == 0 {
			log.Fatalf("Error: no kind ([component|system|domain|group|user|location]) or entityRef ({kind}:{namespace}/{entity}) provided. Please specify one to check them")
		} else if len(args) > 2 {
			log.Fatalf("Error: too many arguments provided. Please specify either one or two arguments")
		}

		filter := parseArgs(args)

		params := fmt.Sprintf("fields=kind,metadata.namespace,metadata.name,relations&%s", filter)
		entities := fetchEntitiesByQuery(params)

		relationTarget := make(map[string][]string)
		for _, entity := range entities {
			entityRef := getEntityRef(entity)
			for _, rel := range entity.Relations {
				if rel.Type == "dependsOn" || rel.Type == "partOf" || rel.Type == "ownedBy" {
					relationTarget[rel.TargetRef] = append(relationTarget[rel.TargetRef], entityRef)
				}
			}
		}
		if len(relationTarget) == 0 {
			log.Fatalf("Relations 'dependsOn' or 'partOf' for these entities don't exist")
		}

		filterNotFoundEntities, _ := cmd.Flags().GetString("filter")
		filterNotFoundEntities = addNamespaceDefault(filterNotFoundEntities)

		var verifyEntityRef []string
		for notFoundEntity := range relationTarget {
			if strings.Contains(notFoundEntity, filterNotFoundEntities) {
				verifyEntityRef = append(verifyEntityRef, notFoundEntity)
			}
		}

		payload := Payload{
			EntityRefs: verifyEntityRef,
			Fields:     []string{"kind", "metadata.name"},
		}

		if len(verifyEntityRef) > 0 {
			entities = fetchEntitiesByRefs(payload)
		}

		var data [][]string
		for i, entity := range entities {
			if entity.Kind == "" {
				entityNotFound := cleanNamespaceDefault(verifyEntityRef[i])
				for _, usedin := range relationTarget[verifyEntityRef[i]] {
					entityRef := usedin
					_, namespace, name := getKindNamespaceName(entityRef)
					row := []string{
						namespace,
						name,
						entityNotFound,
						getEntityUrlFromRef(addNamespaceDefault(entityRef)),
					}
					data = append(data, row)
				}
			}
		}

		outputFormat, _ := cmd.Flags().GetString("output")
		header := []string{"NAMESPACE", "NAME", "ENTITYNOTFOUND", "URL"}
		formatOutput(header, data, outputFormat)
	},
}

func init() {
	checkCmd.AddCommand(orphanCmd)
	checkCmd.AddCommand(missingAnnotationCmd)
	checkCmd.AddCommand(entityNotFoundCmd)

	orphanCmd.Flags().StringP("output", "o", "table", "Output format [table|json]")
	missingAnnotationCmd.Flags().StringP("output", "o", "table", "Output format [table|json]")
	entityNotFoundCmd.Flags().StringP("output", "o", "table", "Output format [table|json]")
	entityNotFoundCmd.Flags().StringP("filter", "f", "", "Filter output on ENTITYNOTFOUND")

	rootCmd.AddCommand(checkCmd)
}
