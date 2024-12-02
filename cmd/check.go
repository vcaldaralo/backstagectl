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
			row := []string{getEntityRef(entity), getEntityUrl(entity)}
			data = append(data, row)
		}
		header := []string{"ENTITYREF", "URL"}
		displayEntities(header, data)
	},
}

var missingAnnotationCmd = &cobra.Command{
	Use:   "missing-annotation [annotation] [kind]",
	Short: "Annotation that is missing for a group of entities",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth()

		var annotation string

		if len(args) == 0 {
			log.Fatalf("Error: no annotation key provided. Please specify an annotation key.")
		} else if len(args) > 2 {
			log.Fatalf("Error: too many arguments provided. Please specify either one or two arguments.")
		}

		annotation = args[0]

		kind := ""
		filter := ""
		if len(args) == 2 {
			kind = args[1]
		}
		if kind != "" {
			if len(kind) > 0 && kind[len(kind)-1] == 's' {
				kind = kind[:len(kind)-1] // Remove the last character
			}
			filter = fmt.Sprintf("filter=kind=%s", kind)
		}
		params := fmt.Sprintf("fields=kind,metadata.namespace,metadata.name,metadata.annotation&%s", filter)

		entities := fetchEntitiesByQuery(params)

		var data [][]string
		for _, entity := range entities {
			_, ok := entity.Metadata.Annotations[annotation].(string)
			if !ok {
				row := []string{annotation, getEntityRef(entity), getEntityUrl(entity)}
				data = append(data, row)
			}
		}
		header := []string{"MISSINGANNOTATION", "ENTITYREF", "URL"}
		displayEntities(header, data)

	},
}

// Subcommand for checking relations
var entityNotFoundCmd = &cobra.Command{
	Use:   "entity-not-found [kind|ref] [name]",
	Short: "Relations that don't exist for an entity",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		if len(args) == 0 {
			log.Fatalf("Error: no kind or entityRef provided. Please specify one to check them")
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
					row := []string{
						entityNotFound,
						cleanNamespaceDefault(entityRef),
						getEntityUrlfromRef(addNamespaceDefault(entityRef)),
					}
					data = append(data, row)
				}
			}
		}
		header := []string{"ENTITYNOTFOUND", "USEDIN", "URL"}
		displayEntities(header, data)
	},
}

func init() {
	// Add subcommands to the check command
	checkCmd.AddCommand(orphanCmd)
	checkCmd.AddCommand(missingAnnotationCmd)
	checkCmd.AddCommand(entityNotFoundCmd)
	entityNotFoundCmd.Flags().StringP("filter", "f", "", "Filter output")

	rootCmd.AddCommand(checkCmd)
}
