package cmd

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check various properties of Backstage entities",
}

// Subcommand for checking owners
var badOwnerCmd = &cobra.Command{
	Use:   "bad-owner [kind]",
	Short: "Check the owner of an entity",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		var kind string
		filter := ""

		if len(args) > 0 {
			kind = args[0] // Assign first argument to kind

			allowedKinds := map[string]bool{
				"resource":  true,
				"component": true,
				"system":    true,
				"domain":    true,
				"user":      true,
				"group":     true,
				"location":  true,
			} // Define allowed kinds

			if len(kind) > 0 && kind[len(kind)-1] == 's' {
				kind = kind[:len(kind)-1] // Remove the last character
			}
			if !allowedKinds[kind] {
				log.Fatalf("error: backstage doesn't have a resource kind '%s'\nAllowed kinds are: %v", kind, allowedKinds)
			}
		}

		if kind != "" {
			if len(kind) > 0 && kind[len(kind)-1] == 's' {
				kind = kind[:len(kind)-1] // Remove the last character
			}
			filter = fmt.Sprintf("&filter=kind=%s", kind)
		}

		params := fmt.Sprintf("?fields=kind,metadata.name,metadata.namespace,metadata.annotations,spec.owner%s", filter)

		// Fetch all entities
		entities := fetchEntities(params)

		// Fetch all users and groups
		owners := fetchEntities("?filter=kind=group&filter=kind=user&fields=kind,metadata.name,metadata.namespace")

		// Create a set of valid owners
		validOwners := make(map[string]bool)
		for _, owner := range owners {
			ownerRef := getEntityRef(owner)
			validOwners[ownerRef] = true
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		defer w.Flush()
		fmt.Fprintln(w, "MISSINGANNOTATION\tENTITYREF\tURL")
		// Check each entity's owner
		for _, entity := range entities {
			owner, ok := entity.Spec["owner"].(string)
			if !validOwners[owner] && ok {
				entityRef := getEntityRef(entity)
				fmt.Fprintf(w, "%s\t%s\t%s\n",
					owner,
					entityRef,
					entity.Metadata.Annotations["backstage.io/view-url"],
				)
			}
		}
	},
}

// Subcommand for checking annotations
var missingAnnotationCmd = &cobra.Command{
	Use:   "missing-annotation [annotation] [kind]",
	Short: "Check annotations of an entity",
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
			filter = fmt.Sprintf("&filter=kind=%s", kind)
		}

		params := fmt.Sprintf("?fields=kind,metadata.name,metadata.namespace,metadata.annotation%s", filter)

		entities := fetchEntities(params)

		validOwners := make(map[string]bool)
		for _, entity := range entities {
			entityRef := getEntityRef(entity)
			validOwners[entityRef] = true
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		defer w.Flush()
		fmt.Fprintln(w, "MISSINGANNOTATION\tENTITYREF\tURL")
		for _, entity := range entities {
			_, ok := entity.Metadata.Annotations[annotation].(string)
			if !ok {
				entityRef := getEntityRef(entity)
				fmt.Fprintf(w, "%s\t%s\t%s\n",
					annotation,
					entityRef,
					entity.Metadata.Annotations["backstage.io/view-url"],
				)
			}
		}

	},
}

// Subcommand for checking relations
var missingRelationCmd = &cobra.Command{
	Use:   "missing-relation [kind|ref] [name]",
	Short: "Check relations that doesn't exist for an entity",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication

		var kind, name string
		filter := ""

		if len(args) > 0 {
			arg := args[0] // Assign first argument to kind

			pattern := `^[^:]+:[^/]+/[^/]+$`
			matched, _ := regexp.MatchString(pattern, arg)
			if matched {
				ref := strings.Split(arg, ":")
				kind = ref[0]
				name = strings.Split(arg, "/")[1]
			} else {
				kind = arg
			}

			allowedKinds := map[string]bool{
				"resource":  true,
				"component": true,
				"system":    true,
				"domain":    true,
				"user":      true,
				"group":     true,
				"location":  true,
			} // Define allowed kinds

			if len(kind) > 0 && kind[len(kind)-1] == 's' {
				kind = kind[:len(kind)-1] // Remove the last character
			}
			if !allowedKinds[kind] {
				log.Fatalf("error: backstage doesn't have a resource kind '%s'\nAllowed kinds are: %v", kind, allowedKinds)
			}
		}

		if len(args) > 1 && name == "" {
			name = args[1] // Assign second argument to name
		}

		if kind != "" {
			if len(kind) > 0 && kind[len(kind)-1] == 's' {
				kind = kind[:len(kind)-1] // Remove the last character
			}
			filter = fmt.Sprintf("&filter=kind=%s", kind)
		}
		if name != "" {
			if filter != "" {
				filter += fmt.Sprintf(",metadata.name=%s", name)
			} else {
				filter = fmt.Sprintf("&filter=metadata.name=%s", name)
			}
		}

		params := fmt.Sprintf("?fields=kind,metadata.name,metadata.namespace,relations%s", filter)
		entities := fetchEntities(params)

		validRelation := make(map[string][]string) // Map with TargetRef as key and slice of entityRefs as value
		for _, entity := range entities {
			entityRef := getEntityRef(entity) // Get the entity reference
			for _, rel := range entity.Relations {
				if rel.Type == "dependsOn" || rel.Type == "partOf" { // Check the relation type
					validRelation[rel.TargetRef] = append(validRelation[rel.TargetRef], entityRef) // Append entityRef to the slice
				}
			}
		}

		// for targetRef := range validRelation { // Iterate over each key in validRelation
		// 	fmt.Printf("%s\n", targetRef) // Print the key
		// }

		yamlData, err := yaml.Marshal(validRelation)
		if err != nil {
			log.Fatalf("error: %v", err)
		}
		fmt.Println(string(yamlData))

	},
}

var orphanCmd = &cobra.Command{
	Use:   "orphan",
	Short: "Check orphan entities",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication
		entities := fetchEntities("?filter=metadata.annotations.backstage.io/orphan=true")
		displayEntities(entities)
	},
}

func init() {
	// Add subcommands to the check command
	checkCmd.AddCommand(badOwnerCmd)
	checkCmd.AddCommand(missingAnnotationCmd)
	checkCmd.AddCommand(missingRelationCmd)
	checkCmd.AddCommand(orphanCmd)

	// Add check command to the root command
	rootCmd.AddCommand(checkCmd)
}
