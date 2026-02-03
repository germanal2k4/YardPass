package main

import (
	"fmt"
	"os"

	"yardpass/internal/config"
	"yardpass/internal/setup"

	"github.com/spf13/cobra"
)

var rootParams struct {
	ConfigPath string
}

var rootCmd = &cobra.Command{
	Use:   "yardpass-api",
	Short: "YardPass API server",
	Long:  "YardPass API server is a REST API for the YardPass parking management system.",
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := setup.SetupApi(rootParams.ConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to setup application: %v\n", err)
			os.Exit(1)
		}
		app.Run()

		return app.Err()
	},
}

var generateConfigCmd = &cobra.Command{
	Use:   "gen-cfg",
	Short: "Generate default config file",
	Long:  "Generate default config file with all fields and their default values.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return config.ApplyDefaults(rootParams.ConfigPath)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootParams.ConfigPath, "config", "c", "", "Path to config file")

	rootCmd.AddCommand(generateConfigCmd)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute command: %v\n", err)
		os.Exit(1)
	}
}
