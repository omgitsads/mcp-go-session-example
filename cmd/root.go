package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mcp",
	Short: "MCP (Model Context Protocol) server implementation",
	Long: `A Model Context Protocol (MCP) server implementation using the Go SDK.
This CLI provides commands to start and manage MCP servers.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
