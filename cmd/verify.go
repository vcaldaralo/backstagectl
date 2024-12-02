package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify properties of Backstage entities",
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

// Subcommand for checking owners
var missingOwnerCmd = &cobra.Command{
	Use:   "missing-owner [kind|ref] [name]",
	Short: "Wheter an entity has a misconfigured owner",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		if len(args) == 0 {
			log.Fatalf("error: no kind or entity ref provided. Please specify one of them")
		} else if len(args) > 2 {
			log.Fatalf("error: too many arguments provided. Please specify either one or two arguments.")
		}

		_, _, _, filter := parseArgs(args)

		params := fmt.Sprintf("fields=kind,metadata.namespace,metadata.name,metadata.annotations,spec.owner&%s", filter)

		// Fetch all entities
		entities := fetchEntitiesByQuery(params)

		// Fetch all users and groups
		owners := fetchEntitiesByQuery("filter=kind=group&filter=kind=user&fields=kind,metadata.namespace,metadata.name")

		// Create a set of valid owners
		validOwners := make(map[string]bool)
		for _, owner := range owners {
			ownerRef := getEntityRef(owner)
			validOwners[ownerRef] = true
		}

		var data [][]string
		for _, entity := range entities {
			owner, ok := entity.Spec["owner"].(string)
			if !validOwners[owner] && ok {
				row := []string{owner, getEntityRef(entity), getEntityUrl(entity)}
				data = append(data, row)
			}
		}
		header := []string{"OWNERNOTFOUND", "ENTITYREF", "URL"}
		displayEntities(header, data)
	},
}

// Subcommand for checking annotations
var missingAnnotationCmd = &cobra.Command{
	Use:   "missing-annotation [annotation] [kind]",
	Short: "Annotation that is missing for a group of entities",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth()

		var annotation string

		if len(args) == 0 {
			log.Fatalf("error: no annotation key provided. Please specify an annotation key.")
		} else if len(args) > 2 {
			log.Fatalf("error: too many arguments provided. Please specify either one or two arguments.")
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
	Short: "Relations that doesn't exist for an entity",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		if len(args) == 0 {
			log.Fatalf("error: no kind or entity ref provided. Please specify one to check them")
		} else if len(args) > 2 {
			log.Fatalf("error: too many arguments provided. Please specify either one or two arguments")
		}

		_, _, _, filter := parseArgs(args)

		params := fmt.Sprintf("fields=kind,metadata.namespace,metadata.name,relations&%s", filter)
		entities := fetchEntitiesByQuery(params)

		relationTarget := make(map[string][]string)
		for _, entity := range entities {
			entityRef := getEntityRef(entity)
			for _, rel := range entity.Relations {
				if rel.Type == "dependsOn" || rel.Type == "partOf" {
					relationTarget[rel.TargetRef] = append(relationTarget[rel.TargetRef], entityRef)
				}
			}
		}
		if len(relationTarget) == 0 {
			log.Fatalf("Relations 'dependsOn' or 'partOf' for these entities don't exist")
		}

		f, _ := cmd.Flags().GetString("filter")
		f = addNamespaceDefault(f)

		var verifiEntityRef []string
		for key := range relationTarget {
			if strings.Contains(key, f) {
				verifiEntityRef = append(verifiEntityRef, key)
			}
		}

		payload := Payload{
			EntityRefs: verifiEntityRef,
			Fields:     []string{"kind", "metadata.name"},
		}

		entities = fetchEntitiesByRefs(payload)

		header := []string{"ENTITYNOTFOUND", "USEDIN"}
		targetNotFound := make(map[string][]string)
		var data [][]string
		for i, entity := range entities {
			if entity.Kind == "" {
				entityNotFound := cleanNamespaceDefault(verifiEntityRef[i])
				usedin := strings.Join(relationTarget[verifiEntityRef[i]], ", ")
				row := []string{entityNotFound, usedin}
				targetNotFound[entityNotFound] = relationTarget[verifiEntityRef[i]]
				if len(targetNotFound) == 1 {
					header = []string{"ENTITYNOTFOUND", "USEDIN", "URL"}
					for _, entity := range relationTarget[verifiEntityRef[i]] {
						row := []string{entityNotFound, entity, getEntityUrlfromRef(addNamespaceDefault(entity))}
						data = append(data, row)
					}
				} else {
					data = append(data, row)
				}
			}
		}
		displayEntities(header, data)
	},
}

func init() {
	// Add subcommands to the check command
	verifyCmd.AddCommand(orphanCmd)
	verifyCmd.AddCommand(missingOwnerCmd)
	verifyCmd.AddCommand(missingAnnotationCmd)
	verifyCmd.AddCommand(entityNotFoundCmd)
	entityNotFoundCmd.Flags().StringP("filter", "f", "", "Filter output")

	rootCmd.AddCommand(verifyCmd)
}
