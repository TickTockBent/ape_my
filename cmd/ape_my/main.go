package main

import (
	"fmt"
	"os"
)

const version = "0.1.0-dev"

func main() {
	fmt.Fprintf(os.Stderr, "ape_my v%s\n", version)
	fmt.Fprintln(os.Stderr, "Mock API server - Coming soon!")
	os.Exit(0)
}
