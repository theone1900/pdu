// Package main is the main entry point for PDU (PostgreSQL Data Unloader).
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wublabdubdub/pdu/internal/cmd"
)

var (
	// Version is the current version of PDU
	Version = "3.0.25.12"
	// BuildTime is the build time of PDU
	BuildTime = "2025-12-11"
)

func main() {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "pdu",
		Short: "PostgreSQL Data Unloader - A tool to read PostgreSQL data files directly",
		Long: `PDU (PostgreSQL Data Unloader) is a comprehensive disaster recovery and data extraction tool for PostgreSQL databases.
It can read PostgreSQL data files directly without requiring a running database instance, making it ideal for disaster recovery,
forensic analysis, and data extraction scenarios.`,
		Version: fmt.Sprintf("%s (built %s)", Version, BuildTime),
	}

	// Initialize configuration
	initConfig()

	// Add commands
	cmd.AddCommands(rootCmd)

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// initConfig initializes the configuration from file and environment variables
func initConfig() {
	// Set default configuration
	viper.SetDefault("PGDATA", ".")
	viper.SetDefault("ARCHIVE_DEST", "./pg_wal")
	viper.SetDefault("DISK_PATH", ".")
	viper.SetDefault("BLOCK_INTERVAL", 20)

	// Read configuration from file
	viper.SetConfigName("pdu")
	viper.SetConfigType("ini")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/pdu/")

	// Read environment variables with PDU_ prefix
	viper.SetEnvPrefix("PDU")
	viper.AutomaticEnv()

	// Try to read configuration file, ignore error if not found
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Warning: Error reading config file: %v\n", err)
		}
	}
}
