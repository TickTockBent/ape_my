# Contributing to Ape_my

Thank you for considering contributing to Ape_my! This document provides guidelines and instructions for contributing.

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help maintain a welcoming environment for all contributors

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue with:

1. **Clear title**: Describe the issue briefly
2. **Steps to reproduce**: List the exact steps to reproduce the bug
3. **Expected behavior**: What should happen
4. **Actual behavior**: What actually happens
5. **Environment**: Go version, OS, etc.
6. **Schema file**: If applicable, include your schema.json

### Suggesting Features

Feature requests are welcome! Please open an issue with:

1. **Use case**: Describe the problem you're trying to solve
2. **Proposed solution**: How you envision the feature working
3. **Alternatives considered**: Other approaches you've thought about
4. **Impact**: How this would benefit other users

### Pull Requests

1. **Fork the repository** and create a branch from `main`
2. **Make your changes** following the coding standards below
3. **Add tests** for any new functionality
4. **Ensure all tests pass**: Run `go test ./...`
5. **Update documentation** as needed
6. **Write a clear commit message** describing your changes
7. **Submit a pull request** with a detailed description

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Git

### Setup

```bash
# Clone the repository
git clone https://github.com/ticktockbent/ape_my.git
cd ape_my

# Install dependencies
go mod tidy

# Build the project
go build -o bin/ape_my ./cmd/ape_my

# Run tests
go test ./...
```

## Coding Standards

### Go Style

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Run `go vet` to catch common issues
- Use descriptive variable names (per project guidelines)

### Code Organization

```
ape_my/
├── cmd/ape_my/          # Main application entry point
├── internal/            # Private application code
│   ├── cli/            # Command-line interface
│   ├── schema/         # Schema parsing and validation
│   ├── storage/        # Data storage implementations
│   └── server/         # HTTP server and routing
├── pkg/types/          # Public types and interfaces
├── examples/           # Example schemas and seed data
├── tests/              # Integration tests
└── docs/               # Documentation
```

### Testing

- Write unit tests for all new functionality
- Aim for meaningful test coverage
- Use table-driven tests where appropriate
- Test error cases and edge cases

Example test structure:

```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid case", "input", "expected", false},
        {"error case", "bad", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FunctionName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Commit Messages

- Use clear, descriptive commit messages
- Start with a verb in present tense (e.g., "Add", "Fix", "Update")
- Keep the first line under 72 characters
- Add details in the body if needed

Good examples:
```
Add support for array field types in schema

Fix panic when entity type is missing

Update README with installation instructions
```

### Documentation

- Document all exported functions and types
- Update README.md for user-facing changes
- Update docs/ for significant features
- Include examples where helpful

## Project Phases

Ape_my is being developed in phases. See [docs/build_plan_v0.1.0.md](docs/build_plan_v0.1.0.md) for the current development roadmap.

When contributing, please consider which phase your contribution fits into and whether it aligns with the current development goals.

## Questions?

If you have questions about contributing, feel free to:

- Open a GitHub issue with the "question" label
- Start a discussion in GitHub Discussions

## License

By contributing to Ape_my, you agree that your contributions will be licensed under the MIT License.

## Thank You!

Your contributions make Ape_my better for everyone. Thank you for taking the time to contribute!
