package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "backstagectl",
	Short: "A CLI tool to interact with Backstage API",
	Long: `backstagectl is a command line interface tool that allows you to 
interact with Backstage API. You can fetch information 
about entities, APIs, and other entities.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
