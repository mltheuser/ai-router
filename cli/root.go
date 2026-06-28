// Package cli wires up the cobra command tree for the ai-router binary.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ai-router",
	Short: "Unified LLM router — one API for all providers",
	Long: `ai-router is a unified LLM gateway that routes requests to multiple
cloud providers (OpenRouter, Google, Anthropic) and local runners (Ollama, vLLM)
through a single, purpose-built API.`,
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
