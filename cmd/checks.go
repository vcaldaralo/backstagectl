package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check various properties of Backstage entities",
}

// Subcommand for checking owners
var ownerCmd = &cobra.Command{
	Use:   "owner",
	Short: "Check the owner of an entity",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		// Fetch all entities
		entities := fetchEntities("?filter=kind=resource&filter=kind=component&filter=kind=system&filter=kind=domain&fields=kind,metadata.name,metadata.namespace,metadata.annotations,spec.owner")

		// Fetch all users and groups
		owners := fetchEntities("?filter=kind=group&filter=kind=user&fields=kind,metadata.name,metadata.namespace")

		// Create a set of valid owners
		validOwners := make(map[string]bool)
		for _, owner := range owners {
			ownerRef := getEntityRef(owner)
			validOwners[ownerRef] = true
		}

		// Check each entity's owner
		for _, entity := range entities {
			owner, ok := entity.Spec["owner"].(string)
			if !validOwners[owner] && ok {
				entityRef := getEntityRef(entity)
				log.Printf("WARNING: Owner '%s' does not exist for Entity '%s'", owner, entityRef)
			}
		}
	},
}

// Subcommand for checking annotations
var annotationCmd = &cobra.Command{
	Use:   "annotation",
	Short: "Check annotations of an entity",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication
		// Implement annotation check logic here
	},
}

// Subcommand for checking relations
var relationCmd = &cobra.Command{
	Use:   "relation",
	Short: "Check relations of an entity",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication
		// Implement relation check logic here
	},
}

var orphanCmd = &cobra.Command{
	Use:   "orphan",
	Short: "Check orphan entities",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication
		// Implement relation check logic here
	},
}

func init() {
	// Add subcommands to the check command
	checkCmd.AddCommand(ownerCmd)
	checkCmd.AddCommand(annotationCmd)
	checkCmd.AddCommand(relationCmd)
	checkCmd.AddCommand(orphanCmd)

	// Add check command to the root command
	rootCmd.AddCommand(checkCmd)
}
