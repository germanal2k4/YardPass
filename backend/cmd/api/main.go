package main

import (
	"fmt"
	"os"

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

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootParams.ConfigPath, "config", "c", "", "Path to config file")
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute command: %v\n", err)
		os.Exit(1)
	}
}
