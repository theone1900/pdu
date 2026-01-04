// Package cmd contains the command handlers for PDU.
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/wublabdubdub/pdu/internal/cmd/bootstrap"
	"github.com/wublabdubdub/pdu/internal/cmd/dropscan"
	"github.com/wublabdubdub/pdu/internal/cmd/info"
	"github.com/wublabdubdub/pdu/internal/cmd/restore"
	"github.com/wublabdubdub/pdu/internal/cmd/scan"
	"github.com/wublabdubdub/pdu/internal/cmd/unload"
)

// AddCommands adds all the commands to the root command.
func AddCommands(rootCmd *cobra.Command) {
	// Add bootstrap command
	bootstrap.AddCommand(rootCmd)

	// Add unload command
	unload.AddCommand(rootCmd)

	// Add scan command
	scan.AddCommand(rootCmd)

	// Add restore command
	restore.AddCommand(rootCmd)

	// Add dropscan command
	dropscan.AddCommand(rootCmd)

	// Add info command
	info.AddCommand(rootCmd)
}
