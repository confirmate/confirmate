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

### Service Configuration Pattern

Services should use a `Config` struct with a `DefaultConfig` and a `WithConfig()` option function for configuration, rather than individual option functions for each field. This pattern provides better organization and makes it easier to manage service configuration.

**Good:**
```go
// DefaultConfig is the default configuration for the orchestrator [Service].
var DefaultConfig = Config{
    CreateDefaultTargetOfEvaluation: true,
}

// Config represents the configuration for the orchestrator [Service].
type Config struct {
    // CatalogsFolder is the folder where catalogs are stored.
    CatalogsFolder string
    // MetricsFolder is the folder where metrics are stored.
    MetricsFolder string
    // CreateDefaultTargetOfEvaluation controls whether to create a default target of evaluation.
    CreateDefaultTargetOfEvaluation bool
}

// WithConfig sets the service configuration, overriding the default configuration.
func WithConfig(cfg Config) service.Option[Service] {
    return func(svc *Service) {
        svc.cfg = cfg
    }
}

// In NewService:
func NewService(opts ...service.Option[Service]) (handler Handler, err error) {
    svc = &Service{
        db:  db,
        cfg: DefaultConfig,  // Initialize with defaults
        // ...
    }
    
    // Apply options
    for _, opt := range opts {
        opt(svc)
    }
    // ...
}

// Usage in commands:
svc, err := orchestrator.NewService(
    orchestrator.WithConfig(orchestrator.Config{
        CatalogsFolder:                  cmd.String("catalogs-folder"),
        MetricsFolder:                   cmd.String("metrics-folder"),
        CreateDefaultTargetOfEvaluation: cmd.Bool("create-default-target-of-evaluation"),
    }),
)
```

**Bad:**
```go
// Multiple individual option functions
func WithCatalogsFolder(folder string) service.Option[Service] {
    return func(svc *Service) {
        svc.catalogsFolder = folder
    }
}

func WithMetricsFolder(folder string) service.Option[Service] {
    return func(svc *Service) {
        svc.metricsFolder = folder
    }
}

// Usage requires multiple option calls:
svc, err := orchestrator.NewService(
    orchestrator.WithCatalogsFolder(cmd.String("catalogs-folder")),
    orchestrator.WithMetricsFolder(cmd.String("metrics-folder")),
    // ...
)
```

**Benefits of the Config pattern:**
- All configuration fields are clearly visible in one place
- Easier to set defaults for multiple fields
- Simpler to pass configuration between components
- Reduces the number of exported option functions
- Matches the pattern used in `core/server` for consistency

## Error Handling

### Database Error Handling in Services

When handling database errors in service methods, use the `service.HandleDatabaseError` helper function from `core/service`. This function translates database errors into appropriate Connect RPC errors:

- `persistence.ErrRecordNotFound` → `connect.CodeNotFound`
- Other errors → `connect.CodeInternal`

**Example:**
```go
import (
    "confirmate.io/core/service"
)

func (svc *Service) GetCertificate(
    ctx context.Context,
    req *connect.Request[orchestrator.GetCertificateRequest],
) (res *connect.Response[orchestrator.Certificate], err error) {
    var (
        cert orchestrator.Certificate
    )

    err = svc.db.Get(&cert, "id = ?", req.Msg.CertificateId)
    if err = service.HandleDatabaseError(err, service.ErrNotFound("certificate")); err != nil {
        return nil, err
    }

    res = connect.NewResponse(&cert)
    return
}
```

This pattern ensures consistent error handling across all service methods and reduces code duplication.

### Request Validation

All service methods must validate incoming requests using the `service.Validate` helper function from `core/service`. This function uses `protovalidate` to validate the request message and returns a `connect.CodeInvalidArgument` error if validation fails.

**Validation must be the first operation in every service method, before extracting data from the request:**

```go
import (
    "confirmate.io/core/service"
)

func (svc *Service) CreateCatalog(
    ctx context.Context,
    req *connect.Request[orchestrator.CreateCatalogRequest],
) (res *connect.Response[orchestrator.Catalog], err error) {
    var (
        catalog *orchestrator.Catalog
    )

    // Validate the request FIRST
    if err = service.Validate(req); err != nil {
        return nil, err
    }

    // Extract data from request AFTER validation passes
    catalog = req.Msg.Catalog

    // Continue with business logic...
    err = svc.db.Create(catalog)
    // ...
}
```

**Pattern:** Declare variables in the `var` block, validate the request, then assign values from `req.Msg` only after validation succeeds. This ensures we don't process invalid data.

The validation rules are defined in the protobuf files using `buf/validate` annotations. See [protovalidate documentation](https://buf.build/docs/protovalidate/overview) for details.

### Protocol Buffers Code Generation

After making changes to `.proto` files, regenerate the Go code using:

```bash
go generate ./...
```

This will run all `//go:generate` directives, including multiple `buf generate` commands for:
- Main API definitions
- OpenAPI specifications
- Ontology definitions
- Go struct tags

**Important:** Always run `go generate` from the repository root to ensure all proto files are regenerated correctly.

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

Tests should use the table-driven test pattern as much as possible. This pattern makes tests more maintainable and easier to extend. It specifies the `args` of the call and optionally the `fields` of the struct, if the to-be-called function is a method.

The actual test body should be kept as short and clear as possible. Instead of extensive logic or repetitive code, prefer using `assert.WantErr` or `assert.Want` from the `core/util/assert` package to make checks concise and precise.

Also try to avoid hard-coded strings, especially in table tests. Instead common mock objects (e.g. `orchestratortest.MockMetric1`) should be used.

### Use Mock Constants from `orchestratortest`

**Always use mock constants from the `orchestratortest` package instead of hard-coding IDs or entities in tests.** This ensures consistency across tests and makes them easier to maintain.

The `orchestratortest` package provides a comprehensive set of mock UUIDs, string IDs, and pre-configured entities for all orchestrator types. See [`orchestratortest/mock.go`](core/service/orchestrator/orchestratortest/mock.go) for the complete list of available mocks.

**Good:**
```go
func TestService_GetMetric(t *testing.T) {
    tests := []struct {
        name string
        args args
        want assert.Want[*connect.Response[assessment.Metric]]
        wantErr assert.WantErr
    }{
        {
            name: "happy path",
            args: args{
                req: &orchestrator.GetMetricRequest{
                    MetricId: orchestratortest.MockMetric1.Id,
                },
            },
            // ...
        },
        {
            name: "not found",
            args: args{
                req: &orchestrator.GetMetricRequest{
                    MetricId: orchestratortest.MockNonExistentID,
                },
            },
            // ...
        },
    }
}
```

**Bad:**
```go
func TestService_GetMetric(t *testing.T) {
    tests := []struct {
        name string
        args args
        want assert.Want[*connect.Response[assessment.Metric]]
        wantErr assert.WantErr
    }{
        {
            name: "happy path",
            args: args{
                req: &orchestrator.GetMetricRequest{
                    MetricId: "metric-1",  // Hard-coded ID
                },
            },
            // ...
        },
        {
            name: "not found",
            args: args{
                req: &orchestrator.GetMetricRequest{
                    MetricId: "non-existent",  // Hard-coded ID
                },
            },
            // ...
        },
    }
}
```

If you need a mock entity or ID that doesn't exist, **add it to `orchestratortest/mock.go`** rather than hard-coding it in your test. Follow the existing naming conventions:
- UUID constants: `Mock<EntityType>ID<Number>` (e.g., `MockToeID1`, `MockResultID2`)
- String ID constants: `Mock<EntityType>ID<Number>` (e.g., `MockMetricID1`, `MockCatalogID2`)
- Entity variables: `Mock<EntityType><Number>` (e.g., `MockMetric1`, `MockCatalog2`)
- Special constants: `MockNonExistentID`, `MockEmptyUUID`

**Example:**
```go
func TestService_GetMetric(t *testing.T) {
	type args struct {
		req *orchestrator.GetMetricRequest
	}
	type fields struct {
		db *persistence.DB
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    assert.Want[*connect.Response[assessment.Metric]]
		wantErr assert.WantErr
	}{
		{
			name: "happy path",
			args: args{
				req: &orchestrator.GetMetricRequest{
					MetricId: orchestratortest.MockMetric1.Id,
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables, func(d *persistence.DB) {
					assert.NoError(t, d.Create(orchestratortest.MockMetric1))
				}),
			},
			want: func(t *testing.T, got *connect.Response[assessment.Metric], args ...any) bool {
				assert.NotNil(t, got.Msg)
				return assert.Equal(t, orchestratortest.MockMetric1.Id, got.Msg.Id)
			},
			wantErr: assert.NoError,
		},
		{
			name: "not found",
			args: args{
				req: &orchestrator.GetMetricRequest{
					MetricId: "non-existent",
				},
			},
			fields: fields{
				db: persistencetest.NewInMemoryDB(t, types, joinTables),
			},
			want: assert.Nil[*connect.Response[assessment.Metric]],
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				cErr := assert.Is[*connect.Error](t, err)
				return assert.Equal(t, connect.CodeNotFound, cErr.Code())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db: tt.fields.db,
			}

			res, err := svc.GetMetric(context.Background(), connect.NewRequest(tt.args.req))
			tt.want(t, res)
			tt.wantErr(t, err)
		})
	}
}
```

### Integration Tests

Integration tests are an exception to the table-driven test pattern. They can be written in a more straightforward, sequential style when it makes the test clearer.

### Database Assertions in Tests

For create/update operations that modify the database, include a `wantDB` field in your table tests to verify that the database state matches expectations after the operation. This ensures data integrity and catches issues with persistence logic.

The `wantDB` function receives the database instance and the response from the operation, allowing you to verify both that the data was saved and that it matches the expected state.

**Example:**
```go
func TestService_CreateTargetOfEvaluation(t *testing.T) {
    type args struct {
        req *orchestrator.CreateTargetOfEvaluationRequest
    }
    type fields struct {
        db *persistence.DB
    }
    tests := []struct {
        name    string
        args    args
        fields  fields
        want    assert.Want[*connect.Response[orchestrator.TargetOfEvaluation]]
        wantErr assert.WantErr
        wantDB  assert.Want[*persistence.DB]
    }{
        {
            name: "validation error - empty request",
            args: args{
                req: &orchestrator.CreateTargetOfEvaluationRequest{},
            },
            fields: fields{
                db: persistencetest.NewInMemoryDB(t, types, joinTables),
            },
            want: assert.Nil[*connect.Response[orchestrator.TargetOfEvaluation]],
            wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
                return assert.IsConnectError(t, err, connect.CodeInvalidArgument)
            },
            wantDB: assert.NotNil[*persistence.DB],
        },
        {
            name: "happy path",
            args: args{
                req: &orchestrator.CreateTargetOfEvaluationRequest{
                    TargetOfEvaluation: &orchestrator.TargetOfEvaluation{
                        Name: "test-toe",
                    },
                },
            },
            fields: fields{
                db: persistencetest.NewInMemoryDB(t, types, joinTables),
            },
            want: func(t *testing.T, got *connect.Response[orchestrator.TargetOfEvaluation], args ...any) bool {
                return assert.NotNil(t, got.Msg) &&
                    assert.Equal(t, "test-toe", got.Msg.Name) &&
                    assert.NotEmpty(t, got.Msg.Id)
            },
            wantErr: assert.NoError,
            wantDB: func(t *testing.T, db *persistence.DB, msgAndArgs ...any) bool {
                res := assert.Is[*connect.Response[orchestrator.TargetOfEvaluation]](t, msgAndArgs[0])
                assert.NotNil(t, res)

                toe := assert.InDB[orchestrator.TargetOfEvaluation](t, db, res.Msg.Id)
                assert.Equal(t, "test-toe", toe.Name)
                return true
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc := &Service{
                db: tt.fields.db,
            }
            res, err := svc.CreateTargetOfEvaluation(context.Background(), connect.NewRequest(tt.args.req))
            tt.want(t, res)
            tt.wantErr(t, err)
            tt.wantDB(t, tt.fields.db, res)
        })
    }
}
```

**Key points:**
- For error cases, use `assert.NotNil[*persistence.DB]` to simply verify the DB still exists
- For success cases, use `assert.Is` to type-assert the response, then `assert.InDB` to retrieve and verify the persisted entity
- Always assert `NotNil` on the response before accessing nested fields like `res.Msg.Id`
- Pass the complete response object to `wantDB`, not just the message field

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
