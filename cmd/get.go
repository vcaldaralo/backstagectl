package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Display one or many Backstage entities",
	Run: func(cmd *cobra.Command, args []string) {
		initAuth() // Initialize authentication
		annotation, _ := cmd.Flags().GetString("annotation")

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
			filter = fmt.Sprintf("?filter=kind=%s", kind)
		}
		if name != "" {
			if filter != "" {
				filter += fmt.Sprintf(",metadata.name=%s", name)
			} else {
				filter = fmt.Sprintf("?filter=metadata.name=%s", name)
			}
		}
		if annotation != "" {
			if filter != "" {
				filter += fmt.Sprintf(",metadata.annotations.%s", annotation)
			} else {
				filter = fmt.Sprintf("?filter=metadata.annotations.%s", annotation)
			}
		}

		entities := fetchEntities(filter)
		displayEntities(entities)
	},
}

func init() {
	getCmd.Flags().StringP("annotation", "a", "", "Filter entities by annotation key")
	rootCmd.AddCommand(getCmd)
}
