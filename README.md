<img src="https://github.com/confirmate/core/assets/12459061/ff0843ea-a144-4a48-8639-30027eea20d5" width="30%"/>

# Confirmate

[![CI](https://github.com/confirmate/confirmate/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/confirmate/confirmate/actions/workflows/build.yml)
[![codecov](https://codecov.io/gh/confirmate/confirmate/branch/main/graph/badge.svg)](https://codecov.io/gh/confirmate/confirmate)

This is work in progress. We are currently in the process of preparing the open-source release for Confirmate Core. In the mean-time you can have a sneak peak at our [UI](http://github.com/confirmate/ui) or look at the [Clouditor](http://github.com/clouditor/clouditor) project, which is the technological basis for Confirmate.

## Build and usage

### Orchestrator (core service)

Build from the repository root:

`go build -o bin/orchestrator ./core/cmd/orchestrator`

Run with an in-memory database (useful for local testing):

`./bin/orchestrator --db-in-memory --api-port 8080`

The orchestrator exposes the API at `http://localhost:8080/` by default. See available flags with:

`./bin/orchestrator --help`

### cf (CLI)

Install from the `core/` directory:

`go install ./cmd/cf/cf.go`

Use the CLI against a running orchestrator (default address is `http://localhost:8080`):

`cf targets list`

You can override the server address with `--addr`:

`cf --addr http://localhost:8080 targets list`

### Confirmate (all-in-one)

This binary is in progress. Build and usage instructions will be added once the PR lands.

### Assessment and Evidence

These binaries are in PR. Build and usage instructions will be added once they land.

