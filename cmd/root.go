package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	baseURL     string
	token       string
	tlsCertPath string
	tlsKeyPath  string
)

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
	rootCmd.PersistentFlags().StringVar(&baseURL, "url", "", "Backstage API base URL (required)")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "Authentication token")
	rootCmd.PersistentFlags().StringVar(&tlsCertPath, "tls-cert", "", "Path to Teleport TLS certificate file")
	rootCmd.PersistentFlags().StringVar(&tlsKeyPath, "tls-key", "", "Path to Teleport TLS key file")

	rootCmd.MarkPersistentFlagRequired("url")
}
