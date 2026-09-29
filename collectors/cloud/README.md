# Cloud Collectors

The cloud collectors gather resources from cloud providers and forward them as evidence to Confirmate.

## What Is Included

The cloud collector service currently supports:

- `aws`
- `azure`
- `openstack`
- `k8s`
- `csaf`

## Build

From the repository root:

```bash
go build -o bin/cloud-collector ./collectors/cloud/cmd
```

## Run With Confirmate Framework (Recommended)

Start Confirmate first (includes orchestrator, assessment, and evidence store):

```bash
go build -o bin/confirmate ./core/cmd/confirmate
./bin/confirmate --db-in-memory --api-port 8080
```

Then start the cloud collector in a second terminal:

```bash
./bin/cloud-collector \
  --collector-provider azure \
  --collector-auto-start \
  --target-of-evaluation-id <target-of-evaluation-uuid> \
  --evidence-store-address http://localhost:8080
```

Notes:

- `--collector-provider` is required.
- `--collector-auto-start` starts periodic collection immediately.
- `--target-of-evaluation-id` should be the UUID of the target to associate evidence with.
- `--evidence-store-address` should point to the Confirmate API base URL.

## Alternative: Run Against Another Evidence Store Address

If your evidence service is exposed on another address, set it explicitly:

```bash
./bin/cloud-collector \
  --collector-provider aws \
  --collector-auto-start \
  --target-of-evaluation-id <target-of-evaluation-uuid> \
  --evidence-store-address http://<host>:<port>
```

## Common Provider Example: Azure

```bash
./bin/cloud-collector \
  --collector-provider azure \
  --collector-auto-start \
  --collector-resource-group <resource-group> \
  --target-of-evaluation-id 00000000-0000-0000-0000-000000000000 \
  --evidence-store-address http://localhost:8080
```


## Runtime Flags

```text
--collector-provider string, -p string                Cloud provider (aws, azure, openstack, k8s, csaf)
--collector-tool-id string, -t string                 Collector Tool ID to identify the collector instance
--collector-resource-group string, -r string          Limit the scope of the collector to a specific resource group
--collector-csaf-domain string, -d string             CSAF domain to fetch the CSAF documents from
--target-of-evaluation-id string, -e string           Target of evaluation ID for which to collect cloud evidence
--collector-interval int, -i int                      Interval in minutes for periodic collection
--collector-auto-start, -a                            Start collector automatically after launch
--evidence-store-address string, -s string  Address of the evidence store service
--evidence-store-oauth2-enabled                    Authenticate the connection to the evidence store using OAuth 2.0 client credentials (default: false)
--service-oauth2-token-endpoint string      OAuth 2.0 token URL for service-to-service auth with the evidence store
--service-oauth2-client-id string           OAuth 2.0 client ID for service-to-service auth with the evidence store
--service-oauth2-client-secret string       OAuth 2.0 client secret for service-to-service auth with the evidence store
```

## Authenticating Against The Evidence Store

If the evidence store enforces authentication (`--auth-enabled`), enable OAuth 2.0 client credentials on the collector:

```bash
./bin/cloud-collector \
  --collector-provider azure \
  --collector-auto-start \
  --target-of-evaluation-id <target-of-evaluation-uuid> \
  --evidence-store-address http://localhost:8080 \
  --evidence-store-oauth2-enabled \
  --service-oauth2-client-id <client-id> \
  --service-oauth2-client-secret <client-secret>
```

## Credentials And Access

The collector uses provider SDK authentication and expects credentials to be configured in your environment before startup:

- Azure: Default Azure credential chain (for example `az login`, service principal, or managed identity)
- AWS: Standard AWS SDK credential chain (for example env vars, shared credentials file, or role)
- Kubernetes: kubeconfig / in-cluster configuration
- OpenStack: OpenStack auth environment variables (see `collectors/cloud/service/openstack/README.md`)
- CSAF: network access to the configured provider domain

## Verify It Works

A running collector logs scheduling and collection events. After startup, verify evidence arrives by querying your Confirmate instance (for example with `cf`) for the target of evaluation you used.
