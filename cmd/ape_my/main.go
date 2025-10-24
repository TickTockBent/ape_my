package main

import (
	"fmt"
	"os"

	"github.com/ticktockbent/ape_my/internal/cli"
)

func main() {
	// Parse command line arguments
	config, err := cli.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		cli.PrintHelp()
		os.Exit(1)
	}

	// Handle help flag
	if config.ShowHelp {
		cli.PrintHelp()
		os.Exit(0)
	}

	// Handle version flag
	if config.ShowVersion {
		cli.PrintVersion()
		os.Exit(0)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Print configuration
	fmt.Fprintf(os.Stderr, "ape_my v%s\n", cli.Version)
	fmt.Fprintf(os.Stderr, "Configuration: %s\n", config.String())
	fmt.Fprintln(os.Stderr, "\nServer starting... (functionality coming in Phase 2-5)")

	// TODO: Phase 2 - Load and parse schema
	// TODO: Phase 3 - Initialize storage
	// TODO: Phase 4 - Start HTTP server
	// TODO: Phase 5 - Handle requests

	os.Exit(0)
}
