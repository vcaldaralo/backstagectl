package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Entity struct {
	ApiVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Metadata   struct {
		Name        string                 `json:"name"`
		Namespace   string                 `json:"namespace"`
		Description string                 `json:"description"`
		Annotations map[string]interface{} `json:"annotations"`
	} `json:"metadata"`
	Spec map[string]interface{} `json:"spec"`
}

// type EntitiesResponse struct {
// 	Items []Entity `json:"items"`
// }

type Entities []Entity

var rootCmd = &cobra.Command{
	Use:   "backstage-cli",
	Short: "A CLI tool to interact with Backstage API",
	Long: `backstage-cli is a command line interface tool that allows you to 
interact with Backstage API in read-only mode. You can fetch information 
about entities, APIs, and other entities.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
}
