# Contributing to Confirmate

Thank you for your interest in contributing to Confirmate! This document provides guidelines for developers (both human and AI) to ensure consistency and quality across the codebase.

## Table of Contents

- [Development Setup](#development-setup)
- [Code Style Guidelines](#code-style-guidelines)
- [Documentation Guidelines](#documentation-guidelines)
- [Testing Guidelines](#testing-guidelines)
- [Migration Guidelines](#migration-guidelines)
- [Pull Request Process](#pull-request-process)

## Development Setup

### Prerequisites

- Go 1.21 or later
- [buf](https://buf.build) for Protocol Buffer generation
- Git for version control

### Building the Project

```bash
cd core
go install github.com/srikrsna/protoc-gen-gotag
go generate ./...
mkdir -p build
go build -o build ./cmd/orchestrator
```

### Running Tests

```bash
cd core
go test ./...
```

## Code Style Guidelines

### Named Return Values

Always use named return values in function signatures, especially for error returns. This improves code readability and makes the function's intent clearer.

**Good:**
```go
func ProcessData(input string) (result *Data, err error) {
    var (
        processed string
        valid     bool
    )
    
    // Function implementation
    return result, nil
}
```

**Bad:**
```go
func ProcessData(input string) (*Data, error) {
    // Function implementation
}
```

### Variable Declaration

Use `var` blocks at the beginning of functions to group all variables needed in the function. This makes it easier to understand what variables are used throughout the function.

**Good:**
```go
func NewService() (orchestratorconnect.OrchestratorHandler, error) {
    var (
        svc = &service{}
        err error
    )
    
    // Function implementation
    return svc, nil
}
```

**Bad:**
```go
func NewService() (orchestratorconnect.OrchestratorHandler, error) {
    svc := &service{}
    err := initialize()
    // Variables scattered throughout
}
```

### Short Variable Declaration in Tests

The use of `:=` (short variable declaration) is acceptable and encouraged in test functions, as tests often need to quickly declare and use variables. However, production code should follow the named return and `var` block guidelines.

## Documentation Guidelines

### Use godoc

Prefer using godoc comments for documentation instead of creating separate `README.md` files within code directories. All exported functions, types, methods, and packages should have comprehensive godoc comments.

**Good:**
```go
// Package orchestrator implements the orchestration service for Confirmate.
// It provides functionality for managing targets of evaluation and their
// associated metrics.
package orchestrator

// service implements the Orchestrator service handler (see
// [orchestratorconnect.OrchestratorHandler]).
type service struct {
    orchestratorconnect.UnimplementedOrchestratorHandler
    db *persistence.DB
}

// NewService creates a new orchestrator service and returns a
// [orchestratorconnect.OrchestratorHandler].
//
// It initializes the database with auto-migration for the required types and sets up the necessary
// join tables.
func NewService() (orchestratorconnect.OrchestratorHandler, error) {
    // Implementation
}
```

### README.md Files

`README.md` files should only be added at the core of packages such as:
- Repository root
- Major package directories like `core`, `ui`, `collectors`
- Standalone components that need user-facing documentation

Avoid adding `README.md` files in internal code directories. Instead, use package-level godoc comments.

### Documentation Quality

- Document the purpose and behavior of exported functions and types
- Include examples where appropriate
- Document any side effects or important considerations
- Use proper Go documentation formatting (see [Go documentation comments](https://go.dev/doc/comment))

## Testing Guidelines

### Table-Driven Tests

Tests should use the table-driven test pattern as much as possible. This pattern makes tests more maintainable and easier to extend.

**Example:**
```go
func TestEqual(t *testing.T) {
    type args struct {
        t    TestingT
        want any
        got  any
        opts []cmp.Option
    }
    tests := []struct {
        name string
        args args
        want bool
    }{
        {
            name: "compare literals",
            args: args{
                t:    t,
                want: "5",
                got:  "5",
            },
            want: true,
        },
        {
            name: "compare structs with unexported fields",
            args: args{
                t:    t,
                want: &MyStruct{A: "test", b: 1},
                got:  &MyStruct{A: "test", b: 1},
                opts: []cmp.Option{CompareAllUnexported()},
            },
            want: true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Equal(tt.args.t, tt.args.want, tt.args.got, tt.args.opts...); got != tt.want {
                t.Errorf("Equal() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

Integration tests are an exception to the table-driven test pattern. They can be written in a more straightforward, sequential style when it makes the test clearer.

### Use core/util/assert

For test assertions, use the `core/util/assert` package instead of direct comparison or external assertion libraries. This helps maintain consistency and allows us to track and optimize assertion usage.

**Example:**
```go
func Test_DB_Create(t *testing.T) {
    var (
        err    error
        s      *DB
        metric *assessment.Metric
    )

    metric = &assessment.Metric{
        Id:          MockMetricID1,
        Category:    MockMetricCategory1,
        Description: MockMetricDescription1,
    }

    s, err = NewDB(
        WithInMemory(),
        WithAutoMigration(&assessment.Metric{}),
    )
    assert.NoError(t, err)

    err = s.Create(metric)
    assert.NoError(t, err)

    err = s.Create(metric)
    assert.Error(t, err)
}
```

### Short Variable Declaration in Tests

In test functions, it's acceptable to use `:=` for variable declarations:

```go
func TestSomething(t *testing.T) {
    // This is fine in tests
    got := doSomething()
    want := "expected"
    assert.Equal(t, want, got)
}
```

## Migration Guidelines

When migrating components from Clouditor to Confirmate, follow these guidelines:

### Reduce Dependencies

- Aim to reduce the number of transitive dependencies
- Only add dependencies when absolutely necessary
- Prefer standard library solutions when available

### Configuration and CLI

- Use [github.com/urfave/cli](https://github.com/urfave/cli) v3 instead of cobra/viper for configuration
- No migrated code should refer to cobra or viper
- Follow the cli v3 patterns for command-line interfaces

### Logging

- Use `slog` (standard library structured logging) instead of `log`
- Remove all references to `logrus`
- Use structured logging with appropriate log levels

**Example:**
```go
import "log/slog"

slog.Info("service started", "port", port)
slog.Error("failed to connect", "error", err)
```

### Database and Storage

- Keep using `gorm` for database operations
- Remove the intermediate "storage" layer
- Allow (more or less) direct access to Gorm, wrapped by small utility functions for saving, updating, etc.
- Use the utility functions in `core/persistence` for common operations

### Testing Assertions

- Use `core/util/assert` for test assertions
- This package provides a consistent interface and may change implementations in the future
- Always use this package instead of importing external assertion libraries directly

## Pull Request Process

1. **Before Making Changes:**
   - Ensure you understand the issue or feature request
   - Check existing code patterns and follow them
   - Review these contributing guidelines

2. **Making Changes:**
   - Keep changes minimal and focused
   - Follow the code style guidelines
   - Add or update tests as appropriate
   - Update documentation (godoc comments) for changed code

3. **Before Submitting:**
   - Run tests: `go test ./...`
   - Run linters if available
   - Ensure your code builds successfully
   - Write clear, descriptive commit messages

4. **Pull Request:**
   - Provide a clear description of the changes
   - Reference any related issues
   - Respond to review feedback promptly

## License

By contributing to Confirmate, you agree that your contributions will be licensed under the Apache License 2.0.
