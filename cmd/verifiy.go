package cmd

import (
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify various properties of Backstage entities",
}

// Subcommand for checking owners
var badOwnerCmd = &cobra.Command{
	Use:   "bad-owner [kind|ref] [name]",
	Short: "Verify wheter an entity has misconfigured owner",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		_, _, filter := parseArgs(args)

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

		// Initialize the array with some values
		output := [][]string{
			{"ENTITY", "OWNER", "URL"},
		}

		for _, entity := range entities {
			owner, ok := entity.Spec["owner"].(string)
			if !validOwners[owner] && ok {
				entityRef := getEntityRef(entity)
				viewUrl := entity.Metadata.Annotations["backstage.io/view-url"].(string)
				newRow := []string{entityRef, owner, viewUrl}
				output = append(output, newRow)
			}
		}

		displayEntities(output)
	},
}

// Subcommand for checking annotations
var missingAnnotationCmd = &cobra.Command{
	Use:   "missing-annotation [annotation] [kind]",
	Short: "Verify an annotation is missing for an entity",
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

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		defer w.Flush()
		fmt.Fprintln(w, "MISSINGANNOTATION\tENTITYREF")
		for _, entity := range entities {
			_, ok := entity.Metadata.Annotations[annotation].(string)
			if !ok {
				entityRef := getEntityRef(entity)
				fmt.Fprintf(w, "%s\t%s\n",
					annotation,
					entityRef,
				)
			}
		}

	},
}

// Subcommand for checking relations
var missingRelationCmd = &cobra.Command{
	Use:   "missing-relation [kind|ref] [name]",
	Short: "Verify relations that doesn't exist for an entity",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		_, _, filter := parseArgs(args)

		params := fmt.Sprintf("fields=kind,metadata.namespace,metadata.name,relations&%s", filter)
		entities := fetchEntitiesByQuery(params)

		relationTarget := make(map[string][]string) // Map with TargetRef as key and slice of entityRefs as value
		for _, entity := range entities {
			entityRef := getEntityRef(entity) // Get the entity reference
			for _, rel := range entity.Relations {
				if rel.Type == "dependsOn" || rel.Type == "partOf" { // Verify the relation type
					relationTarget[rel.TargetRef] = append(relationTarget[rel.TargetRef], entityRef) // Append entityRef to the slice
				}
			}
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

		output := [][]string{
			{"KIND", "NAME", "ENTITIYREF"},
		}

		for i, entity := range entities {
			if entity.Kind == "" {
				kind, _, name := getKindNamespaceName(verifiEntityRef[i])
				// fmt.Printf("Key '%s' in validRelation does not have a corresponding entityRef\n", key)
				newRow := []string{kind, name, verifiEntityRef[i]}
				output = append(output, newRow)
			}
		}
		displayEntities(output)
	},
}

var orphanCmd = &cobra.Command{
	Use:   "orphan",
	Short: "Verify orphan entities",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		entities := fetchEntitiesByQuery("filter=metadata.annotations.backstage.io/orphan=true")

		output := [][]string{
			{"KIND", "NAME", "URL"},
		}

		for _, entity := range entities {
			viewUrl := entity.Metadata.Annotations["backstage.io/view-url"].(string)
			newRow := []string{entity.Kind, entity.Metadata.Name, viewUrl}
			output = append(output, newRow)
		}

		displayEntities(output)
	},
}

func init() {
	// Add subcommands to the check command
	verifyCmd.AddCommand(badOwnerCmd)
	verifyCmd.AddCommand(missingAnnotationCmd)
	verifyCmd.AddCommand(missingRelationCmd)
	verifyCmd.AddCommand(orphanCmd)

	// Add check command to the root command
	rootCmd.AddCommand(verifyCmd)
}
