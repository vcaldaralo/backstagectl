package cmd

import (
	"fmt"
	"log"

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
			viewUrl := getEntityUrl(entity)
			newRow := []string{entity.Metadata.Namespace, entity.Kind, entity.Metadata.Name, viewUrl}
			data = append(data, newRow)
		}
		header := []string{"NAMESPACE", "KIND", "NAME", "URL"}
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
				entityRef := getEntityRef(entity)
				// viewUrl := entity.Metadata.Annotations["backstage.io/view-url"].(string)
				webUrl := getEntityUrl(entity)
				newRow := []string{entity.Metadata.Namespace, entityRef, owner, entityRef, webUrl}
				data = append(data, newRow)
			}
		}
		header := []string{"NAMESPACE", "ENTITY", "OWNERNOTFOUND", "USEDIN"}
		displayEntities(header, data)
	},
}

// Subcommand for checking annotations
var missingAnnotationCmd = &cobra.Command{
	Use:   "missing-annotation [annotation] [kind]",
	Short: "Annotation that is missing for a group of entities",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication
		// Implement annotation check logic here
		var annotation string

		if len(args) == 0 {
			log.Fatalf("error: no annotation key provided. Please specify an annotation key.")
		} else if len(args) > 2 {
			log.Fatalf("error: too many arguments provided. Please specify either one or two arguments.")
		}

		annotation = args[0] // Assign first argument to annotation

		kind := ""
		filter := ""
		if len(args) == 2 {
			kind = args[1] // Assign second argument to kind if provided
		}
		if kind != "" {
			if len(kind) > 0 && kind[len(kind)-1] == 's' {
				kind = kind[:len(kind)-1] // Remove the last character
			}
			filter = fmt.Sprintf("filter=kind=%s", kind)
		}
		params := fmt.Sprintf("fields=kind,metadata.namespace,metadata.name,metadata.annotation&%s", filter)

		entities := fetchEntitiesByQuery(params)

		validOwners := make(map[string]bool)
		for _, entity := range entities {
			entityRef := getEntityRef(entity)
			validOwners[entityRef] = true
		}

		var data [][]string
		for _, entity := range entities {
			_, ok := entity.Metadata.Annotations[annotation].(string)
			if !ok {
				newRow := []string{annotation, getEntityRef(entity), getEntityUrl(entity)}
				data = append(data, newRow)
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
			log.Fatalf("error: no kind or entity ref provided. Please specify one of them")
		} else if len(args) > 2 {
			log.Fatalf("error: too many arguments provided. Please specify either one or two arguments.")
		}

		_, _, _, filter := parseArgs(args)

		params := fmt.Sprintf("fields=kind,metadata.namespace,metadata.name,relations&%s", filter)
		entities := fetchEntitiesByQuery(params) //to verify missing relations

		relationTarget := make(map[string][]string) // Map with TargetRef as key and slice of entityRefs as value
		for _, entity := range entities {
			entityRef := getEntityRef(entity) // Get the entity reference
			for _, rel := range entity.Relations {
				if rel.Type == "dependsOn" || rel.Type == "partOf" { // the relation type
					relationTarget[rel.TargetRef] = append(relationTarget[rel.TargetRef], entityRef) // Append entityRef to the slice
				}
			}
		}
		if len(relationTarget) == 0 {
			log.Fatalf("Relations 'dependsOn' or 'partOf' for this entities don't exist")
		}

		var verifiEntityRef []string
		for key := range relationTarget {
			verifiEntityRef = append(verifiEntityRef, key)
		}

		payload := Payload{
			EntityRefs: verifiEntityRef,
			Fields:     []string{"kind", "metadata.name"},
		}

		entities = fetchEntitiesByRefs(payload)

		var data [][]string
		for i, entity := range entities {
			if entity.Kind == "" {
				// fmt.Printf("Key '%s' in validRelation does not have a corresponding entityRef\n", key)

				//IMPLEMENT GET LIST OF ENTITIES WHERE verifiEntityRef[i] IS USED IN

				newRow := []string{verifiEntityRef[i]}
				printYaml(relationTarget[verifiEntityRef[i]])
				data = append(data, newRow)
			}
		}
		header := []string{"ENTITYNOTFOUND"}
		displayEntities(header, data)
	},
}

func init() {
	// Add subcommands to the check command
	verifyCmd.AddCommand(orphanCmd)
	verifyCmd.AddCommand(missingOwnerCmd)
	verifyCmd.AddCommand(missingAnnotationCmd)
	verifyCmd.AddCommand(entityNotFoundCmd)

	// Add check command to the root command
	rootCmd.AddCommand(verifyCmd)
}
