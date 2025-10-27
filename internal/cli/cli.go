package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	// DefaultPort is the default port for the server
	DefaultPort = 8080

	// Version is the current version
	Version = "0.1.0"
)

var (
	// ErrNoSchemaFile is returned when no schema file is provided
	ErrNoSchemaFile = errors.New("no schema file provided")

	// ErrInvalidPort is returned when the port is invalid
	ErrInvalidPort = errors.New("invalid port number")

	// ErrSchemaNotFound is returned when the schema file doesn't exist
	ErrSchemaNotFound = errors.New("schema file not found")
)

// Config holds the parsed CLI configuration
type Config struct {
	SchemaFile  string
	SeedFile    string
	Port        int
	ShowHelp    bool
	ShowVersion bool
}

// Parse parses command line arguments and returns a Config
func Parse(args []string) (*Config, error) {
	config := &Config{
		Port: DefaultPort,
	}

	// Handle empty args
	if len(args) == 0 {
		return nil, ErrNoSchemaFile
	}

	// Check for flags
	if args[0] == "--help" || args[0] == "-h" {
		config.ShowHelp = true
		return config, nil
	}

	if args[0] == "--version" || args[0] == "-v" {
		config.ShowVersion = true
		return config, nil
	}

	// First argument should be the schema file
	config.SchemaFile = args[0]

	// Parse remaining arguments in natural language style
	i := 1
	for i < len(args) {
		switch args[i] {
		case "with":
			// Next argument should be seed file
			if i+1 >= len(args) {
				return nil, fmt.Errorf("expected seed file after 'with'")
			}
			config.SeedFile = args[i+1]
			i += 2

		case "on":
			// Next argument should be port
			if i+1 >= len(args) {
				return nil, fmt.Errorf("expected port number after 'on'")
			}
			port, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, ErrInvalidPort
			}
			if port < 1 || port > 65535 {
				return nil, fmt.Errorf("%w: must be between 1 and 65535", ErrInvalidPort)
			}
			config.Port = port
			i += 2

		default:
			return nil, fmt.Errorf("unexpected argument: %s", args[i])
		}
	}

	return config, nil
}

// Validate checks if the configuration is valid and files exist
func (c *Config) Validate() error {
	// Skip validation for help/version
	if c.ShowHelp || c.ShowVersion {
		return nil
	}

	// Check if schema file exists
	if _, err := os.Stat(c.SchemaFile); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrSchemaNotFound, c.SchemaFile)
	}

	// Check if seed file exists (if provided)
	if c.SeedFile != "" {
		if _, err := os.Stat(c.SeedFile); os.IsNotExist(err) {
			return fmt.Errorf("seed file not found: %s", c.SeedFile)
		}
	}

	return nil
}

// PrintHelp prints the help message
func PrintHelp() {
	help := `ape_my - A minimalist mock API server

USAGE:
    ape_my <schema.json> [with <seed.json>] [on <port>]
    ape_my --help
    ape_my --version

ARGUMENTS:
    <schema.json>       Path to the JSON schema file (required)

OPTIONS:
    with <seed.json>    Load initial seed data from a JSON file
    on <port>           Specify the port to run on (default: 8080)
    --help, -h          Show this help message
    --version, -v       Show version information

EXAMPLES:
    # Start with an empty API on default port 8080
    ape_my schema.json

    # Start with seed data
    ape_my schema.json with seed.json

    # Start on a custom port
    ape_my schema.json on 3000

    # Combine options
    ape_my schema.json with seed.json on 8080

DOCUMENTATION:
    See README.md for complete documentation
    Schema format: docs/schema_format.md
`
	fmt.Fprint(os.Stderr, help)
}

// PrintVersion prints the version information
func PrintVersion() {
	fmt.Fprintf(os.Stderr, "ape_my version %s\n", Version)
}

// String returns a string representation of the config
func (c *Config) String() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Schema: %s", c.SchemaFile))

	if c.SeedFile != "" {
		parts = append(parts, fmt.Sprintf("Seed: %s", c.SeedFile))
	}

	parts = append(parts, fmt.Sprintf("Port: %d", c.Port))

	return strings.Join(parts, ", ")
}
