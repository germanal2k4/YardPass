package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

const (
	defaultDSN = "postgres://yardpass:password@localhost:5432/yardpass?sslmode=disable"
	defaultDir = "migrations"
)

var rootParams struct {
	DSN string
	Dir string
}

var rootCmd = &cobra.Command{
	Use:   "migrate [command] [args]",
	Short: "YardPass database migration tool",
	Long: `YardPass Database Migration Tool

Applies and manages database migrations using goose.

Commands passed as arguments:
  up                   Apply all pending migrations
  up-by-one            Apply the next pending migration
  up-to VERSION        Migrate up to a specific version
  down                 Roll back the last migration
  down-to VERSION      Roll back to a specific version
  redo                 Roll back and re-apply the last migration
  status               Show the status of all migrations
  version              Print the current migration version
  create NAME sql      Create a new SQL migration file`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		_ = godotenv.Load()

		dsn := rootParams.DSN
		if dsn == "" {
			dsn = envOrDefault("DATABASE_URL", defaultDSN)
		}

		dir := rootParams.Dir
		if dir == "" {
			dir = envOrDefault("MIGRATIONS_DIR", defaultDir)
		}

		db, err := sql.Open("pgx", dsn)
		if err != nil {
			return fmt.Errorf("open database: %w", err)
		}
		defer func() {
			if errClose := db.Close(); errClose != nil {
				err = errors.Join(err, fmt.Errorf("close database: %w", errClose))
			}
		}()

		if err := db.Ping(); err != nil {
			return fmt.Errorf("connect to database: %w", err)
		}

		goose.SetTableName("goose_db_version")

		command := args[0]
		gooseArgs := args[1:]

		if err := goose.RunContext(context.Background(), command, db, dir, gooseArgs...); err != nil {
			return fmt.Errorf("goose %s: %w", command, err)
		}

		return nil
	},
}

func init() {
	rootCmd.Flags().StringVar(&rootParams.DSN, "dsn", "", "PostgreSQL connection string (env: DATABASE_URL)")
	rootCmd.Flags().StringVar(&rootParams.Dir, "dir", "", "Path to migrations directory (env: MIGRATIONS_DIR)")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute command: %v\n", err)
		os.Exit(1)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
