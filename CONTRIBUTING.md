# Contributing to Confirmate

Thank you for your interest in contributing to Confirmate! This document provides guidelines for developers (both human and AI) to ensure consistency and quality across the codebase.

## Table of Contents

- [Code Style Guidelines](#code-style-guidelines)
- [Documentation Guidelines](#documentation-guidelines)
- [Testing Guidelines](#testing-guidelines)
- [Dependencies and Libraries](#dependencies-and-libraries)
- [Pull Request Process](#pull-request-process)

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
func NewService() (svc orchestratorconnect.OrchestratorHandler, err error) {
    var (
        db *persistence.DB
    )
    
    db, err = persistence.NewDB(...)
    if err != nil {
        return nil, err
    }
    
    svc = &service{db: db}
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

### Short Variable Declaration

Avoid using `:=` (short variable declaration) in production code. Instead, use `var` blocks and named return values as shown above.

However, the use of `:=` is acceptable and encouraged in test functions, as tests often need to quickly declare and use variables.

### Import Formatting

- Use `goimports` to automatically format and organize imports
- Imports should be grouped into standard library, internal packages and external packages
- Run `goimports -w .` before committing to ensure consistent import formatting

**Example:**
```go
import (
    // Standard library
    "context"
    "fmt"
    "log/slog"

    // Internal packages
    "github.com/confirmate/confirmate/core/persistence"
    "github.com/confirmate/confirmate/core/util/assert"

    // External packages
    "github.com/google/uuid"
    "github.com/lmittmann/tint"
)
```

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
func NewService() (svc orchestratorconnect.OrchestratorHandler, err error) {
    // Implementation
}

// CreateMetric creates a new metric in the database.
// It returns an error if the metric already exists or if the database operation fails.
func (s *service) CreateMetric(ctx context.Context, req *connect.Request[v1.CreateMetricRequest]) (res *connect.Response[v1.Metric], err error) {
    // Implementation
}
```

**Bad:**
```go
package orchestrator

type service struct {
    orchestratorconnect.UnimplementedOrchestratorHandler
    db *persistence.DB
}

func NewService() (orchestratorconnect.OrchestratorHandler, error) {
    // Implementation
}

func (s *service) CreateMetric(ctx context.Context, req *connect.Request[v1.CreateMetricRequest]) (*connect.Response[v1.Metric], error) {
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

The actual test body should be kept as short and clear as possible. Instead of extensive logic or repetitive code, prefer using `assert.WantErr` or `assert.Want` from the `core/util/assert` package to make checks concise and precise.

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
import (
    "testing"

    "github.com/confirmate/confirmate/core/util/assert"
)

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

## Dependencies and Libraries

### Reduce Dependencies

- Aim to reduce the number of transitive dependencies
- Only add dependencies when absolutely necessary
- Prefer standard library solutions when available

### Configuration and CLI

- Use [github.com/urfave/cli](https://github.com/urfave/cli) v3 for command-line interfaces and configuration
- Do not use cobra or viper
- Follow the cli v3 patterns for command-line interfaces

### Logging

- Use `slog` (standard library structured logging) for all logging
- Do not use `log` or `logrus`
- Use structured logging with appropriate log levels
- Prefer using `slog.Any()` and typed attribute functions (e.g., `slog.String()`, `slog.Int()`) for clarity and to make key-value pairs more explicit
- Use `tint.Err(err)` instead of `slog.Any("error", err)` for error logging

**Example:**
```go
import (
    "log/slog"

    "github.com/lmittmann/tint"
)

slog.Info("service started", slog.Int("port", port))
slog.Error("failed to connect", tint.Err(err))

// For multiple parameters:
slog.Error("failed to connect stream",
    slog.Any("stream", svc.streamName),
    slog.Any("address", svc.address),
    slog.Int("port", svc.port),
    slog.String("evidence", evidence.Id),
    tint.Err(err)
)
```

### Database and Storage

- Use `gorm` for database operations
- Access Gorm directly, wrapped by small utility functions for saving, updating, etc.
- Use the utility functions in `core/persistence` for common operations
- Avoid creating intermediate "storage" abstraction layers

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
   - When merging to the main branch, use a concise, descriptive commit message that follows the Go 
     project style: `<pkg>: <description>` (e.g., `core/service/orchestrator: implement metric export functionality`,
     `core/persistence: fix database timeout issue`, `doc: update API documentation`)

## License

By contributing to Confirmate, you agree that your contributions will be licensed under the Apache License 2.0.
