// Package cli wires up the cobra command tree for the ai-router binary.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ai-router",
	Short: "Unified AI router — one API per use case, across all providers",
	Long: `ai-router routes requests for each AI use case (chat, embedding, ...) to
cloud providers and local runners through one purpose-built API per use case.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(serveCmd)
}
